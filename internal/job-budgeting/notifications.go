package budgeting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// NotificationService handles sending budget alerts through various channels
type NotificationService struct {
	httpClient *http.Client
	retries    int
	timeout    time.Duration
}

// NewNotificationService creates a new notification service
func NewNotificationService() *NotificationService {
	return &NotificationService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		retries: 3,
		timeout: 30 * time.Second,
	}
}

// SendBudgetAlert sends a budget alert through configured notification channels
func (n *NotificationService) SendBudgetAlert(alert BudgetAlert) {
	// In a real implementation, this would look up the budget to get notification channels
	// For now, we'll create a sample implementation

	// Create notification message
	message := n.formatAlertMessage(alert)

	// Log the alert (always done)
	log.Printf("BUDGET ALERT [%s]: %s", strings.ToUpper(alert.AlertType), message)

	// Send to configured channels
	// This would be expanded to support multiple notification types
	n.sendLogNotification(alert, message)
}

// formatAlertMessage creates a human-readable message for the alert
func (n *NotificationService) formatAlertMessage(alert BudgetAlert) string {
	switch alert.AlertType {
	case "budget_warning":
		return fmt.Sprintf("Budget warning for tenant %s: $%.2f/$%.2f (%.1f%%) - Consider optimizing job costs",
			alert.TenantID, alert.CurrentSpend, alert.BudgetAmount, alert.Utilization*100)

	case "budget_throttle":
		return fmt.Sprintf("Budget throttling active for tenant %s: $%.2f/$%.2f (%.1f%%) - Job processing slowed",
			alert.TenantID, alert.CurrentSpend, alert.BudgetAmount, alert.Utilization*100)

	case "budget_exceeded":
		return fmt.Sprintf("Budget exceeded for tenant %s: $%.2f/$%.2f (%.1f%%) - New jobs blocked",
			alert.TenantID, alert.CurrentSpend, alert.BudgetAmount, alert.Utilization*100)

	case "budget_reset":
		return fmt.Sprintf("Budget reset for tenant %s - New budget period started",
			alert.TenantID)

	default:
		return fmt.Sprintf("Budget alert for tenant %s: %s", alert.TenantID, alert.AlertType)
	}
}

// sendLogNotification logs the notification (always enabled)
func (n *NotificationService) sendLogNotification(alert BudgetAlert, message string) {
	logData := map[string]interface{}{
		"timestamp":     time.Now().Format(time.RFC3339),
		"type":          "budget_alert",
		"tenant_id":     alert.TenantID,
		"queue_name":    alert.QueueName,
		"alert_type":    alert.AlertType,
		"current_spend": alert.CurrentSpend,
		"budget_amount": alert.BudgetAmount,
		"utilization":   alert.Utilization,
		"message":       message,
	}

	logJSON, _ := json.Marshal(logData)
	log.Printf("BUDGET_NOTIFICATION: %s", string(logJSON))
}

// SendEmailNotification sends an email notification
func (n *NotificationService) SendEmailNotification(channel NotificationChannel, alert BudgetAlert) error {
	// This would integrate with an email service like SendGrid, SES, etc.
	// For now, we'll create a mock implementation

	subject := fmt.Sprintf("Budget Alert: %s - %s", strings.Title(alert.AlertType), alert.TenantID)
	body := n.formatEmailBody(alert)

	// Mock email sending
	log.Printf("EMAIL NOTIFICATION to %s: Subject: %s, Body: %s", channel.Target, subject, body)

	return nil
}

// formatEmailBody creates an HTML email body for the alert
func (n *NotificationService) formatEmailBody(alert BudgetAlert) string {
	return fmt.Sprintf(`
		<h2>Budget Alert: %s</h2>
		<p><strong>Tenant:</strong> %s</p>
		<p><strong>Queue:</strong> %s</p>
		<p><strong>Current Spend:</strong> $%.2f</p>
		<p><strong>Budget Amount:</strong> $%.2f</p>
		<p><strong>Utilization:</strong> %.1f%%</p>
		<p><strong>Alert Type:</strong> %s</p>
		<p><strong>Time:</strong> %s</p>

		<h3>Recommended Actions:</h3>
		%s
	`,
		strings.Title(alert.AlertType),
		alert.TenantID,
		alert.QueueName,
		alert.CurrentSpend,
		alert.BudgetAmount,
		alert.Utilization*100,
		alert.AlertType,
		alert.CreatedAt.Format("2006-01-02 15:04:05 UTC"),
		n.getRecommendedActions(alert),
	)
}

// SendSlackNotification sends a Slack notification
func (n *NotificationService) SendSlackNotification(channel NotificationChannel, alert BudgetAlert) error {
	webhookURL := channel.Target

	color := n.getSlackColor(alert.AlertType)

	payload := map[string]interface{}{
		"text": fmt.Sprintf("Budget Alert: %s", alert.TenantID),
		"attachments": []map[string]interface{}{
			{
				"color":  color,
				"title":  fmt.Sprintf("Budget %s Alert", strings.Title(alert.AlertType)),
				"fields": []map[string]interface{}{
					{"title": "Tenant", "value": alert.TenantID, "short": true},
					{"title": "Queue", "value": alert.QueueName, "short": true},
					{"title": "Current Spend", "value": fmt.Sprintf("$%.2f", alert.CurrentSpend), "short": true},
					{"title": "Budget", "value": fmt.Sprintf("$%.2f", alert.BudgetAmount), "short": true},
					{"title": "Utilization", "value": fmt.Sprintf("%.1f%%", alert.Utilization*100), "short": true},
					{"title": "Alert Type", "value": alert.AlertType, "short": true},
				},
				"footer": "Budget Monitoring System",
				"ts":     alert.CreatedAt.Unix(),
			},
		},
	}

	return n.sendWebhook(webhookURL, payload)
}

// SendWebhookNotification sends a generic webhook notification
func (n *NotificationService) SendWebhookNotification(channel NotificationChannel, alert BudgetAlert) error {
	webhookURL := channel.Target

	payload := map[string]interface{}{
		"event_type":    "budget_alert",
		"tenant_id":     alert.TenantID,
		"queue_name":    alert.QueueName,
		"alert_type":    alert.AlertType,
		"current_spend": alert.CurrentSpend,
		"budget_amount": alert.BudgetAmount,
		"utilization":   alert.Utilization,
		"timestamp":     alert.CreatedAt.Format(time.RFC3339),
		"message":       n.formatAlertMessage(alert),
	}

	// Add any channel-specific metadata
	for key, value := range channel.Metadata {
		payload[key] = value
	}

	return n.sendWebhook(webhookURL, payload)
}

// sendWebhook sends an HTTP POST request with JSON payload
func (n *NotificationService) sendWebhook(url string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < n.retries; attempt++ {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create webhook request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Budget-Monitor/1.0")

		resp, err := n.httpClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * time.Second) // Exponential backoff
			continue
		}

		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil // Success
		}

		lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			// Client error, don't retry
			break
		}

		time.Sleep(time.Duration(attempt+1) * time.Second)
	}

	return fmt.Errorf("webhook failed after %d attempts: %w", n.retries, lastErr)
}

// getSlackColor returns appropriate Slack message color for alert type
func (n *NotificationService) getSlackColor(alertType string) string {
	switch alertType {
	case "budget_warning":
		return "warning"
	case "budget_throttle":
		return "warning"
	case "budget_exceeded":
		return "danger"
	case "budget_reset":
		return "good"
	default:
		return "warning"
	}
}

// getRecommendedActions returns context-specific recommendations
func (n *NotificationService) getRecommendedActions(alert BudgetAlert) string {
	switch alert.AlertType {
	case "budget_warning":
		return `
			<ul>
				<li>Review job efficiency and optimize expensive operations</li>
				<li>Consider spreading workload across time to reduce peak costs</li>
				<li>Monitor daily spend trends to identify cost drivers</li>
			</ul>
		`
	case "budget_throttle":
		return `
			<ul>
				<li>Job processing is now throttled to 50% capacity</li>
				<li>Optimize job algorithms to reduce CPU and memory usage</li>
				<li>Consider increasing budget if current spend is justified</li>
				<li>Review cost breakdown to identify optimization opportunities</li>
			</ul>
		`
	case "budget_exceeded":
		return `
			<ul>
				<li>New jobs are now blocked to prevent further overspend</li>
				<li>Immediate action required: increase budget or optimize costs</li>
				<li>Review recent job patterns for unusual activity</li>
				<li>Consider emergency budget increase if critical operations are affected</li>
			</ul>
		`
	default:
		return "<p>Review budget configuration and spending patterns.</p>"
	}
}

// SendTestNotification sends a test notification to verify channel configuration
func (n *NotificationService) SendTestNotification(channel NotificationChannel) error {
	testAlert := BudgetAlert{
		TenantID:     "test-tenant",
		QueueName:    "test-queue",
		AlertType:    "test",
		CurrentSpend: 50.0,
		BudgetAmount: 100.0,
		Utilization:  0.5,
		CreatedAt:    time.Now(),
	}

	switch channel.Type {
	case "email":
		return n.SendEmailNotification(channel, testAlert)
	case "slack":
		return n.SendSlackNotification(channel, testAlert)
	case "webhook":
		return n.SendWebhookNotification(channel, testAlert)
	default:
		return fmt.Errorf("unsupported notification channel type: %s", channel.Type)
	}
}

// ValidateNotificationChannel validates a notification channel configuration
func (n *NotificationService) ValidateNotificationChannel(channel NotificationChannel) error {
	if channel.Type == "" {
		return NewValidationError("type", channel.Type, "required", "notification type is required")
	}

	if channel.Target == "" {
		return NewValidationError("target", channel.Target, "required", "notification target is required")
	}

	switch channel.Type {
	case "email":
		if !strings.Contains(channel.Target, "@") {
			return NewValidationError("target", channel.Target, "format", "email target must be a valid email address")
		}
	case "slack":
		if !strings.HasPrefix(channel.Target, "https://hooks.slack.com/") {
			return NewValidationError("target", channel.Target, "format", "Slack target must be a valid webhook URL")
		}
	case "webhook":
		if !strings.HasPrefix(channel.Target, "http://") && !strings.HasPrefix(channel.Target, "https://") {
			return NewValidationError("target", channel.Target, "format", "webhook target must be a valid URL")
		}
	default:
		return NewValidationError("type", channel.Type, "supported", "supported types: email, slack, webhook")
	}

	return nil
}