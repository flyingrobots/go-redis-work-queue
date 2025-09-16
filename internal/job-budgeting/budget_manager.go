package budgeting

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// BudgetManager handles budget creation, updates, and enforcement
type BudgetManager struct {
	db         *sql.DB
	aggregator *CostAggregator
	enforcer   *BudgetEnforcer
	notifier   *NotificationService
}

// NewBudgetManager creates a new budget manager
func NewBudgetManager(db *sql.DB, aggregator *CostAggregator, notifier *NotificationService) *BudgetManager {
	manager := &BudgetManager{
		db:         db,
		aggregator: aggregator,
		notifier:   notifier,
	}

	manager.enforcer = NewBudgetEnforcer(manager, notifier)
	return manager
}

// CreateBudget creates a new budget
func (b *BudgetManager) CreateBudget(budget *Budget) error {
	if err := b.validateBudget(budget); err != nil {
		return err
	}

	// Generate ID if not provided
	if budget.ID == "" {
		budget.ID = uuid.New().String()
	}

	budget.CreatedAt = time.Now()
	budget.UpdatedAt = time.Now()
	budget.Active = true

	// Serialize enforcement policy and notifications
	enforcementJSON, err := json.Marshal(budget.EnforcementPolicy)
	if err != nil {
		return fmt.Errorf("failed to serialize enforcement policy: %w", err)
	}

	notificationsJSON, err := json.Marshal(budget.Notifications)
	if err != nil {
		return fmt.Errorf("failed to serialize notifications: %w", err)
	}

	tagsJSON, err := json.Marshal(budget.Tags)
	if err != nil {
		return fmt.Errorf("failed to serialize tags: %w", err)
	}

	query := `
		INSERT INTO budgets (
			id, tenant_id, queue_name, period_type, period_start, period_end,
			amount, currency, warning_threshold, throttle_threshold, block_threshold,
			enforcement_policy, notifications, tags, active, created_at, updated_at, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
	`

	_, err = b.db.Exec(query,
		budget.ID,
		budget.TenantID,
		budget.QueueName,
		budget.Period.Type,
		budget.Period.StartDate,
		budget.Period.EndDate,
		budget.Amount,
		budget.Currency,
		budget.WarningThreshold,
		budget.ThrottleThreshold,
		budget.BlockThreshold,
		enforcementJSON,
		notificationsJSON,
		tagsJSON,
		budget.Active,
		budget.CreatedAt,
		budget.UpdatedAt,
		budget.CreatedBy,
	)

	if err != nil {
		return NewBudgetError(budget.ID, budget.TenantID, budget.QueueName, "create", err)
	}

	return nil
}

// GetBudget retrieves a budget by ID
func (b *BudgetManager) GetBudget(budgetID string) (*Budget, error) {
	query := `
		SELECT
			id, tenant_id, queue_name, period_type, period_start, period_end,
			amount, currency, warning_threshold, throttle_threshold, block_threshold,
			enforcement_policy, notifications, tags, active, created_at, updated_at, created_by
		FROM budgets
		WHERE id = $1
	`

	budget := &Budget{}
	var enforcementJSON, notificationsJSON, tagsJSON []byte

	err := b.db.QueryRow(query, budgetID).Scan(
		&budget.ID,
		&budget.TenantID,
		&budget.QueueName,
		&budget.Period.Type,
		&budget.Period.StartDate,
		&budget.Period.EndDate,
		&budget.Amount,
		&budget.Currency,
		&budget.WarningThreshold,
		&budget.ThrottleThreshold,
		&budget.BlockThreshold,
		&enforcementJSON,
		&notificationsJSON,
		&tagsJSON,
		&budget.Active,
		&budget.CreatedAt,
		&budget.UpdatedAt,
		&budget.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBudgetNotFound
		}
		return nil, err
	}

	// Deserialize JSON fields
	if err := json.Unmarshal(enforcementJSON, &budget.EnforcementPolicy); err != nil {
		return nil, fmt.Errorf("failed to deserialize enforcement policy: %w", err)
	}

	if err := json.Unmarshal(notificationsJSON, &budget.Notifications); err != nil {
		return nil, fmt.Errorf("failed to deserialize notifications: %w", err)
	}

	if err := json.Unmarshal(tagsJSON, &budget.Tags); err != nil {
		return nil, fmt.Errorf("failed to deserialize tags: %w", err)
	}

	return budget, nil
}

// GetBudgetForTenant retrieves the active budget for a tenant/queue
func (b *BudgetManager) GetBudgetForTenant(tenantID, queueName string) (*Budget, error) {
	query := `
		SELECT
			id, tenant_id, queue_name, period_type, period_start, period_end,
			amount, currency, warning_threshold, throttle_threshold, block_threshold,
			enforcement_policy, notifications, tags, active, created_at, updated_at, created_by
		FROM budgets
		WHERE tenant_id = $1 AND (queue_name = $2 OR queue_name = '') AND active = true
		ORDER BY queue_name DESC, created_at DESC
		LIMIT 1
	`

	budget := &Budget{}
	var enforcementJSON, notificationsJSON, tagsJSON []byte

	err := b.db.QueryRow(query, tenantID, queueName).Scan(
		&budget.ID,
		&budget.TenantID,
		&budget.QueueName,
		&budget.Period.Type,
		&budget.Period.StartDate,
		&budget.Period.EndDate,
		&budget.Amount,
		&budget.Currency,
		&budget.WarningThreshold,
		&budget.ThrottleThreshold,
		&budget.BlockThreshold,
		&enforcementJSON,
		&notificationsJSON,
		&tagsJSON,
		&budget.Active,
		&budget.CreatedAt,
		&budget.UpdatedAt,
		&budget.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBudgetNotFound
		}
		return nil, err
	}

	// Deserialize JSON fields
	if err := json.Unmarshal(enforcementJSON, &budget.EnforcementPolicy); err != nil {
		return nil, fmt.Errorf("failed to deserialize enforcement policy: %w", err)
	}

	if err := json.Unmarshal(notificationsJSON, &budget.Notifications); err != nil {
		return nil, fmt.Errorf("failed to deserialize notifications: %w", err)
	}

	if err := json.Unmarshal(tagsJSON, &budget.Tags); err != nil {
		return nil, fmt.Errorf("failed to deserialize tags: %w", err)
	}

	return budget, nil
}

// UpdateBudget updates an existing budget
func (b *BudgetManager) UpdateBudget(budget *Budget) error {
	if err := b.validateBudget(budget); err != nil {
		return err
	}

	budget.UpdatedAt = time.Now()

	// Serialize JSON fields
	enforcementJSON, err := json.Marshal(budget.EnforcementPolicy)
	if err != nil {
		return fmt.Errorf("failed to serialize enforcement policy: %w", err)
	}

	notificationsJSON, err := json.Marshal(budget.Notifications)
	if err != nil {
		return fmt.Errorf("failed to serialize notifications: %w", err)
	}

	tagsJSON, err := json.Marshal(budget.Tags)
	if err != nil {
		return fmt.Errorf("failed to serialize tags: %w", err)
	}

	query := `
		UPDATE budgets SET
			tenant_id = $2, queue_name = $3, period_type = $4, period_start = $5, period_end = $6,
			amount = $7, currency = $8, warning_threshold = $9, throttle_threshold = $10,
			block_threshold = $11, enforcement_policy = $12, notifications = $13, tags = $14,
			active = $15, updated_at = $16
		WHERE id = $1
	`

	result, err := b.db.Exec(query,
		budget.ID,
		budget.TenantID,
		budget.QueueName,
		budget.Period.Type,
		budget.Period.StartDate,
		budget.Period.EndDate,
		budget.Amount,
		budget.Currency,
		budget.WarningThreshold,
		budget.ThrottleThreshold,
		budget.BlockThreshold,
		enforcementJSON,
		notificationsJSON,
		tagsJSON,
		budget.Active,
		budget.UpdatedAt,
	)

	if err != nil {
		return NewBudgetError(budget.ID, budget.TenantID, budget.QueueName, "update", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrBudgetNotFound
	}

	return nil
}

// DeleteBudget deletes a budget
func (b *BudgetManager) DeleteBudget(budgetID string) error {
	query := `DELETE FROM budgets WHERE id = $1`

	result, err := b.db.Exec(query, budgetID)
	if err != nil {
		return NewBudgetError(budgetID, "", "", "delete", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrBudgetNotFound
	}

	return nil
}

// ListBudgets returns all budgets for a tenant
func (b *BudgetManager) ListBudgets(tenantID string, activeOnly bool) ([]Budget, error) {
	query := `
		SELECT
			id, tenant_id, queue_name, period_type, period_start, period_end,
			amount, currency, warning_threshold, throttle_threshold, block_threshold,
			enforcement_policy, notifications, tags, active, created_at, updated_at, created_by
		FROM budgets
		WHERE tenant_id = $1
	`

	args := []interface{}{tenantID}

	if activeOnly {
		query += " AND active = true"
	}

	query += " ORDER BY created_at DESC"

	rows, err := b.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []Budget
	for rows.Next() {
		budget := Budget{}
		var enforcementJSON, notificationsJSON, tagsJSON []byte

		err := rows.Scan(
			&budget.ID,
			&budget.TenantID,
			&budget.QueueName,
			&budget.Period.Type,
			&budget.Period.StartDate,
			&budget.Period.EndDate,
			&budget.Amount,
			&budget.Currency,
			&budget.WarningThreshold,
			&budget.ThrottleThreshold,
			&budget.BlockThreshold,
			&enforcementJSON,
			&notificationsJSON,
			&tagsJSON,
			&budget.Active,
			&budget.CreatedAt,
			&budget.UpdatedAt,
			&budget.CreatedBy,
		)

		if err != nil {
			return nil, err
		}

		// Deserialize JSON fields
		if err := json.Unmarshal(enforcementJSON, &budget.EnforcementPolicy); err != nil {
			return nil, fmt.Errorf("failed to deserialize enforcement policy: %w", err)
		}

		if err := json.Unmarshal(notificationsJSON, &budget.Notifications); err != nil {
			return nil, fmt.Errorf("failed to deserialize notifications: %w", err)
		}

		if err := json.Unmarshal(tagsJSON, &budget.Tags); err != nil {
			return nil, fmt.Errorf("failed to deserialize tags: %w", err)
		}

		budgets = append(budgets, budget)
	}

	return budgets, rows.Err()
}

// GetCurrentSpend returns the current spending for a budget
func (b *BudgetManager) GetCurrentSpend(tenantID, queueName string) (float64, error) {
	budget, err := b.GetBudgetForTenant(tenantID, queueName)
	if err != nil {
		return 0, err
	}

	return b.aggregator.GetCurrentSpend(tenantID, queueName, budget.Period)
}

// GetBudgetStatus returns the current status of a budget
func (b *BudgetManager) GetBudgetStatus(budgetID string) (*BudgetStatus, error) {
	budget, err := b.GetBudget(budgetID)
	if err != nil {
		return nil, err
	}

	currentSpend, err := b.aggregator.GetCurrentSpend(budget.TenantID, budget.QueueName, budget.Period)
	if err != nil {
		return nil, err
	}

	// Calculate period metrics
	now := time.Now()
	periodStart := budget.Period.StartDate
	periodEnd := budget.Period.EndDate

	totalDays := int(periodEnd.Sub(periodStart).Hours() / 24)
	daysElapsed := int(now.Sub(periodStart).Hours() / 24)
	daysRemaining := int(periodEnd.Sub(now).Hours() / 24)

	if daysRemaining < 0 {
		daysRemaining = 0
	}

	if daysElapsed <= 0 {
		daysElapsed = 1 // Avoid division by zero
	}

	dailyBurnRate := currentSpend / float64(daysElapsed)
	projectedSpend := dailyBurnRate * float64(totalDays)
	utilization := currentSpend / budget.Amount

	// Determine current threshold
	currentThreshold := "none"
	if utilization >= budget.BlockThreshold {
		currentThreshold = "block"
	} else if utilization >= budget.ThrottleThreshold {
		currentThreshold = "throttle"
	} else if utilization >= budget.WarningThreshold {
		currentThreshold = "warning"
	}

	return &BudgetStatus{
		BudgetID:         budget.ID,
		TenantID:         budget.TenantID,
		QueueName:        budget.QueueName,
		CurrentSpend:     currentSpend,
		BudgetAmount:     budget.Amount,
		Utilization:      utilization,
		DaysInPeriod:     totalDays,
		DaysRemaining:    daysRemaining,
		DailyBurnRate:    dailyBurnRate,
		ProjectedSpend:   projectedSpend,
		IsOverBudget:     currentSpend > budget.Amount,
		CurrentThreshold: currentThreshold,
		UpdatedAt:        time.Now(),
	}, nil
}

// CheckBudgetCompliance checks if a tenant/queue is compliant with budget limits
func (b *BudgetManager) CheckBudgetCompliance(tenantID, queueName string) (EnforcementAction, error) {
	return b.enforcer.CheckBudgetCompliance(tenantID, queueName)
}

// validateBudget validates budget configuration
func (b *BudgetManager) validateBudget(budget *Budget) error {
	if budget.TenantID == "" {
		return NewValidationError("tenant_id", budget.TenantID, "required", "tenant ID is required")
	}

	if budget.Amount <= 0 {
		return NewValidationError("amount", budget.Amount, "positive", "budget amount must be positive")
	}

	if budget.WarningThreshold < 0 || budget.WarningThreshold > 1 {
		return NewValidationError("warning_threshold", budget.WarningThreshold, "range", "warning threshold must be between 0 and 1")
	}

	if budget.ThrottleThreshold < 0 || budget.ThrottleThreshold > 1 {
		return NewValidationError("throttle_threshold", budget.ThrottleThreshold, "range", "throttle threshold must be between 0 and 1")
	}

	if budget.BlockThreshold < 0 || budget.BlockThreshold > 1 {
		return NewValidationError("block_threshold", budget.BlockThreshold, "range", "block threshold must be between 0 and 1")
	}

	if budget.WarningThreshold > budget.ThrottleThreshold {
		return NewValidationError("warning_threshold", budget.WarningThreshold, "order", "warning threshold must be <= throttle threshold")
	}

	if budget.ThrottleThreshold > budget.BlockThreshold {
		return NewValidationError("throttle_threshold", budget.ThrottleThreshold, "order", "throttle threshold must be <= block threshold")
	}

	if budget.Period.StartDate.After(budget.Period.EndDate) {
		return NewValidationError("period", budget.Period, "order", "period start date must be before end date")
	}

	if budget.EnforcementPolicy.ThrottleFactor < 0 || budget.EnforcementPolicy.ThrottleFactor > 1 {
		return NewValidationError("throttle_factor", budget.EnforcementPolicy.ThrottleFactor, "range", "throttle factor must be between 0 and 1")
	}

	// Validate notification channels
	for i, channel := range budget.Notifications {
		if channel.Type == "" {
			return NewValidationError(fmt.Sprintf("notifications[%d].type", i), channel.Type, "required", "notification type is required")
		}

		if channel.Target == "" {
			return NewValidationError(fmt.Sprintf("notifications[%d].target", i), channel.Target, "required", "notification target is required")
		}
	}

	return nil
}