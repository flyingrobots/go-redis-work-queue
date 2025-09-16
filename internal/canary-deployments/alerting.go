package canary_deployments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// WebhookAlerter implements the Alerter interface using webhooks
type WebhookAlerter struct {
	webhookURLs  []string
	httpClient   *http.Client
	logger       *slog.Logger
	redis        *redis.Client
	cooldownMap  map[string]time.Time
	mu           sync.RWMutex
	cooldown     time.Duration
}

// NewWebhookAlerter creates a new webhook-based alerter
func NewWebhookAlerter(webhookURLs []string, logger *slog.Logger) *WebhookAlerter {
	return &WebhookAlerter{
		webhookURLs: webhookURLs,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger:      logger,
		cooldownMap: make(map[string]time.Time),
		cooldown:    5 * time.Minute, // 5-minute cooldown for similar alerts
	}
}

// SendAlert sends an alert to configured webhooks
func (wa *WebhookAlerter) SendAlert(ctx context.Context, alert *Alert) error {
	// Check cooldown to prevent alert spam
	if wa.isInCooldown(alert) {
		wa.logger.Debug("Alert suppressed due to cooldown",
			"alert_id", alert.ID,
			"type", alert.Level)
		return nil
	}

	// Store alert
	if wa.redis != nil {
		if err := wa.storeAlert(ctx, alert); err != nil {
			wa.logger.Warn("Failed to store alert", "error", err)
		}
	}

	// Send to webhooks
	for _, webhookURL := range wa.webhookURLs {
		if err := wa.sendWebhook(ctx, webhookURL, alert); err != nil {
			wa.logger.Error("Failed to send webhook",
				"url", webhookURL,
				"alert_id", alert.ID,
				"error", err)
		}
	}

	// Update cooldown
	wa.setCooldown(alert)

	wa.logger.Info("Alert sent",
		"alert_id", alert.ID,
		"level", alert.Level,
		"deployment_id", alert.DeploymentID)

	return nil
}

// GetAlerts retrieves alerts for a deployment
func (wa *WebhookAlerter) GetAlerts(ctx context.Context, deploymentID string) ([]*Alert, error) {
	if wa.redis == nil {
		return []*Alert{}, nil
	}

	key := fmt.Sprintf("canary:alerts:%s", deploymentID)
	results, err := wa.redis.ZRevRange(ctx, key, 0, 50).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}

	alerts := make([]*Alert, 0, len(results))
	for _, result := range results {
		var alert Alert
		if err := json.Unmarshal([]byte(result), &alert); err != nil {
			wa.logger.Warn("Failed to unmarshal alert", "error", err)
			continue
		}
		alerts = append(alerts, &alert)
	}

	return alerts, nil
}

// ResolveAlert marks an alert as resolved
func (wa *WebhookAlerter) ResolveAlert(ctx context.Context, alertID string) error {
	if wa.redis == nil {
		return NewCanaryError(CodeSystemNotReady, "Redis not available")
	}

	// Find the alert
	alert, err := wa.findAlert(ctx, alertID)
	if err != nil {
		return err
	}

	// Mark as resolved
	alert.Resolved = true
	now := time.Now()
	alert.ResolvedAt = &now

	// Update in Redis
	if err := wa.updateAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	wa.logger.Info("Alert resolved", "alert_id", alertID)
	return nil
}

// Private methods

func (wa *WebhookAlerter) isInCooldown(alert *Alert) bool {
	wa.mu.RLock()
	defer wa.mu.RUnlock()

	key := wa.getCooldownKey(alert)
	lastSent, exists := wa.cooldownMap[key]
	if !exists {
		return false
	}

	return time.Since(lastSent) < wa.cooldown
}

func (wa *WebhookAlerter) setCooldown(alert *Alert) {
	wa.mu.Lock()
	defer wa.mu.Unlock()

	key := wa.getCooldownKey(alert)
	wa.cooldownMap[key] = time.Now()
}

func (wa *WebhookAlerter) getCooldownKey(alert *Alert) string {
	return fmt.Sprintf("%s:%s:%s", alert.DeploymentID, alert.Level, alert.Action)
}

func (wa *WebhookAlerter) sendWebhook(ctx context.Context, url string, alert *Alert) error {
	payload := WebhookPayload{
		Alert:     alert,
		Timestamp: time.Now(),
		Source:    "canary-deployment-system",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "canary-deployment-alerter/1.0")
	req.Header.Set("X-Alert-ID", alert.ID)

	resp, err := wa.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned error status: %d", resp.StatusCode)
	}

	return nil
}

func (wa *WebhookAlerter) storeAlert(ctx context.Context, alert *Alert) error {
	key := fmt.Sprintf("canary:alerts:%s", alert.DeploymentID)
	score := float64(alert.Timestamp.Unix())

	data, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	if err := wa.redis.ZAdd(ctx, key, &redis.Z{
		Score:  score,
		Member: data,
	}).Err(); err != nil {
		return fmt.Errorf("failed to store alert: %w", err)
	}

	// Also store in global alerts index
	globalKey := "canary:alerts:all"
	if err := wa.redis.ZAdd(ctx, globalKey, &redis.Z{
		Score:  score,
		Member: data,
	}).Err(); err != nil {
		wa.logger.Warn("Failed to store alert in global index", "error", err)
	}

	// Set expiration to prevent unlimited growth
	wa.redis.Expire(ctx, key, 30*24*time.Hour) // 30 days
	wa.redis.Expire(ctx, globalKey, 30*24*time.Hour)

	return nil
}

func (wa *WebhookAlerter) findAlert(ctx context.Context, alertID string) (*Alert, error) {
	// Search in global alerts index
	key := "canary:alerts:all"
	results, err := wa.redis.ZRevRange(ctx, key, 0, 1000).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to search alerts: %w", err)
	}

	for _, result := range results {
		var alert Alert
		if err := json.Unmarshal([]byte(result), &alert); err != nil {
			continue
		}
		if alert.ID == alertID {
			return &alert, nil
		}
	}

	return nil, NewCanaryError(CodeDeploymentNotFound, "alert not found")
}

func (wa *WebhookAlerter) updateAlert(ctx context.Context, alert *Alert) error {
	// Remove old version and add updated version
	deploymentKey := fmt.Sprintf("canary:alerts:%s", alert.DeploymentID)
	globalKey := "canary:alerts:all"

	data, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	score := float64(alert.Timestamp.Unix())

	// Update in deployment-specific alerts
	if err := wa.redis.ZAdd(ctx, deploymentKey, &redis.Z{
		Score:  score,
		Member: data,
	}).Err(); err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	// Update in global alerts
	if err := wa.redis.ZAdd(ctx, globalKey, &redis.Z{
		Score:  score,
		Member: data,
	}).Err(); err != nil {
		wa.logger.Warn("Failed to update alert in global index", "error", err)
	}

	return nil
}

// WebhookPayload represents the payload sent to webhooks
type WebhookPayload struct {
	Alert     *Alert    `json:"alert"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

// SlackAlerter implements Slack-specific alerting
type SlackAlerter struct {
	webhookURL string
	channel    string
	httpClient *http.Client
	logger     *slog.Logger
}

// NewSlackAlerter creates a new Slack alerter
func NewSlackAlerter(webhookURL, channel string, logger *slog.Logger) *SlackAlerter {
	return &SlackAlerter{
		webhookURL: webhookURL,
		channel:    channel,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger,
	}
}

// SendAlert sends an alert to Slack
func (sa *SlackAlerter) SendAlert(ctx context.Context, alert *Alert) error {
	message := sa.formatSlackMessage(alert)

	payload := SlackPayload{
		Channel:   sa.channel,
		Username:  "Canary Bot",
		IconEmoji: ":canary:",
		Text:      message.Text,
		Attachments: []SlackAttachment{
			{
				Color:  message.Color,
				Title:  message.Title,
				Text:   message.Details,
				Fields: message.Fields,
				Footer: "Canary Deployment System",
				Ts:     alert.Timestamp.Unix(),
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", sa.webhookURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := sa.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Slack returned error status: %d", resp.StatusCode)
	}

	sa.logger.Info("Slack alert sent",
		"alert_id", alert.ID,
		"channel", sa.channel)

	return nil
}

func (sa *SlackAlerter) formatSlackMessage(alert *Alert) SlackMessage {
	var color string
	var emoji string

	switch alert.Level {
	case InfoAlert:
		color = "good"
		emoji = ":information_source:"
	case WarningAlert:
		color = "warning"
		emoji = ":warning:"
	case CriticalAlert:
		color = "danger"
		emoji = ":rotating_light:"
	default:
		color = "good"
		emoji = ":speech_balloon:"
	}

	title := fmt.Sprintf("%s Canary Alert", emoji)
	text := fmt.Sprintf("*%s*: %s", alert.Level, alert.Message)

	fields := []SlackField{
		{
			Title: "Deployment ID",
			Value: alert.DeploymentID,
			Short: true,
		},
		{
			Title: "Alert Level",
			Value: string(alert.Level),
			Short: true,
		},
		{
			Title: "Action",
			Value: string(alert.Action),
			Short: true,
		},
		{
			Title: "Timestamp",
			Value: alert.Timestamp.Format(time.RFC3339),
			Short: true,
		},
	}

	// Add details if available
	if alert.Details != nil {
		if details, ok := alert.Details.(map[string]interface{}); ok {
			for key, value := range details {
				fields = append(fields, SlackField{
					Title: key,
					Value: fmt.Sprintf("%v", value),
					Short: true,
				})
			}
		}
	}

	return SlackMessage{
		Text:    text,
		Color:   color,
		Title:   title,
		Details: alert.Message,
		Fields:  fields,
	}
}

// Slack message types
type SlackMessage struct {
	Text    string       `json:"text"`
	Color   string       `json:"color"`
	Title   string       `json:"title"`
	Details string       `json:"details"`
	Fields  []SlackField `json:"fields"`
}

type SlackPayload struct {
	Channel     string            `json:"channel"`
	Username    string            `json:"username"`
	IconEmoji   string            `json:"icon_emoji"`
	Text        string            `json:"text"`
	Attachments []SlackAttachment `json:"attachments"`
}

type SlackAttachment struct {
	Color  string       `json:"color"`
	Title  string       `json:"title"`
	Text   string       `json:"text"`
	Fields []SlackField `json:"fields"`
	Footer string       `json:"footer"`
	Ts     int64        `json:"ts"`
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// PagerDutyAlerter implements PagerDuty integration
type PagerDutyAlerter struct {
	integrationKey string
	httpClient     *http.Client
	logger         *slog.Logger
}

// NewPagerDutyAlerter creates a new PagerDuty alerter
func NewPagerDutyAlerter(integrationKey string, logger *slog.Logger) *PagerDutyAlerter {
	return &PagerDutyAlerter{
		integrationKey: integrationKey,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
		logger:         logger,
	}
}

// SendAlert sends an alert to PagerDuty
func (pda *PagerDutyAlerter) SendAlert(ctx context.Context, alert *Alert) error {
	// Only send critical alerts to PagerDuty
	if alert.Level != CriticalAlert {
		return nil
	}

	severity := "error"
	if alert.Action == ForceRollback {
		severity = "critical"
	}

	event := PagerDutyEvent{
		RoutingKey:  pda.integrationKey,
		EventAction: "trigger",
		DedupKey:    fmt.Sprintf("canary-%s", alert.DeploymentID),
		Payload: PagerDutyPayload{
			Summary:   alert.Message,
			Source:    "canary-deployment-system",
			Severity:  severity,
			Component: "canary-deployment",
			Group:     "deployment",
			Class:     "deployment-failure",
			CustomDetails: map[string]interface{}{
				"deployment_id": alert.DeploymentID,
				"alert_level":   alert.Level,
				"action":        alert.Action,
				"details":       alert.Details,
			},
		},
		Client:    "Canary Deployment System",
		ClientURL: fmt.Sprintf("https://console.example.com/deployments/%s", alert.DeploymentID),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal PagerDuty event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://events.pagerduty.com/v2/enqueue", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := pda.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send PagerDuty event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("PagerDuty returned error status: %d", resp.StatusCode)
	}

	pda.logger.Info("PagerDuty alert sent",
		"alert_id", alert.ID,
		"dedup_key", event.DedupKey)

	return nil
}

// PagerDuty event types
type PagerDutyEvent struct {
	RoutingKey   string            `json:"routing_key"`
	EventAction  string            `json:"event_action"`
	DedupKey     string            `json:"dedup_key"`
	Payload      PagerDutyPayload  `json:"payload"`
	Client       string            `json:"client"`
	ClientURL    string            `json:"client_url"`
}

type PagerDutyPayload struct {
	Summary       string                 `json:"summary"`
	Source        string                 `json:"source"`
	Severity      string                 `json:"severity"`
	Component     string                 `json:"component"`
	Group         string                 `json:"group"`
	Class         string                 `json:"class"`
	CustomDetails map[string]interface{} `json:"custom_details"`
}

// CompositeAlerter combines multiple alerters
type CompositeAlerter struct {
	alerters []Alerter
	redis    *redis.Client
	logger   *slog.Logger
}

// NewCompositeAlerter creates a new composite alerter
func NewCompositeAlerter(redis *redis.Client, logger *slog.Logger) *CompositeAlerter {
	return &CompositeAlerter{
		alerters: make([]Alerter, 0),
		redis:    redis,
		logger:   logger,
	}
}

// AddAlerter adds an alerter to the composite
func (ca *CompositeAlerter) AddAlerter(alerter Alerter) {
	ca.alerters = append(ca.alerters, alerter)
}

// SendAlert sends an alert to all configured alerters
func (ca *CompositeAlerter) SendAlert(ctx context.Context, alert *Alert) error {
	var lastErr error

	for _, alerter := range ca.alerters {
		if err := alerter.SendAlert(ctx, alert); err != nil {
			ca.logger.Error("Alerter failed to send alert",
				"alerter_type", fmt.Sprintf("%T", alerter),
				"alert_id", alert.ID,
				"error", err)
			lastErr = err
		}
	}

	return lastErr
}

// GetAlerts retrieves alerts (delegates to the first alerter that supports it)
func (ca *CompositeAlerter) GetAlerts(ctx context.Context, deploymentID string) ([]*Alert, error) {
	for _, alerter := range ca.alerters {
		if alerts, err := alerter.GetAlerts(ctx, deploymentID); err == nil {
			return alerts, nil
		}
	}

	return []*Alert{}, nil
}

// ResolveAlert resolves an alert (delegates to the first alerter that supports it)
func (ca *CompositeAlerter) ResolveAlert(ctx context.Context, alertID string) error {
	for _, alerter := range ca.alerters {
		if err := alerter.ResolveAlert(ctx, alertID); err == nil {
			return nil
		}
	}

	return NewCanaryError(CodeAlertFailed, "no alerter could resolve the alert")
}

// AlertManager provides high-level alert management
type AlertManager struct {
	alerter Alerter
	logger  *slog.Logger
}

// NewAlertManager creates a new alert manager
func NewAlertManager(alerter Alerter, logger *slog.Logger) *AlertManager {
	return &AlertManager{
		alerter: alerter,
		logger:  logger,
	}
}

// SendDeploymentAlert sends an alert for a deployment event
func (am *AlertManager) SendDeploymentAlert(ctx context.Context, deployment *CanaryDeployment, level AlertLevel, message string, action AlertAction) error {
	alert := &Alert{
		ID:           "alert_" + uuid.New().String(),
		DeploymentID: deployment.ID,
		Level:        level,
		Message:      message,
		Action:       action,
		Timestamp:    time.Now(),
		Details: map[string]interface{}{
			"queue":           deployment.QueueName,
			"stable_version":  deployment.StableVersion,
			"canary_version":  deployment.CanaryVersion,
			"current_percent": deployment.CurrentPercent,
			"status":          deployment.Status,
		},
	}

	return am.alerter.SendAlert(ctx, alert)
}

// SendHealthAlert sends an alert for a health issue
func (am *AlertManager) SendHealthAlert(ctx context.Context, deployment *CanaryDeployment, health *CanaryHealthStatus) error {
	level := InfoAlert
	action := NoAction

	if health.OverallStatus == FailingCanary {
		level = CriticalAlert
		action = SuggestRollback
	} else if health.OverallStatus == WarningCanary {
		level = WarningAlert
		action = SuggestPause
	}

	message := fmt.Sprintf("Canary health status: %s - %s", health.OverallStatus, health.GetFailureReason())

	alert := &Alert{
		ID:           "alert_" + uuid.New().String(),
		DeploymentID: deployment.ID,
		Level:        level,
		Message:      message,
		Action:       action,
		Timestamp:    time.Now(),
		Details: map[string]interface{}{
			"health_status":     health.OverallStatus,
			"error_rate_check":  health.ErrorRateCheck.Passing,
			"latency_check":     health.LatencyCheck.Passing,
			"throughput_check":  health.ThroughputCheck.Passing,
			"sample_size_check": health.SampleSizeCheck.Passing,
		},
	}

	return am.alerter.SendAlert(ctx, alert)
}

// SendMetricsAlert sends an alert for metrics anomalies
func (am *AlertManager) SendMetricsAlert(ctx context.Context, deployment *CanaryDeployment, anomalies []*PerformanceAnomaly) error {
	if len(anomalies) == 0 {
		return nil
	}

	level := WarningAlert
	action := SuggestPause

	// Determine severity based on worst anomaly
	for _, anomaly := range anomalies {
		if anomaly.Severity == "critical" {
			level = CriticalAlert
			action = SuggestRollback
			break
		}
	}

	message := fmt.Sprintf("Detected %d performance anomalies in canary deployment", len(anomalies))

	details := make(map[string]interface{})
	details["anomaly_count"] = len(anomalies)
	for i, anomaly := range anomalies {
		details[fmt.Sprintf("anomaly_%d", i)] = map[string]interface{}{
			"type":        anomaly.Type,
			"severity":    anomaly.Severity,
			"description": anomaly.Description,
			"value":       anomaly.Value,
		}
	}

	alert := &Alert{
		ID:           "alert_" + uuid.New().String(),
		DeploymentID: deployment.ID,
		Level:        level,
		Message:      message,
		Action:       action,
		Timestamp:    time.Now(),
		Details:      details,
	}

	return am.alerter.SendAlert(ctx, alert)
}