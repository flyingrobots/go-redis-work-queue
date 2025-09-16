package controllers

import (
	"context"
	"fmt"
	"math"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	queuev1 "github.com/flyingrobots/go-redis-work-queue/internal/kubernetes-operator/apis/v1"
)

// WorkerPoolReconciler reconciles a WorkerPool object
type WorkerPoolReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	AdminAPIClient AdminAPIClient
	MetricsClient  MetricsClient
}

// MetricsClient interface for gathering metrics for autoscaling
type MetricsClient interface {
	GetQueueBacklog(ctx context.Context, queueName string) (int64, error)
	GetQueueLatency(ctx context.Context, queueName string, percentile float64) (float64, error)
	GetWorkerMetrics(ctx context.Context, namespace, workerPoolName string) (*WorkerMetrics, error)
}

// WorkerMetrics represents worker performance metrics
type WorkerMetrics struct {
	ProcessingRate    float64   `json:"processingRate"`    // Jobs per second per worker
	AverageLatency    float64   `json:"averageLatency"`    // Average job processing latency
	ActiveWorkers     int32     `json:"activeWorkers"`     // Number of active workers
	LastUpdated       time.Time `json:"lastUpdated"`
}

// +kubebuilder:rbac:groups=queue.example.com,resources=workerpools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=queue.example.com,resources=workerpools/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=queue.example.com,resources=workerpools/finalizers,verbs=update
// +kubebuilder:rbac:groups=queue.example.com,resources=workerpools/scale,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *WorkerPoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the WorkerPool instance
	var workerPool queuev1.WorkerPool
	if err := r.Get(ctx, req.NamespacedName, &workerPool); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("WorkerPool resource not found, ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get WorkerPool")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if workerPool.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, &workerPool)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(&workerPool, WorkerPoolFinalizerName) {
		controllerutil.AddFinalizer(&workerPool, WorkerPoolFinalizerName)
		return ctrl.Result{}, r.Update(ctx, &workerPool)
	}

	// Reconcile the worker pool
	return r.reconcileWorkerPool(ctx, &workerPool)
}

const WorkerPoolFinalizerName = "workerpool.example.com/finalizer"

// handleDeletion handles worker pool deletion with proper draining
func (r *WorkerPoolReconciler) handleDeletion(ctx context.Context, workerPool *queuev1.WorkerPool) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(workerPool, WorkerPoolFinalizerName) {
		// Perform graceful shutdown with draining
		if err := r.drainWorkerPool(ctx, workerPool); err != nil {
			logger.Error(err, "Failed to drain worker pool", "workerPool", workerPool.Name)
			return ctrl.Result{RequeueAfter: 30 * time.Second}, err
		}

		// Delete the deployment
		deployment := &appsv1.Deployment{}
		deploymentName := types.NamespacedName{
			Namespace: workerPool.Namespace,
			Name:      r.getDeploymentName(workerPool),
		}

		if err := r.Get(ctx, deploymentName, deployment); err == nil {
			if err := r.Delete(ctx, deployment); err != nil {
				logger.Error(err, "Failed to delete deployment", "deployment", deploymentName.Name)
				return ctrl.Result{RequeueAfter: time.Minute}, err
			}
		}

		// Remove finalizer
		controllerutil.RemoveFinalizer(workerPool, WorkerPoolFinalizerName)
		if err := r.Update(ctx, workerPool); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// reconcileWorkerPool handles the main reconciliation logic
func (r *WorkerPoolReconciler) reconcileWorkerPool(ctx context.Context, workerPool *queuev1.WorkerPool) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Calculate desired replicas (including autoscaling)
	desiredReplicas, err := r.calculateDesiredReplicas(ctx, workerPool)
	if err != nil {
		logger.Error(err, "Failed to calculate desired replicas", "workerPool", workerPool.Name)
		r.updateWorkerPoolStatus(ctx, workerPool, queuev1.WorkerPoolPhaseFailed, fmt.Sprintf("Failed to calculate replicas: %v", err))
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}

	// Create or update deployment
	deployment, err := r.reconcileDeployment(ctx, workerPool, desiredReplicas)
	if err != nil {
		logger.Error(err, "Failed to reconcile deployment", "workerPool", workerPool.Name)
		r.updateWorkerPoolStatus(ctx, workerPool, queuev1.WorkerPoolPhaseFailed, fmt.Sprintf("Failed to reconcile deployment: %v", err))
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}

	// Update status based on deployment status
	r.updateWorkerPoolStatusFromDeployment(ctx, workerPool, deployment, desiredReplicas)

	// Requeue for periodic reconciliation and autoscaling
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// calculateDesiredReplicas determines the desired number of replicas including autoscaling
func (r *WorkerPoolReconciler) calculateDesiredReplicas(ctx context.Context, workerPool *queuev1.WorkerPool) (int32, error) {
	// Start with spec replicas or default
	desiredReplicas := int32(1)
	if workerPool.Spec.Replicas != nil {
		desiredReplicas = *workerPool.Spec.Replicas
	}

	// Apply autoscaling if enabled
	if workerPool.Spec.AutoScaling != nil {
		scaledReplicas, err := r.calculateAutoscaledReplicas(ctx, workerPool)
		if err != nil {
			return desiredReplicas, err
		}
		desiredReplicas = scaledReplicas
	}

	return desiredReplicas, nil
}

// calculateAutoscaledReplicas implements autoscaling logic
func (r *WorkerPoolReconciler) calculateAutoscaledReplicas(ctx context.Context, workerPool *queuev1.WorkerPool) (int32, error) {
	autoScaling := workerPool.Spec.AutoScaling
	logger := log.FromContext(ctx)

	// Get current replicas
	currentReplicas := workerPool.Status.Replicas
	if currentReplicas == 0 {
		currentReplicas = 1 // Default to 1 if not set
	}

	// Check cooldown periods
	if workerPool.Status.AutoScaling != nil && !workerPool.Status.AutoScaling.LastScaleTime.IsZero() {
		timeSinceLastScale := time.Since(workerPool.Status.AutoScaling.LastScaleTime.Time)

		// Don't scale up too frequently
		if currentReplicas < autoScaling.MaxReplicas && timeSinceLastScale < autoScaling.ScaleUpCooldown.Duration {
			logger.Info("Scale up cooldown in effect", "timeSinceLastScale", timeSinceLastScale, "cooldown", autoScaling.ScaleUpCooldown.Duration)
			return currentReplicas, nil
		}

		// Don't scale down too frequently
		if currentReplicas > *autoScaling.MinReplicas && timeSinceLastScale < autoScaling.ScaleDownCooldown.Duration {
			logger.Info("Scale down cooldown in effect", "timeSinceLastScale", timeSinceLastScale, "cooldown", autoScaling.ScaleDownCooldown.Duration)
			return currentReplicas, nil
		}
	}

	// Get queue metrics for scaling decision
	queueName := r.getQueueNameFromSelector(workerPool.Spec.QueueSelector)
	if queueName == "" {
		return currentReplicas, fmt.Errorf("cannot determine queue name from selector")
	}

	backlog, err := r.MetricsClient.GetQueueBacklog(ctx, queueName)
	if err != nil {
		logger.Error(err, "Failed to get queue backlog", "queue", queueName)
		return currentReplicas, err
	}

	// Calculate desired replicas based on backlog
	var desiredReplicas int32

	if autoScaling.TargetBacklogPerWorker > 0 {
		// Scale based on backlog per worker
		desiredFromBacklog := int32(math.Ceil(float64(backlog) / float64(autoScaling.TargetBacklogPerWorker)))
		desiredReplicas = desiredFromBacklog
	} else {
		desiredReplicas = currentReplicas
	}

	// Check latency SLO if configured
	if autoScaling.LatencySLO != nil {
		currentLatency, err := r.MetricsClient.GetQueueLatency(ctx, queueName, autoScaling.LatencySLO.TargetPercentile)
		if err != nil {
			logger.Error(err, "Failed to get queue latency", "queue", queueName)
		} else {
			targetLatency := float64(autoScaling.LatencySLO.TargetLatencyMs)
			if currentLatency > targetLatency {
				// Scale up if latency is above target
				latencyScaleFactor := currentLatency / targetLatency
				latencyDesiredReplicas := int32(math.Ceil(float64(currentReplicas) * latencyScaleFactor))
				if latencyDesiredReplicas > desiredReplicas {
					desiredReplicas = latencyDesiredReplicas
				}
			}
		}
	}

	// Apply min/max constraints
	minReplicas := int32(1)
	if autoScaling.MinReplicas != nil {
		minReplicas = *autoScaling.MinReplicas
	}

	if desiredReplicas < minReplicas {
		desiredReplicas = minReplicas
	}
	if desiredReplicas > autoScaling.MaxReplicas {
		desiredReplicas = autoScaling.MaxReplicas
	}

	logger.Info("Autoscaling calculation",
		"queue", queueName,
		"backlog", backlog,
		"currentReplicas", currentReplicas,
		"desiredReplicas", desiredReplicas,
		"targetBacklogPerWorker", autoScaling.TargetBacklogPerWorker)

	return desiredReplicas, nil
}

// getQueueNameFromSelector extracts queue name from queue selector
func (r *WorkerPoolReconciler) getQueueNameFromSelector(selector queuev1.QueueSelector) string {
	if selector.Queue != "" {
		return selector.Queue
	}
	// Could implement more complex selector logic here
	return ""
}

// reconcileDeployment creates or updates the worker deployment
func (r *WorkerPoolReconciler) reconcileDeployment(ctx context.Context, workerPool *queuev1.WorkerPool, desiredReplicas int32) (*appsv1.Deployment, error) {
	deploymentName := r.getDeploymentName(workerPool)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: workerPool.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "worker",
				"app.kubernetes.io/instance":   workerPool.Name,
				"app.kubernetes.io/component":  "worker",
				"app.kubernetes.io/managed-by": "queue-operator",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &desiredReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":     "worker",
					"app.kubernetes.io/instance": workerPool.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":     "worker",
						"app.kubernetes.io/instance": workerPool.Name,
						"app.kubernetes.io/component": "worker",
					},
				},
				Spec: r.buildPodSpec(workerPool),
			},
			Strategy: r.buildUpdateStrategy(workerPool),
		},
	}

	// Apply template metadata if provided
	if workerPool.Spec.Template.Metadata != nil {
		for k, v := range workerPool.Spec.Template.Metadata.Labels {
			deployment.Spec.Template.ObjectMeta.Labels[k] = v
		}
		for k, v := range workerPool.Spec.Template.Metadata.Annotations {
			if deployment.Spec.Template.ObjectMeta.Annotations == nil {
				deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
			}
			deployment.Spec.Template.ObjectMeta.Annotations[k] = v
		}
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(workerPool, deployment, r.Scheme); err != nil {
		return nil, err
	}

	// Create or update deployment
	existing := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: workerPool.Namespace}, existing)
	if errors.IsNotFound(err) {
		// Create new deployment
		if err := r.Create(ctx, deployment); err != nil {
			return nil, err
		}
		return deployment, nil
	} else if err != nil {
		return nil, err
	}

	// Update existing deployment
	existing.Spec = deployment.Spec
	if err := r.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// buildPodSpec creates the pod specification for workers
func (r *WorkerPoolReconciler) buildPodSpec(workerPool *queuev1.WorkerPool) corev1.PodSpec {
	workerSpec := workerPool.Spec.Template.Spec

	// Build environment variables
	env := append([]corev1.EnvVar{}, workerSpec.Env...)
	env = append(env,
		corev1.EnvVar{
			Name:  "WORKER_CONCURRENCY",
			Value: fmt.Sprintf("%d", workerSpec.Concurrency),
		},
		corev1.EnvVar{
			Name:  "WORKER_MAX_IN_FLIGHT",
			Value: fmt.Sprintf("%d", workerSpec.MaxInFlight),
		},
		corev1.EnvVar{
			Name:  "WORKER_QUEUE",
			Value: workerPool.Spec.QueueSelector.Queue,
		},
	)

	// Add priorities if specified
	if len(workerPool.Spec.QueueSelector.Priorities) > 0 {
		prioritiesStr := ""
		for i, priority := range workerPool.Spec.QueueSelector.Priorities {
			if i > 0 {
				prioritiesStr += ","
			}
			prioritiesStr += priority
		}
		env = append(env, corev1.EnvVar{
			Name:  "WORKER_PRIORITIES",
			Value: prioritiesStr,
		})
	}

	// Build container
	container := corev1.Container{
		Name:            "worker",
		Image:           workerSpec.Image,
		ImagePullPolicy: workerSpec.ImagePullPolicy,
		Env:             env,
		EnvFrom:         workerSpec.EnvFrom,
		Resources:       workerSpec.Resources,
		ReadinessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/health/ready",
					Port: intstr.FromInt(8080),
				},
			},
			InitialDelaySeconds: 10,
			PeriodSeconds:       10,
		},
		LivenessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/health/live",
					Port: intstr.FromInt(8080),
				},
			},
			InitialDelaySeconds: 30,
			PeriodSeconds:       30,
		},
	}

	// Build pod spec
	podSpec := corev1.PodSpec{
		ServiceAccountName: workerSpec.ServiceAccountName,
		SecurityContext:    workerSpec.SecurityContext,
		NodeSelector:       workerSpec.NodeSelector,
		Tolerations:        workerSpec.Tolerations,
		Affinity:           workerSpec.Affinity,
		Containers:         []corev1.Container{container},
		RestartPolicy:      corev1.RestartPolicyAlways,
	}

	// Add graceful shutdown configuration
	if workerPool.Spec.DrainPolicy != nil {
		gracePeriod := int64(workerPool.Spec.DrainPolicy.GracePeriod.Duration.Seconds())
		podSpec.TerminationGracePeriodSeconds = &gracePeriod

		// Add preStop hook for graceful draining
		container.Lifecycle = &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/drain",
					Port: intstr.FromInt(8080),
				},
			},
		}
	}

	return podSpec
}

// buildUpdateStrategy creates the deployment update strategy
func (r *WorkerPoolReconciler) buildUpdateStrategy(workerPool *queuev1.WorkerPool) appsv1.DeploymentStrategy {
	strategy := appsv1.DeploymentStrategy{
		Type: appsv1.RollingUpdateDeploymentStrategyType,
	}

	if workerPool.Spec.UpdateStrategy != nil {
		if workerPool.Spec.UpdateStrategy.Type == "Recreate" {
			strategy.Type = appsv1.RecreateDeploymentStrategyType
		} else if workerPool.Spec.UpdateStrategy.RollingUpdate != nil {
			strategy.RollingUpdate = &appsv1.RollingUpdateDeployment{
				MaxUnavailable: workerPool.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable,
				MaxSurge:       workerPool.Spec.UpdateStrategy.RollingUpdate.MaxSurge,
			}
		}
	}

	// Default rolling update configuration
	if strategy.RollingUpdate == nil {
		strategy.RollingUpdate = &appsv1.RollingUpdateDeployment{
			MaxUnavailable: &intstr.IntOrString{Type: intstr.String, StrVal: "25%"},
			MaxSurge:       &intstr.IntOrString{Type: intstr.String, StrVal: "25%"},
		}
	}

	return strategy
}

// drainWorkerPool performs graceful shutdown of worker pods
func (r *WorkerPoolReconciler) drainWorkerPool(ctx context.Context, workerPool *queuev1.WorkerPool) error {
	logger := log.FromContext(ctx)

	// Get current pods
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(workerPool.Namespace),
		client.MatchingLabels{
			"app.kubernetes.io/name":     "worker",
			"app.kubernetes.io/instance": workerPool.Name,
		},
	}

	if err := r.List(ctx, podList, listOpts...); err != nil {
		return err
	}

	// If no drain policy, skip draining
	if workerPool.Spec.DrainPolicy == nil || !workerPool.Spec.DrainPolicy.WaitForCompletion {
		return nil
	}

	// Call drain endpoint on each pod
	for _, pod := range podList.Items {
		if pod.Status.Phase == corev1.PodRunning {
			logger.Info("Draining worker pod", "pod", pod.Name)
			// In a real implementation, you would make an HTTP call to the pod's drain endpoint
			// For now, we'll just log the action
		}
	}

	// Wait for drain completion (simplified implementation)
	timeout := workerPool.Spec.DrainPolicy.TimeoutPeriod.Duration
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	logger.Info("Waiting for worker drain completion", "timeout", timeout)
	time.Sleep(timeout) // Simplified - in reality, you'd poll for completion

	return nil
}

// getDeploymentName generates the deployment name for a worker pool
func (r *WorkerPoolReconciler) getDeploymentName(workerPool *queuev1.WorkerPool) string {
	return fmt.Sprintf("%s-worker", workerPool.Name)
}

// updateWorkerPoolStatus updates the worker pool status
func (r *WorkerPoolReconciler) updateWorkerPoolStatus(ctx context.Context, workerPool *queuev1.WorkerPool, phase queuev1.WorkerPoolPhase, message string) {
	logger := log.FromContext(ctx)

	workerPool.Status.Phase = phase
	workerPool.Status.ObservedGeneration = workerPool.Generation

	// Update conditions
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "WorkerPoolReady",
		Message:            message,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}

	if phase == queuev1.WorkerPoolPhaseFailed {
		condition.Status = metav1.ConditionFalse
		condition.Reason = "WorkerPoolFailed"
	}

	// Update or add condition
	found := false
	for i, cond := range workerPool.Status.Conditions {
		if cond.Type == condition.Type {
			workerPool.Status.Conditions[i] = condition
			found = true
			break
		}
	}
	if !found {
		workerPool.Status.Conditions = append(workerPool.Status.Conditions, condition)
	}

	// Update status
	if err := r.Status().Update(ctx, workerPool); err != nil {
		logger.Error(err, "Failed to update worker pool status")
	}
}

// updateWorkerPoolStatusFromDeployment updates status based on deployment status
func (r *WorkerPoolReconciler) updateWorkerPoolStatusFromDeployment(ctx context.Context, workerPool *queuev1.WorkerPool, deployment *appsv1.Deployment, desiredReplicas int32) {
	// Update replica counts
	workerPool.Status.Replicas = deployment.Status.Replicas
	workerPool.Status.ReadyReplicas = deployment.Status.ReadyReplicas
	workerPool.Status.UpdatedReplicas = deployment.Status.UpdatedReplicas
	workerPool.Status.AvailableReplicas = deployment.Status.AvailableReplicas

	// Update autoscaling status
	if workerPool.Spec.AutoScaling != nil {
		if workerPool.Status.AutoScaling == nil {
			workerPool.Status.AutoScaling = &queuev1.AutoScalingStatus{}
		}

		workerPool.Status.AutoScaling.DesiredReplicas = desiredReplicas

		// Update last scale time if replicas changed
		if workerPool.Status.AutoScaling.DesiredReplicas != deployment.Status.Replicas {
			workerPool.Status.AutoScaling.LastScaleTime = metav1.NewTime(time.Now())
		}
	}

	// Determine phase
	phase := queuev1.WorkerPoolPhaseActive
	message := "Worker pool is active"

	if deployment.Status.Replicas != desiredReplicas {
		phase = queuev1.WorkerPoolPhaseScaling
		message = fmt.Sprintf("Scaling to %d replicas", desiredReplicas)
	} else if deployment.Status.UpdatedReplicas < deployment.Status.Replicas {
		phase = queuev1.WorkerPoolPhaseUpdating
		message = "Rolling update in progress"
	} else if deployment.Status.ReadyReplicas < deployment.Status.Replicas {
		phase = queuev1.WorkerPoolPhaseActive
		message = "Waiting for pods to become ready"
	}

	r.updateWorkerPoolStatus(ctx, workerPool, phase, message)
}

// SetupWithManager sets up the controller with the Manager.
func (r *WorkerPoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&queuev1.WorkerPool{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
