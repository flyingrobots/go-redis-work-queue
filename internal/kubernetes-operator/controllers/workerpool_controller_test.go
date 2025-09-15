package controllers

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	queuev1 "github.com/example/go-redis-work-queue/internal/kubernetes-operator/apis/v1"
)

// MockMetricsClient for testing
type MockMetricsClient struct {
	backlogs map[string]int64
	latencies map[string]float64
	workerMetrics map[string]*WorkerMetrics
	errors map[string]error
}

func NewMockMetricsClient() *MockMetricsClient {
	return &MockMetricsClient{
		backlogs: make(map[string]int64),
		latencies: make(map[string]float64),
		workerMetrics: make(map[string]*WorkerMetrics),
		errors: make(map[string]error),
	}
}

func (m *MockMetricsClient) GetQueueBacklog(ctx context.Context, queueName string) (int64, error) {
	if err, exists := m.errors["backlog"]; exists {
		return 0, err
	}
	if backlog, exists := m.backlogs[queueName]; exists {
		return backlog, nil
	}
	return 0, nil
}

func (m *MockMetricsClient) GetQueueLatency(ctx context.Context, queueName string, percentile float64) (float64, error) {
	if err, exists := m.errors["latency"]; exists {
		return 0, err
	}
	if latency, exists := m.latencies[queueName]; exists {
		return latency, nil
	}
	return 100.0, nil // Default latency
}

func (m *MockMetricsClient) GetWorkerMetrics(ctx context.Context, namespace, workerPoolName string) (*WorkerMetrics, error) {
	if err, exists := m.errors["worker"]; exists {
		return nil, err
	}
	key := namespace + "/" + workerPoolName
	if metrics, exists := m.workerMetrics[key]; exists {
		return metrics, nil
	}
	return &WorkerMetrics{
		ProcessingRate: 10.0,
		AverageLatency: 100.0,
		ActiveWorkers: 1,
		LastUpdated: time.Now(),
	}, nil
}

func (m *MockMetricsClient) SetBacklog(queueName string, backlog int64) {
	m.backlogs[queueName] = backlog
}

func (m *MockMetricsClient) SetLatency(queueName string, latency float64) {
	m.latencies[queueName] = latency
}

func (m *MockMetricsClient) SetError(operation string, err error) {
	m.errors[operation] = err
}

func TestWorkerPoolController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WorkerPool Controller Suite")
}

var _ = Describe("WorkerPoolController", func() {
	var (
		ctx           context.Context
		k8sClient     client.Client
		reconciler    *WorkerPoolReconciler
		mockAdminAPI  *MockAdminAPIClient
		mockMetrics   *MockMetricsClient
		scheme        *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(queuev1.AddToScheme(scheme)).To(Succeed())
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(appsv1.AddToScheme(scheme)).To(Succeed())

		k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		mockAdminAPI = NewMockAdminAPIClient()
		mockMetrics = NewMockMetricsClient()

		reconciler = &WorkerPoolReconciler{
			Client:         k8sClient,
			Scheme:         scheme,
			AdminAPIClient: mockAdminAPI,
			MetricsClient:  mockMetrics,
		}
	})

	Describe("Reconciling a WorkerPool", func() {
		var workerPool *queuev1.WorkerPool

		BeforeEach(func() {
			replicas := int32(2)
			workerPool = &queuev1.WorkerPool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-worker-pool",
					Namespace: "default",
				},
				Spec: queuev1.WorkerPoolSpec{
					Replicas: &replicas,
					QueueSelector: queuev1.QueueSelector{
						Queue: "test-queue",
					},
					Template: queuev1.WorkerPodTemplate{
						Spec: queuev1.WorkerPodSpec{
							Image:       "worker:latest",
							Concurrency: 10,
							MaxInFlight: 100,
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
						},
					},
				},
			}
		})

		Context("When creating a new WorkerPool", func() {
			It("Should create a deployment successfully", func() {
				Expect(k8sClient.Create(ctx, workerPool)).To(Succeed())

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      workerPool.Name,
						Namespace: workerPool.Namespace,
					},
				}

				result, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RequeueAfter).To(Equal(30 * time.Second))

				// Verify deployment was created
				deployment := &appsv1.Deployment{}
				deploymentName := types.NamespacedName{
					Namespace: workerPool.Namespace,
					Name:      "test-worker-pool-worker",
				}
				Expect(k8sClient.Get(ctx, deploymentName, deployment)).To(Succeed())
				Expect(*deployment.Spec.Replicas).To(Equal(int32(2)))

				// Verify container configuration
				container := deployment.Spec.Template.Spec.Containers[0]
				Expect(container.Image).To(Equal("worker:latest"))
				Expect(container.Env).To(ContainElement(corev1.EnvVar{Name: "WORKER_CONCURRENCY", Value: "10"}))
				Expect(container.Env).To(ContainElement(corev1.EnvVar{Name: "WORKER_MAX_IN_FLIGHT", Value: "100"}))
				Expect(container.Env).To(ContainElement(corev1.EnvVar{Name: "WORKER_QUEUE", Value: "test-queue"}))
			})

			It("Should handle multiple priorities", func() {
				workerPool.Spec.QueueSelector.Priorities = []string{"high", "medium"}

				Expect(k8sClient.Create(ctx, workerPool)).To(Succeed())

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      workerPool.Name,
						Namespace: workerPool.Namespace,
					},
				}

				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				// Verify deployment includes priorities
				deployment := &appsv1.Deployment{}
				deploymentName := types.NamespacedName{
					Namespace: workerPool.Namespace,
					Name:      "test-worker-pool-worker",
				}
				Expect(k8sClient.Get(ctx, deploymentName, deployment)).To(Succeed())

				container := deployment.Spec.Template.Spec.Containers[0]
				Expect(container.Env).To(ContainElement(corev1.EnvVar{Name: "WORKER_PRIORITIES", Value: "high,medium"}))
			})
		})

		Context("When autoscaling is enabled", func() {
			BeforeEach(func() {
				minReplicas := int32(1)
				workerPool.Spec.AutoScaling = &queuev1.AutoScalingSpec{
					MinReplicas:            &minReplicas,
					MaxReplicas:            10,
					TargetBacklogPerWorker: 50,
					ScaleUpCooldown:        metav1.Duration{Duration: 3 * time.Minute},
					ScaleDownCooldown:      metav1.Duration{Duration: 5 * time.Minute},
				}
			})

			It("Should scale up based on backlog", func() {
				// Set high backlog
				mockMetrics.SetBacklog("test-queue", 500) // Should trigger scale to 10 workers (500/50)

				Expect(k8sClient.Create(ctx, workerPool)).To(Succeed())

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      workerPool.Name,
						Namespace: workerPool.Namespace,
					},
				}

				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				// Verify deployment scaled up
				deployment := &appsv1.Deployment{}
				deploymentName := types.NamespacedName{
					Namespace: workerPool.Namespace,
					Name:      "test-worker-pool-worker",
				}
				Expect(k8sClient.Get(ctx, deploymentName, deployment)).To(Succeed())
				Expect(*deployment.Spec.Replicas).To(Equal(int32(10))) // Capped at max
			})

			It("Should scale down when backlog is low", func() {
				// Set low backlog
				mockMetrics.SetBacklog("test-queue", 10) // Should trigger scale to 1 worker (10/50 = 0.2, rounded up to 1)

				Expect(k8sClient.Create(ctx, workerPool)).To(Succeed())

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      workerPool.Name,
						Namespace: workerPool.Namespace,
					},
				}

				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				// Verify deployment scaled down
				deployment := &appsv1.Deployment{}
				deploymentName := types.NamespacedName{
					Namespace: workerPool.Namespace,
					Name:      "test-worker-pool-worker",
				}
				Expect(k8sClient.Get(ctx, deploymentName, deployment)).To(Succeed())
				Expect(*deployment.Spec.Replicas).To(Equal(int32(1))) // Min replicas
			})

			It("Should scale based on latency SLO", func() {
				workerPool.Spec.AutoScaling.LatencySLO = &queuev1.LatencySLOSpec{
					TargetPercentile: 0.95,
					TargetLatencyMs:  500,
				}

				// Set high latency (double the target)
				mockMetrics.SetLatency("test-queue", 1000.0)
				mockMetrics.SetBacklog("test-queue", 50) // Baseline: 1 worker

				Expect(k8sClient.Create(ctx, workerPool)).To(Succeed())

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      workerPool.Name,
						Namespace: workerPool.Namespace,
					},
				}

				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				// Should scale up due to latency (factor of 2)
				deployment := &appsv1.Deployment{}
				deploymentName := types.NamespacedName{
					Namespace: workerPool.Namespace,
					Name:      "test-worker-pool-worker",
				}
				Expect(k8sClient.Get(ctx, deploymentName, deployment)).To(Succeed())
				Expect(*deployment.Spec.Replicas).To(BeNumerically(">=", 2))
			})
		})

		Context("When handling rolling updates", func() {
			It("Should configure rolling update strategy", func() {
				workerPool.Spec.UpdateStrategy = &queuev1.UpdateStrategySpec{
					Type: "RollingUpdate",
					RollingUpdate: &queuev1.RollingUpdateSpec{
						MaxUnavailable: &intstr.IntOrString{Type: intstr.String, StrVal: "50%"},
						MaxSurge:       &intstr.IntOrString{Type: intstr.String, StrVal: "50%"},
					},
				}

				Expect(k8sClient.Create(ctx, workerPool)).To(Succeed())

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      workerPool.Name,
						Namespace: workerPool.Namespace,
					},
				}

				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				// Verify deployment update strategy
				deployment := &appsv1.Deployment{}
				deploymentName := types.NamespacedName{
					Namespace: workerPool.Namespace,
					Name:      "test-worker-pool-worker",
				}
				Expect(k8sClient.Get(ctx, deploymentName, deployment)).To(Succeed())
				Expect(deployment.Spec.Strategy.Type).To(Equal(appsv1.RollingUpdateDeploymentStrategyType))
				Expect(deployment.Spec.Strategy.RollingUpdate.MaxUnavailable.StrVal).To(Equal("50%"))
				Expect(deployment.Spec.Strategy.RollingUpdate.MaxSurge.StrVal).To(Equal("50%"))
			})
		})

		Context("When handling graceful shutdown", func() {
			It("Should configure drain policy", func() {
				workerPool.Spec.DrainPolicy = &queuev1.DrainPolicySpec{
					GracePeriod:       metav1.Duration{Duration: 45 * time.Second},
					TimeoutPeriod:     metav1.Duration{Duration: 120 * time.Second},
					WaitForCompletion: true,
				}

				Expect(k8sClient.Create(ctx, workerPool)).To(Succeed())

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      workerPool.Name,
						Namespace: workerPool.Namespace,
					},
				}

				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				// Verify deployment includes drain configuration
				deployment := &appsv1.Deployment{}
				deploymentName := types.NamespacedName{
					Namespace: workerPool.Namespace,
					Name:      "test-worker-pool-worker",
				}
				Expect(k8sClient.Get(ctx, deploymentName, deployment)).To(Succeed())

				// Check termination grace period
				Expect(*deployment.Spec.Template.Spec.TerminationGracePeriodSeconds).To(Equal(int64(45)))

				// Check preStop hook
				container := deployment.Spec.Template.Spec.Containers[0]
				Expect(container.Lifecycle).NotTo(BeNil())
				Expect(container.Lifecycle.PreStop).NotTo(BeNil())
				Expect(container.Lifecycle.PreStop.HTTPGet.Path).To(Equal("/drain"))
			})
		})

		Context("When deleting a WorkerPool", func() {
			It("Should perform cleanup and remove finalizer", func() {
				Expect(k8sClient.Create(ctx, workerPool)).To(Succeed())

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      workerPool.Name,
						Namespace: workerPool.Namespace,
					},
				}

				// First reconcile to create and add finalizer
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				// Verify finalizer was added
				var workerPoolWithFinalizer queuev1.WorkerPool
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: workerPool.Name, Namespace: workerPool.Namespace}, &workerPoolWithFinalizer)).To(Succeed())
				Expect(workerPoolWithFinalizer.Finalizers).To(ContainElement(WorkerPoolFinalizerName))

				// Delete the worker pool
				Expect(k8sClient.Delete(ctx, &workerPoolWithFinalizer)).To(Succeed())

				// Reconcile deletion
				_, err = reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				// Verify deployment was deleted
				deployment := &appsv1.Deployment{}
				deploymentName := types.NamespacedName{
					Namespace: workerPool.Namespace,
					Name:      "test-worker-pool-worker",
				}
				err = k8sClient.Get(ctx, deploymentName, deployment)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Autoscaling calculations", func() {
		var workerPool *queuev1.WorkerPool

		BeforeEach(func() {
			minReplicas := int32(1)
			workerPool = &queuev1.WorkerPool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-worker-pool",
					Namespace: "default",
				},
				Spec: queuev1.WorkerPoolSpec{
					QueueSelector: queuev1.QueueSelector{
						Queue: "test-queue",
					},
					AutoScaling: &queuev1.AutoScalingSpec{
						MinReplicas:            &minReplicas,
						MaxReplicas:            5,
						TargetBacklogPerWorker: 100,
					},
				},
				Status: queuev1.WorkerPoolStatus{
					Replicas: 2,
				},
			}
		})

		It("Should calculate replicas based on backlog", func() {
			mockMetrics.SetBacklog("test-queue", 350) // Should scale to 4 workers (350/100 = 3.5, rounded up)

			replicas, err := reconciler.calculateAutoscaledReplicas(ctx, workerPool)
			Expect(err).NotTo(HaveOccurred())
			Expect(replicas).To(Equal(int32(4)))
		})

		It("Should respect min replicas", func() {
			mockMetrics.SetBacklog("test-queue", 10) // Should scale to 1 worker (10/100 = 0.1, but min is 1)

			replicas, err := reconciler.calculateAutoscaledReplicas(ctx, workerPool)
			Expect(err).NotTo(HaveOccurred())
			Expect(replicas).To(Equal(int32(1)))
		})

		It("Should respect max replicas", func() {
			mockMetrics.SetBacklog("test-queue", 1000) // Should scale to 5 workers (max), not 10

			replicas, err := reconciler.calculateAutoscaledReplicas(ctx, workerPool)
			Expect(err).NotTo(HaveOccurred())
			Expect(replicas).To(Equal(int32(5)))
		})

		It("Should handle metrics errors gracefully", func() {
			mockMetrics.SetError("backlog", fmt.Errorf("metrics unavailable"))

			replicas, err := reconciler.calculateAutoscaledReplicas(ctx, workerPool)
			Expect(err).To(HaveOccurred())
			Expect(replicas).To(Equal(int32(2))) // Should return current replicas
		})
	})

	Describe("Pod spec building", func() {
		var workerPool *queuev1.WorkerPool

		BeforeEach(func() {
			workerPool = &queuev1.WorkerPool{
				Spec: queuev1.WorkerPoolSpec{
					QueueSelector: queuev1.QueueSelector{
						Queue: "test-queue",
						Priorities: []string{"high", "medium"},
					},
					Template: queuev1.WorkerPodTemplate{
						Spec: queuev1.WorkerPodSpec{
							Image:           "worker:v1.0",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Concurrency:     20,
							MaxInFlight:     200,
							ServiceAccountName: "worker-sa",
							Env: []corev1.EnvVar{
								{Name: "CUSTOM_VAR", Value: "custom-value"},
							},
							NodeSelector: map[string]string{
								"node-type": "worker",
							},
						},
					},
				},
			}
		})

		It("Should build pod spec correctly", func() {
			podSpec := reconciler.buildPodSpec(workerPool)

			Expect(podSpec.ServiceAccountName).To(Equal("worker-sa"))
			Expect(podSpec.NodeSelector).To(HaveKeyWithValue("node-type", "worker"))
			Expect(podSpec.Containers).To(HaveLen(1))

			container := podSpec.Containers[0]
			Expect(container.Name).To(Equal("worker"))
			Expect(container.Image).To(Equal("worker:v1.0"))
			Expect(container.ImagePullPolicy).To(Equal(corev1.PullIfNotPresent))

			// Check environment variables
			expectedEnvVars := []corev1.EnvVar{
				{Name: "CUSTOM_VAR", Value: "custom-value"},
				{Name: "WORKER_CONCURRENCY", Value: "20"},
				{Name: "WORKER_MAX_IN_FLIGHT", Value: "200"},
				{Name: "WORKER_QUEUE", Value: "test-queue"},
				{Name: "WORKER_PRIORITIES", Value: "high,medium"},
			}

			for _, expectedVar := range expectedEnvVars {
				Expect(container.Env).To(ContainElement(expectedVar))
			}

			// Check probes
			Expect(container.ReadinessProbe).NotTo(BeNil())
			Expect(container.ReadinessProbe.HTTPGet.Path).To(Equal("/health/ready"))
			Expect(container.LivenessProbe).NotTo(BeNil())
			Expect(container.LivenessProbe.HTTPGet.Path).To(Equal("/health/live"))
		})
	})
})