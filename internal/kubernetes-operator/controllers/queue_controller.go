package controllers

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	queuev1 "github.com/flyingrobots/go-redis-work-queue/internal/kubernetes-operator/apis/v1"
)

// QueueReconciler reconciles a Queue object
type QueueReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	AdminAPIClient AdminAPIClient
}

// AdminAPIClient interface for interacting with the queue system
type AdminAPIClient interface {
	CreateQueue(ctx context.Context, config QueueConfig) error
	UpdateQueue(ctx context.Context, name string, config QueueConfig) error
	DeleteQueue(ctx context.Context, name string) error
	GetQueueMetrics(ctx context.Context, name string) (*QueueMetrics, error)
	GetQueueStatus(ctx context.Context, name string) (*QueueStatus, error)
}

// QueueConfig represents queue configuration for Admin API
type QueueConfig struct {
	Name            string                     `json:"name"`
	Priority        string                     `json:"priority"`
	RateLimit       *RateLimitConfig          `json:"rateLimit,omitempty"`
	DeadLetterQueue *DeadLetterQueueConfig    `json:"deadLetterQueue,omitempty"`
	Retention       *RetentionConfig          `json:"retention,omitempty"`
	Redis           *RedisConfig              `json:"redis,omitempty"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond float64 `json:"requestsPerSecond"`
	BurstCapacity     int32   `json:"burstCapacity"`
	Enabled           bool    `json:"enabled"`
}

// DeadLetterQueueConfig represents DLQ configuration
type DeadLetterQueueConfig struct {
	Enabled      bool                 `json:"enabled"`
	MaxRetries   int32               `json:"maxRetries"`
	RetryBackoff *RetryBackoffConfig `json:"retryBackoff,omitempty"`
}

// RetryBackoffConfig represents retry backoff configuration
type RetryBackoffConfig struct {
	InitialDelay time.Duration `json:"initialDelay"`
	MaxDelay     time.Duration `json:"maxDelay"`
	Multiplier   float64       `json:"multiplier"`
}

// RetentionConfig represents retention policy configuration
type RetentionConfig struct {
	CompletedJobs time.Duration `json:"completedJobs"`
	FailedJobs    time.Duration `json:"failedJobs"`
	MaxJobs       int32         `json:"maxJobs"`
}

// RedisConfig represents Redis connection configuration
type RedisConfig struct {
	Addresses []string `json:"addresses"`
	Database  int32    `json:"database"`
	Password  string   `json:"password,omitempty"`
	TLS       *TLSConfig `json:"tls,omitempty"`
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled            bool   `json:"enabled"`
	CAData             []byte `json:"caData,omitempty"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify"`
}

// QueueMetrics represents queue operational metrics
type QueueMetrics struct {
	BacklogSize    int64     `json:"backlogSize"`
	ProcessingRate float64   `json:"processingRate"`
	ErrorRate      float64   `json:"errorRate"`
	AverageLatency float64   `json:"averageLatency"`
	LastUpdated    time.Time `json:"lastUpdated"`
}

// QueueStatus represents queue status from Admin API
type QueueStatus struct {
	State       string `json:"state"`
	Message     string `json:"message,omitempty"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// +kubebuilder:rbac:groups=queue.example.com,resources=queues,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=queue.example.com,resources=queues/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=queue.example.com,resources=queues/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *QueueReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Queue instance
	var queue queuev1.Queue
	if err := r.Get(ctx, req.NamespacedName, &queue); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return without error
			logger.Info("Queue resource not found, ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object
		logger.Error(err, "Failed to get Queue")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if queue.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, &queue)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(&queue, QueueFinalizerName) {
		controllerutil.AddFinalizer(&queue, QueueFinalizerName)
		return ctrl.Result{}, r.Update(ctx, &queue)
	}

	// Reconcile the queue
	return r.reconcileQueue(ctx, &queue)
}

const QueueFinalizerName = "queue.example.com/finalizer"

// handleDeletion handles queue deletion with proper cleanup
func (r *QueueReconciler) handleDeletion(ctx context.Context, queue *queuev1.Queue) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(queue, QueueFinalizerName) {
		// Delete the queue from the Admin API
		if err := r.AdminAPIClient.DeleteQueue(ctx, queue.Spec.Name); err != nil {
			logger.Error(err, "Failed to delete queue from Admin API", "queue", queue.Spec.Name)
			// Update status to indicate deletion failure
			r.updateQueueStatus(ctx, queue, queuev1.QueuePhaseFailed, "Failed to delete queue from Admin API", nil)
			return ctrl.Result{RequeueAfter: time.Minute}, err
		}

		// Remove finalizer to allow deletion
		controllerutil.RemoveFinalizer(queue, QueueFinalizerName)
		if err := r.Update(ctx, queue); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// reconcileQueue handles the main reconciliation logic
func (r *QueueReconciler) reconcileQueue(ctx context.Context, queue *queuev1.Queue) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Convert CRD spec to Admin API config
	config, err := r.buildQueueConfig(ctx, queue)
	if err != nil {
		logger.Error(err, "Failed to build queue config", "queue", queue.Name)
		r.updateQueueStatus(ctx, queue, queuev1.QueuePhaseFailed, err.Error(), nil)
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}

	// Check if queue exists in Admin API
	status, err := r.AdminAPIClient.GetQueueStatus(ctx, queue.Spec.Name)
	if err != nil {
		// Queue doesn't exist, create it
		if err := r.AdminAPIClient.CreateQueue(ctx, *config); err != nil {
			logger.Error(err, "Failed to create queue", "queue", queue.Spec.Name)
			r.updateQueueStatus(ctx, queue, queuev1.QueuePhaseFailed, fmt.Sprintf("Failed to create queue: %v", err), nil)
			return ctrl.Result{RequeueAfter: time.Minute}, err
		}
		logger.Info("Created queue", "queue", queue.Spec.Name)
	} else {
		// Queue exists, update if needed
		if err := r.AdminAPIClient.UpdateQueue(ctx, queue.Spec.Name, *config); err != nil {
			logger.Error(err, "Failed to update queue", "queue", queue.Spec.Name)
			r.updateQueueStatus(ctx, queue, queuev1.QueuePhaseFailed, fmt.Sprintf("Failed to update queue: %v", err), nil)
			return ctrl.Result{RequeueAfter: time.Minute}, err
		}
		logger.Info("Updated queue", "queue", queue.Spec.Name)
	}

	// Get current metrics
	metrics, err := r.AdminAPIClient.GetQueueMetrics(ctx, queue.Spec.Name)
	if err != nil {
		logger.Error(err, "Failed to get queue metrics", "queue", queue.Spec.Name)
		// Don't fail reconciliation for metrics errors
		metrics = &QueueMetrics{}
	}

	// Update status
	queueMetrics := &queuev1.QueueMetrics{
		BacklogSize:    metrics.BacklogSize,
		ProcessingRate: metrics.ProcessingRate,
		ErrorRate:      metrics.ErrorRate,
		AverageLatency: metrics.AverageLatency,
		LastUpdated:    metav1.NewTime(metrics.LastUpdated),
	}

	r.updateQueueStatus(ctx, queue, queuev1.QueuePhaseActive, "Queue is active", queueMetrics)

	// Requeue for periodic status updates
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// buildQueueConfig converts CRD spec to Admin API config
func (r *QueueReconciler) buildQueueConfig(ctx context.Context, queue *queuev1.Queue) (*QueueConfig, error) {
	config := &QueueConfig{
		Name:     queue.Spec.Name,
		Priority: queue.Spec.Priority,
	}

	// Rate limit configuration
	if queue.Spec.RateLimit != nil {
		config.RateLimit = &RateLimitConfig{
			RequestsPerSecond: queue.Spec.RateLimit.RequestsPerSecond,
			BurstCapacity:     queue.Spec.RateLimit.BurstCapacity,
			Enabled:           queue.Spec.RateLimit.Enabled,
		}
	}

	// Dead letter queue configuration
	if queue.Spec.DeadLetterQueue != nil {
		dlqConfig := &DeadLetterQueueConfig{
			Enabled:    queue.Spec.DeadLetterQueue.Enabled,
			MaxRetries: queue.Spec.DeadLetterQueue.MaxRetries,
		}

		if queue.Spec.DeadLetterQueue.RetryBackoff != nil {
			dlqConfig.RetryBackoff = &RetryBackoffConfig{
				InitialDelay: queue.Spec.DeadLetterQueue.RetryBackoff.InitialDelay.Duration,
				MaxDelay:     queue.Spec.DeadLetterQueue.RetryBackoff.MaxDelay.Duration,
				Multiplier:   queue.Spec.DeadLetterQueue.RetryBackoff.Multiplier,
			}
		}

		config.DeadLetterQueue = dlqConfig
	}

	// Retention configuration
	if queue.Spec.Retention != nil {
		config.Retention = &RetentionConfig{
			CompletedJobs: queue.Spec.Retention.CompletedJobs.Duration,
			FailedJobs:    queue.Spec.Retention.FailedJobs.Duration,
			MaxJobs:       queue.Spec.Retention.MaxJobs,
		}
	}

	// Redis configuration
	if queue.Spec.Redis != nil {
		redisConfig := &RedisConfig{
			Addresses: queue.Spec.Redis.Addresses,
			Database:  queue.Spec.Redis.Database,
		}

		// Handle password secret
		if queue.Spec.Redis.PasswordSecret != nil {
			password, err := r.getSecretValue(ctx, queue.Namespace, queue.Spec.Redis.PasswordSecret)
			if err != nil {
				return nil, fmt.Errorf("failed to get Redis password: %w", err)
			}
			redisConfig.Password = password
		}

		// Handle TLS configuration
		if queue.Spec.Redis.TLS != nil {
			tlsConfig := &TLSConfig{
				Enabled:            queue.Spec.Redis.TLS.Enabled,
				InsecureSkipVerify: queue.Spec.Redis.TLS.InsecureSkipVerify,
			}

			if queue.Spec.Redis.TLS.CASecret != nil {
				caData, err := r.getSecretValue(ctx, queue.Namespace, queue.Spec.Redis.TLS.CASecret)
				if err != nil {
					return nil, fmt.Errorf("failed to get CA data: %w", err)
				}
				tlsConfig.CAData = []byte(caData)
			}

			redisConfig.TLS = tlsConfig
		}

		config.Redis = redisConfig
	}

	return config, nil
}

// getSecretValue retrieves a value from a Kubernetes secret
func (r *QueueReconciler) getSecretValue(ctx context.Context, namespace string, selector *corev1.SecretKeySelector) (string, error) {
	var secret corev1.Secret
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      selector.Name,
	}

	if err := r.Get(ctx, key, &secret); err != nil {
		return "", err
	}

	data, exists := secret.Data[selector.Key]
	if !exists {
		return "", fmt.Errorf("key %s not found in secret %s", selector.Key, selector.Name)
	}

	return string(data), nil
}

// updateQueueStatus updates the queue status
func (r *QueueReconciler) updateQueueStatus(ctx context.Context, queue *queuev1.Queue, phase queuev1.QueuePhase, message string, metrics *queuev1.QueueMetrics) {
	logger := log.FromContext(ctx)

	queue.Status.Phase = phase
	queue.Status.ObservedGeneration = queue.Generation

	// Update conditions
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "QueueReady",
		Message:            message,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}

	if phase == queuev1.QueuePhaseFailed {
		condition.Status = metav1.ConditionFalse
		condition.Reason = "QueueFailed"
	}

	// Update or add condition
	found := false
	for i, cond := range queue.Status.Conditions {
		if cond.Type == condition.Type {
			queue.Status.Conditions[i] = condition
			found = true
			break
		}
	}
	if !found {
		queue.Status.Conditions = append(queue.Status.Conditions, condition)
	}

	// Update metrics if provided
	if metrics != nil {
		queue.Status.Metrics = metrics
	}

	// Update status
	if err := r.Status().Update(ctx, queue); err != nil {
		logger.Error(err, "Failed to update queue status")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *QueueReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&queuev1.Queue{}).
		Complete(r)
}
