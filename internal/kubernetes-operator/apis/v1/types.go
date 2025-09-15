package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories=queue
// +kubebuilder:printcolumn:name="Priority",type="string",JSONPath=".spec.priority"
// +kubebuilder:printcolumn:name="Rate Limit",type="string",JSONPath=".spec.rateLimit"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Queue represents a queue configuration managed by the operator
type Queue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QueueSpec   `json:"spec,omitempty"`
	Status QueueStatus `json:"status,omitempty"`
}

// QueueSpec defines the desired state of Queue
type QueueSpec struct {
	// Name is the queue name
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=^[a-zA-Z0-9][a-zA-Z0-9-_]*[a-zA-Z0-9]$
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name"`

	// Priority defines the queue priority level
	// +kubebuilder:validation:Enum=critical;high;medium;low
	// +kubebuilder:default="medium"
	Priority string `json:"priority,omitempty"`

	// RateLimit configuration for the queue
	// +optional
	RateLimit *RateLimitSpec `json:"rateLimit,omitempty"`

	// DeadLetterQueue configuration
	// +optional
	DeadLetterQueue *DeadLetterQueueSpec `json:"deadLetterQueue,omitempty"`

	// Retention policy for completed jobs
	// +optional
	Retention *RetentionSpec `json:"retention,omitempty"`

	// Redis configuration override
	// +optional
	Redis *RedisSpec `json:"redis,omitempty"`
}

// RateLimitSpec defines rate limiting configuration
type RateLimitSpec struct {
	// RequestsPerSecond is the maximum requests per second
	// +kubebuilder:validation:Minimum=0.1
	// +kubebuilder:validation:Maximum=100000
	RequestsPerSecond float64 `json:"requestsPerSecond"`

	// BurstCapacity is the maximum burst size
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100000
	BurstCapacity int32 `json:"burstCapacity"`

	// Enabled controls whether rate limiting is active
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`
}

// DeadLetterQueueSpec defines dead letter queue configuration
type DeadLetterQueueSpec struct {
	// Enabled controls whether DLQ is active
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// MaxRetries before sending to DLQ
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=3
	MaxRetries int32 `json:"maxRetries,omitempty"`

	// RetryBackoff configuration
	// +optional
	RetryBackoff *RetryBackoffSpec `json:"retryBackoff,omitempty"`
}

// RetryBackoffSpec defines retry backoff configuration
type RetryBackoffSpec struct {
	// InitialDelay is the initial retry delay
	// +kubebuilder:default="1s"
	InitialDelay metav1.Duration `json:"initialDelay,omitempty"`

	// MaxDelay is the maximum retry delay
	// +kubebuilder:default="30s"
	MaxDelay metav1.Duration `json:"maxDelay,omitempty"`

	// Multiplier for exponential backoff
	// +kubebuilder:validation:Minimum=1.0
	// +kubebuilder:validation:Maximum=10.0
	// +kubebuilder:default=2.0
	Multiplier float64 `json:"multiplier,omitempty"`
}

// RetentionSpec defines job retention policy
type RetentionSpec struct {
	// CompletedJobs retention period
	// +kubebuilder:default="24h"
	CompletedJobs metav1.Duration `json:"completedJobs,omitempty"`

	// FailedJobs retention period
	// +kubebuilder:default="72h"
	FailedJobs metav1.Duration `json:"failedJobs,omitempty"`

	// MaxJobs is the maximum number of jobs to retain
	// +kubebuilder:validation:Minimum=100
	// +kubebuilder:validation:Maximum=1000000
	// +kubebuilder:default=10000
	MaxJobs int32 `json:"maxJobs,omitempty"`
}

// RedisSpec defines Redis connection configuration
type RedisSpec struct {
	// Addresses of Redis instances
	// +kubebuilder:validation:MinItems=1
	Addresses []string `json:"addresses"`

	// Database number
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=15
	// +kubebuilder:default=0
	Database int32 `json:"database,omitempty"`

	// Password secret reference
	// +optional
	PasswordSecret *corev1.SecretKeySelector `json:"passwordSecret,omitempty"`

	// TLS configuration
	// +optional
	TLS *RedisTLSSpec `json:"tls,omitempty"`
}

// RedisTLSSpec defines Redis TLS configuration
type RedisTLSSpec struct {
	// Enabled controls whether TLS is used
	Enabled bool `json:"enabled"`

	// CASecret reference for custom CA
	// +optional
	CASecret *corev1.SecretKeySelector `json:"caSecret,omitempty"`

	// InsecureSkipVerify disables certificate verification
	// +kubebuilder:default=false
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

// QueueStatus defines the observed state of Queue
type QueueStatus struct {
	// Phase represents the current lifecycle phase
	// +kubebuilder:validation:Enum=Pending;Active;Failed;Terminating
	Phase QueuePhase `json:"phase,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Metrics contain queue statistics
	// +optional
	Metrics *QueueMetrics `json:"metrics,omitempty"`

	// ObservedGeneration reflects the generation observed by the controller
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// QueuePhase represents the lifecycle phase of a Queue
type QueuePhase string

const (
	QueuePhasePending     QueuePhase = "Pending"
	QueuePhaseActive      QueuePhase = "Active"
	QueuePhaseFailed      QueuePhase = "Failed"
	QueuePhaseTerminating QueuePhase = "Terminating"
)

// QueueMetrics contains queue operational metrics
type QueueMetrics struct {
	// BacklogSize is the current number of pending jobs
	BacklogSize int64 `json:"backlogSize,omitempty"`

	// ProcessingRate is jobs processed per second
	ProcessingRate float64 `json:"processingRate,omitempty"`

	// ErrorRate is the percentage of failed jobs
	ErrorRate float64 `json:"errorRate,omitempty"`

	// AverageLatency in milliseconds
	AverageLatency float64 `json:"averageLatency,omitempty"`

	// LastUpdated timestamp
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:object:root=true

// QueueList contains a list of Queue
type QueueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Queue `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
// +kubebuilder:resource:scope=Namespaced,categories=queue
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas"
// +kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.readyReplicas"
// +kubebuilder:printcolumn:name="Queue",type="string",JSONPath=".spec.queueSelector.queue"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// WorkerPool represents a pool of workers managed by the operator
type WorkerPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkerPoolSpec   `json:"spec,omitempty"`
	Status WorkerPoolStatus `json:"status,omitempty"`
}

// WorkerPoolSpec defines the desired state of WorkerPool
type WorkerPoolSpec struct {
	// Replicas is the desired number of worker replicas
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1000
	// +kubebuilder:default=1
	Replicas *int32 `json:"replicas,omitempty"`

	// QueueSelector defines which queues this pool processes
	QueueSelector QueueSelector `json:"queueSelector"`

	// Template defines the worker pod template
	Template WorkerPodTemplate `json:"template"`

	// AutoScaling configuration
	// +optional
	AutoScaling *AutoScalingSpec `json:"autoScaling,omitempty"`

	// UpdateStrategy for rolling updates
	// +optional
	UpdateStrategy *UpdateStrategySpec `json:"updateStrategy,omitempty"`

	// DrainPolicy for graceful shutdown
	// +optional
	DrainPolicy *DrainPolicySpec `json:"drainPolicy,omitempty"`
}

// QueueSelector defines queue selection criteria
type QueueSelector struct {
	// Queue name to process
	// +optional
	Queue string `json:"queue,omitempty"`

	// Priority levels to process
	// +optional
	Priorities []string `json:"priorities,omitempty"`

	// MatchLabels selector
	// +optional
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

// WorkerPodTemplate defines the worker pod specification
type WorkerPodTemplate struct {
	// Metadata for worker pods
	// +optional
	Metadata *metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the worker pod specification
	Spec WorkerPodSpec `json:"spec"`
}

// WorkerPodSpec defines worker-specific pod configuration
type WorkerPodSpec struct {
	// Container image for the worker
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// ImagePullPolicy for the worker image
	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	// +kubebuilder:default="IfNotPresent"
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Resources for the worker container
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Environment variables
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// EnvFrom sources
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`

	// Concurrency is the number of concurrent jobs per worker
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1000
	// +kubebuilder:default=10
	Concurrency int32 `json:"concurrency,omitempty"`

	// MaxInFlight jobs before backpressure
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10000
	// +kubebuilder:default=100
	MaxInFlight int32 `json:"maxInFlight,omitempty"`

	// ServiceAccount for the worker pods
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// SecurityContext for the worker pods
	// +optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`

	// NodeSelector for worker pod scheduling
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations for worker pod scheduling
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Affinity for worker pod scheduling
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
}

// AutoScalingSpec defines autoscaling configuration
type AutoScalingSpec struct {
	// MinReplicas is the minimum number of replicas
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1000
	// +kubebuilder:default=1
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	// MaxReplicas is the maximum number of replicas
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1000
	// +kubebuilder:default=10
	MaxReplicas int32 `json:"maxReplicas"`

	// TargetBacklogPerWorker for scaling decisions
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10000
	// +kubebuilder:default=100
	TargetBacklogPerWorker int32 `json:"targetBacklogPerWorker,omitempty"`

	// LatencySLO target for scaling
	// +optional
	LatencySLO *LatencySLOSpec `json:"latencySLO,omitempty"`

	// ScaleUpCooldown period
	// +kubebuilder:default="3m"
	ScaleUpCooldown metav1.Duration `json:"scaleUpCooldown,omitempty"`

	// ScaleDownCooldown period
	// +kubebuilder:default="5m"
	ScaleDownCooldown metav1.Duration `json:"scaleDownCooldown,omitempty"`
}

// LatencySLOSpec defines latency-based SLO for autoscaling
type LatencySLOSpec struct {
	// TargetPercentile (e.g., 0.95 for p95)
	// +kubebuilder:validation:Minimum=0.5
	// +kubebuilder:validation:Maximum=0.99
	// +kubebuilder:default=0.95
	TargetPercentile float64 `json:"targetPercentile,omitempty"`

	// TargetLatencyMs in milliseconds
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=60000
	// +kubebuilder:default=1000
	TargetLatencyMs int32 `json:"targetLatencyMs,omitempty"`
}

// UpdateStrategySpec defines rolling update strategy
type UpdateStrategySpec struct {
	// Type of update strategy
	// +kubebuilder:validation:Enum=RollingUpdate;Recreate
	// +kubebuilder:default="RollingUpdate"
	Type string `json:"type,omitempty"`

	// RollingUpdate configuration
	// +optional
	RollingUpdate *RollingUpdateSpec `json:"rollingUpdate,omitempty"`
}

// RollingUpdateSpec defines rolling update parameters
type RollingUpdateSpec struct {
	// MaxUnavailable during update
	// +kubebuilder:default="25%"
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`

	// MaxSurge during update
	// +kubebuilder:default="25%"
	MaxSurge *intstr.IntOrString `json:"maxSurge,omitempty"`
}

// DrainPolicySpec defines graceful shutdown behavior
type DrainPolicySpec struct {
	// GracePeriod for job completion
	// +kubebuilder:default="30s"
	GracePeriod metav1.Duration `json:"gracePeriod,omitempty"`

	// TimeoutPeriod for forced termination
	// +kubebuilder:default="60s"
	TimeoutPeriod metav1.Duration `json:"timeoutPeriod,omitempty"`

	// WaitForCompletion of in-flight jobs
	// +kubebuilder:default=true
	WaitForCompletion bool `json:"waitForCompletion,omitempty"`
}

// WorkerPoolStatus defines the observed state of WorkerPool
type WorkerPoolStatus struct {
	// Phase represents the current lifecycle phase
	// +kubebuilder:validation:Enum=Pending;Active;Scaling;Updating;Failed;Terminating
	Phase WorkerPoolPhase `json:"phase,omitempty"`

	// Replicas is the current number of replicas
	Replicas int32 `json:"replicas,omitempty"`

	// ReadyReplicas is the number of ready replicas
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// UpdatedReplicas is the number of updated replicas
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty"`

	// AvailableReplicas is the number of available replicas
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// AutoScaling status
	// +optional
	AutoScaling *AutoScalingStatus `json:"autoScaling,omitempty"`

	// ObservedGeneration reflects the generation observed by the controller
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// WorkerPoolPhase represents the lifecycle phase of a WorkerPool
type WorkerPoolPhase string

const (
	WorkerPoolPhasePending     WorkerPoolPhase = "Pending"
	WorkerPoolPhaseActive      WorkerPoolPhase = "Active"
	WorkerPoolPhaseScaling     WorkerPoolPhase = "Scaling"
	WorkerPoolPhaseUpdating    WorkerPoolPhase = "Updating"
	WorkerPoolPhaseFailed      WorkerPoolPhase = "Failed"
	WorkerPoolPhaseTerminating WorkerPoolPhase = "Terminating"
)

// AutoScalingStatus contains autoscaling status information
type AutoScalingStatus struct {
	// DesiredReplicas from autoscaling calculation
	DesiredReplicas int32 `json:"desiredReplicas,omitempty"`

	// LastScaleTime of the last scaling operation
	LastScaleTime metav1.Time `json:"lastScaleTime,omitempty"`

	// CurrentMetrics used for autoscaling decisions
	// +optional
	CurrentMetrics *AutoScalingMetrics `json:"currentMetrics,omitempty"`
}

// AutoScalingMetrics contains metrics used for autoscaling
type AutoScalingMetrics struct {
	// BacklogPerWorker current ratio
	BacklogPerWorker float64 `json:"backlogPerWorker,omitempty"`

	// CurrentLatencyP95 in milliseconds
	CurrentLatencyP95 float64 `json:"currentLatencyP95,omitempty"`

	// ProcessingRate jobs per second per worker
	ProcessingRate float64 `json:"processingRate,omitempty"`
}

// +kubebuilder:object:root=true

// WorkerPoolList contains a list of WorkerPool
type WorkerPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkerPool `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories=queue
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Policy represents global queue system policies
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicySpec   `json:"spec,omitempty"`
	Status PolicyStatus `json:"status,omitempty"`
}

// PolicySpec defines the desired state of Policy
type PolicySpec struct {
	// CircuitBreaker global configuration
	// +optional
	CircuitBreaker *CircuitBreakerSpec `json:"circuitBreaker,omitempty"`

	// RetryDefaults for all queues
	// +optional
	RetryDefaults *RetryBackoffSpec `json:"retryDefaults,omitempty"`

	// RateLimitDefaults for all queues
	// +optional
	RateLimitDefaults *RateLimitSpec `json:"rateLimitDefaults,omitempty"`

	// SecurityPolicy configuration
	// +optional
	SecurityPolicy *SecurityPolicySpec `json:"securityPolicy,omitempty"`
}

// CircuitBreakerSpec defines circuit breaker configuration
type CircuitBreakerSpec struct {
	// FailureThreshold before tripping
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=5
	FailureThreshold int32 `json:"failureThreshold,omitempty"`

	// RecoveryThreshold for closing
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=3
	RecoveryThreshold int32 `json:"recoveryThreshold,omitempty"`

	// Timeout before allowing test requests
	// +kubebuilder:default="60s"
	Timeout metav1.Duration `json:"timeout,omitempty"`
}

// SecurityPolicySpec defines security policies
type SecurityPolicySpec struct {
	// RequiredServiceAccount for workers
	// +optional
	RequiredServiceAccount string `json:"requiredServiceAccount,omitempty"`

	// PodSecurityStandards enforcement
	// +optional
	PodSecurityStandards *PodSecurityStandardsSpec `json:"podSecurityStandards,omitempty"`

	// NetworkPolicies to apply
	// +optional
	NetworkPolicies []string `json:"networkPolicies,omitempty"`
}

// PodSecurityStandardsSpec defines pod security standards
type PodSecurityStandardsSpec struct {
	// Enforce security standards
	// +kubebuilder:validation:Enum=privileged;baseline;restricted
	// +kubebuilder:default="baseline"
	Enforce string `json:"enforce,omitempty"`

	// AllowPrivilegeEscalation setting
	// +kubebuilder:default=false
	AllowPrivilegeEscalation bool `json:"allowPrivilegeEscalation,omitempty"`

	// RunAsNonRoot requirement
	// +kubebuilder:default=true
	RunAsNonRoot bool `json:"runAsNonRoot,omitempty"`
}

// PolicyStatus defines the observed state of Policy
type PolicyStatus struct {
	// Phase represents the current lifecycle phase
	// +kubebuilder:validation:Enum=Pending;Active;Failed
	Phase PolicyPhase `json:"phase,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration reflects the generation observed by the controller
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// PolicyPhase represents the lifecycle phase of a Policy
type PolicyPhase string

const (
	PolicyPhasePending PolicyPhase = "Pending"
	PolicyPhaseActive  PolicyPhase = "Active"
	PolicyPhaseFailed  PolicyPhase = "Failed"
)

// +kubebuilder:object:root=true

// PolicyList contains a list of Policy
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Policy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Queue{}, &QueueList{})
	SchemeBuilder.Register(&WorkerPool{}, &WorkerPoolList{})
	SchemeBuilder.Register(&Policy{}, &PolicyList{})
}