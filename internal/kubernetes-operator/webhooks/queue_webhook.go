package webhooks

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	queuev1 "github.com/example/go-redis-work-queue/internal/kubernetes-operator/apis/v1"
)

// QueueWebhook handles validation and mutation for Queue resources
type QueueWebhook struct {
	Client  client.Client
	decoder *admission.Decoder
}

// +kubebuilder:webhook:path=/validate-queue-example-com-v1-queue,mutating=false,failurePolicy=fail,sideEffects=None,groups=queue.example.com,resources=queues,verbs=create;update,versions=v1,name=vqueue.kb.io,admissionReviewVersions=v1

// +kubebuilder:webhook:path=/mutate-queue-example-com-v1-queue,mutating=true,failurePolicy=fail,sideEffects=None,groups=queue.example.com,resources=queues,verbs=create;update,versions=v1,name=mqueue.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &QueueWebhook{}
var _ webhook.Defaulter = &QueueWebhook{}

// ValidateCreate implements webhook.Validator
func (w *QueueWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	queue, ok := obj.(*queuev1.Queue)
	if !ok {
		return fmt.Errorf("expected a Queue object")
	}

	return w.validateQueue(ctx, queue, nil)
}

// ValidateUpdate implements webhook.Validator
func (w *QueueWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	newQueue, ok := newObj.(*queuev1.Queue)
	if !ok {
		return fmt.Errorf("expected a Queue object")
	}

	oldQueue, ok := oldObj.(*queuev1.Queue)
	if !ok {
		return fmt.Errorf("expected a Queue object")
	}

	return w.validateQueue(ctx, newQueue, oldQueue)
}

// ValidateDelete implements webhook.Validator
func (w *QueueWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	// Allow all deletions
	return nil
}

// Default implements webhook.Defaulter
func (w *QueueWebhook) Default(ctx context.Context, obj runtime.Object) error {
	queue, ok := obj.(*queuev1.Queue)
	if !ok {
		return fmt.Errorf("expected a Queue object")
	}

	return w.setDefaults(queue)
}

// validateQueue performs comprehensive validation
func (w *QueueWebhook) validateQueue(ctx context.Context, queue *queuev1.Queue, oldQueue *queuev1.Queue) error {
	// Validate queue name
	if err := w.validateQueueName(queue.Spec.Name); err != nil {
		return err
	}

	// Validate priority
	if err := w.validatePriority(queue.Spec.Priority); err != nil {
		return err
	}

	// Validate rate limit configuration
	if queue.Spec.RateLimit != nil {
		if err := w.validateRateLimit(queue.Spec.RateLimit); err != nil {
			return err
		}
	}

	// Validate dead letter queue configuration
	if queue.Spec.DeadLetterQueue != nil {
		if err := w.validateDeadLetterQueue(queue.Spec.DeadLetterQueue); err != nil {
			return err
		}
	}

	// Validate retention configuration
	if queue.Spec.Retention != nil {
		if err := w.validateRetention(queue.Spec.Retention); err != nil {
			return err
		}
	}

	// Validate Redis configuration
	if queue.Spec.Redis != nil {
		if err := w.validateRedis(ctx, queue.Spec.Redis, queue.Namespace); err != nil {
			return err
		}
	}

	// Validate immutable fields during updates
	if oldQueue != nil {
		if err := w.validateImmutableFields(queue, oldQueue); err != nil {
			return err
		}
	}

	// Check for naming conflicts
	if err := w.validateNamingConflicts(ctx, queue); err != nil {
		return err
	}

	return nil
}

// validateQueueName validates the queue name format
func (w *QueueWebhook) validateQueueName(name string) error {
	if name == "" {
		return fmt.Errorf("queue name cannot be empty")
	}

	// Check for reserved names
	reservedNames := []string{
		"system", "admin", "health", "metrics", "default",
		"kube-system", "kube-public", "kube-node-lease",
	}

	for _, reserved := range reservedNames {
		if name == reserved {
			return fmt.Errorf("queue name '%s' is reserved", name)
		}
	}

	// Check format (already handled by kubebuilder validation, but double-check)
	if len(name) > 63 {
		return fmt.Errorf("queue name cannot exceed 63 characters")
	}

	if !isValidDNSSubdomain(name) {
		return fmt.Errorf("queue name must be a valid DNS subdomain")
	}

	return nil
}

// validatePriority validates the priority value
func (w *QueueWebhook) validatePriority(priority string) error {
	validPriorities := []string{"critical", "high", "medium", "low"}

	for _, valid := range validPriorities {
		if priority == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid priority '%s', must be one of: %s", priority, strings.Join(validPriorities, ", "))
}

// validateRateLimit validates rate limiting configuration
func (w *QueueWebhook) validateRateLimit(rateLimit *queuev1.RateLimitSpec) error {
	if rateLimit.RequestsPerSecond <= 0 {
		return fmt.Errorf("requestsPerSecond must be positive")
	}

	if rateLimit.RequestsPerSecond > 100000 {
		return fmt.Errorf("requestsPerSecond cannot exceed 100000")
	}

	if rateLimit.BurstCapacity <= 0 {
		return fmt.Errorf("burstCapacity must be positive")
	}

	if rateLimit.BurstCapacity > 100000 {
		return fmt.Errorf("burstCapacity cannot exceed 100000")
	}

	// Burst capacity should be at least as large as rate per second
	if float64(rateLimit.BurstCapacity) < rateLimit.RequestsPerSecond {
		return fmt.Errorf("burstCapacity should be at least as large as requestsPerSecond")
	}

	return nil
}

// validateDeadLetterQueue validates DLQ configuration
func (w *QueueWebhook) validateDeadLetterQueue(dlq *queuev1.DeadLetterQueueSpec) error {
	if dlq.MaxRetries < 0 || dlq.MaxRetries > 100 {
		return fmt.Errorf("maxRetries must be between 0 and 100")
	}

	if dlq.RetryBackoff != nil {
		if err := w.validateRetryBackoff(dlq.RetryBackoff); err != nil {
			return err
		}
	}

	return nil
}

// validateRetryBackoff validates retry backoff configuration
func (w *QueueWebhook) validateRetryBackoff(backoff *queuev1.RetryBackoffSpec) error {
	if backoff.InitialDelay.Duration <= 0 {
		return fmt.Errorf("initialDelay must be positive")
	}

	if backoff.MaxDelay.Duration <= 0 {
		return fmt.Errorf("maxDelay must be positive")
	}

	if backoff.MaxDelay.Duration < backoff.InitialDelay.Duration {
		return fmt.Errorf("maxDelay must be greater than or equal to initialDelay")
	}

	if backoff.Multiplier < 1.0 || backoff.Multiplier > 10.0 {
		return fmt.Errorf("multiplier must be between 1.0 and 10.0")
	}

	return nil
}

// validateRetention validates retention configuration
func (w *QueueWebhook) validateRetention(retention *queuev1.RetentionSpec) error {
	if retention.CompletedJobs.Duration <= 0 {
		return fmt.Errorf("completedJobs retention must be positive")
	}

	if retention.FailedJobs.Duration <= 0 {
		return fmt.Errorf("failedJobs retention must be positive")
	}

	if retention.MaxJobs < 100 || retention.MaxJobs > 1000000 {
		return fmt.Errorf("maxJobs must be between 100 and 1000000")
	}

	return nil
}

// validateRedis validates Redis configuration
func (w *QueueWebhook) validateRedis(ctx context.Context, redis *queuev1.RedisSpec, namespace string) error {
	if len(redis.Addresses) == 0 {
		return fmt.Errorf("redis addresses cannot be empty")
	}

	for _, addr := range redis.Addresses {
		if addr == "" {
			return fmt.Errorf("redis address cannot be empty")
		}
		if !isValidAddress(addr) {
			return fmt.Errorf("invalid redis address format: %s", addr)
		}
	}

	if redis.Database < 0 || redis.Database > 15 {
		return fmt.Errorf("redis database must be between 0 and 15")
	}

	// Validate secret references if provided
	if redis.PasswordSecret != nil {
		if err := w.validateSecretReference(ctx, redis.PasswordSecret, namespace); err != nil {
			return fmt.Errorf("invalid password secret: %w", err)
		}
	}

	if redis.TLS != nil && redis.TLS.CASecret != nil {
		if err := w.validateSecretReference(ctx, redis.TLS.CASecret, namespace); err != nil {
			return fmt.Errorf("invalid CA secret: %w", err)
		}
	}

	return nil
}

// validateSecretReference validates that a secret reference is valid
func (w *QueueWebhook) validateSecretReference(ctx context.Context, secretRef *queuev1.SecretKeySelector, namespace string) error {
	if secretRef.Name == "" {
		return fmt.Errorf("secret name cannot be empty")
	}

	if secretRef.Key == "" {
		return fmt.Errorf("secret key cannot be empty")
	}

	// Optional: Check if secret exists (might be created later)
	// This could be a configurable behavior

	return nil
}

// validateImmutableFields ensures immutable fields haven't changed
func (w *QueueWebhook) validateImmutableFields(newQueue, oldQueue *queuev1.Queue) error {
	// Queue name is immutable
	if newQueue.Spec.Name != oldQueue.Spec.Name {
		return fmt.Errorf("queue name is immutable")
	}

	// Redis configuration is immutable (for safety)
	if !equalRedisConfigs(newQueue.Spec.Redis, oldQueue.Spec.Redis) {
		return fmt.Errorf("redis configuration is immutable")
	}

	return nil
}

// validateNamingConflicts checks for naming conflicts
func (w *QueueWebhook) validateNamingConflicts(ctx context.Context, queue *queuev1.Queue) error {
	// Check for duplicate queue names in the same namespace
	queueList := &queuev1.QueueList{}
	if err := w.Client.List(ctx, queueList, client.InNamespace(queue.Namespace)); err != nil {
		// If we can't check for conflicts, allow the request
		return nil
	}

	for _, existingQueue := range queueList.Items {
		if existingQueue.Name != queue.Name && existingQueue.Spec.Name == queue.Spec.Name {
			return fmt.Errorf("queue name '%s' is already in use by queue resource '%s'", queue.Spec.Name, existingQueue.Name)
		}
	}

	return nil
}

// setDefaults sets default values for Queue fields
func (w *QueueWebhook) setDefaults(queue *queuev1.Queue) error {
	// Set default priority
	if queue.Spec.Priority == "" {
		queue.Spec.Priority = "medium"
	}

	// Set default rate limit if enabled but not configured
	if queue.Spec.RateLimit != nil && queue.Spec.RateLimit.RequestsPerSecond == 0 {
		queue.Spec.RateLimit.RequestsPerSecond = 100
		queue.Spec.RateLimit.BurstCapacity = 200
		queue.Spec.RateLimit.Enabled = true
	}

	// Set default DLQ configuration
	if queue.Spec.DeadLetterQueue != nil {
		if queue.Spec.DeadLetterQueue.MaxRetries == 0 {
			queue.Spec.DeadLetterQueue.MaxRetries = 3
		}

		if queue.Spec.DeadLetterQueue.RetryBackoff == nil {
			queue.Spec.DeadLetterQueue.RetryBackoff = &queuev1.RetryBackoffSpec{
				InitialDelay: metav1.Duration{Duration: time.Second},
				MaxDelay:     metav1.Duration{Duration: 30 * time.Second},
				Multiplier:   2.0,
			}
		}
	}

	// Set default retention
	if queue.Spec.Retention != nil {
		if queue.Spec.Retention.CompletedJobs.Duration == 0 {
			queue.Spec.Retention.CompletedJobs = metav1.Duration{Duration: 24 * time.Hour}
		}
		if queue.Spec.Retention.FailedJobs.Duration == 0 {
			queue.Spec.Retention.FailedJobs = metav1.Duration{Duration: 72 * time.Hour}
		}
		if queue.Spec.Retention.MaxJobs == 0 {
			queue.Spec.Retention.MaxJobs = 10000
		}
	}

	// Set default Redis database
	if queue.Spec.Redis != nil && queue.Spec.Redis.Database == 0 {
		queue.Spec.Redis.Database = 0
	}

	return nil
}

// Helper functions

// isValidDNSSubdomain checks if a string is a valid DNS subdomain
func isValidDNSSubdomain(name string) bool {
	if len(name) == 0 || len(name) > 253 {
		return false
	}

	// Simple validation - in reality, this would be more comprehensive
	return !strings.Contains(name, " ") && !strings.Contains(name, "..")
}

// isValidAddress checks if an address is valid (host:port format)
func isValidAddress(addr string) bool {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return false
	}

	host, port := parts[0], parts[1]
	if host == "" || port == "" {
		return false
	}

	// Basic port validation
	if len(port) == 0 || len(port) > 5 {
		return false
	}

	return true
}

// equalRedisConfigs compares two Redis configurations
func equalRedisConfigs(a, b *queuev1.RedisSpec) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Compare addresses
	if len(a.Addresses) != len(b.Addresses) {
		return false
	}
	for i, addr := range a.Addresses {
		if addr != b.Addresses[i] {
			return false
		}
	}

	// Compare other fields
	return a.Database == b.Database
}

// SetupWithManager sets up the webhook with the Manager
func (w *QueueWebhook) SetupWithManager(mgr ctrl.Manager) error {
	w.decoder = admission.NewDecoder(mgr.GetScheme())

	return ctrl.NewWebhookManagedBy(mgr).
		For(&queuev1.Queue{}).
		WithValidator(w).
		WithDefaulter(w).
		Complete()
}