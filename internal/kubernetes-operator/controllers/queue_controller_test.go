package controllers

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	queuev1 "github.com/flyingrobots/go-redis-work-queue/internal/kubernetes-operator/apis/v1"
)

// MockAdminAPIClient for testing
type MockAdminAPIClient struct {
	queues map[string]*QueueConfig
	status map[string]*QueueStatus
	metrics map[string]*QueueMetrics
	errors map[string]error
}

func NewMockAdminAPIClient() *MockAdminAPIClient {
	return &MockAdminAPIClient{
		queues:  make(map[string]*QueueConfig),
		status:  make(map[string]*QueueStatus),
		metrics: make(map[string]*QueueMetrics),
		errors:  make(map[string]error),
	}
}

func (m *MockAdminAPIClient) CreateQueue(ctx context.Context, config QueueConfig) error {
	if err, exists := m.errors["create"]; exists {
		return err
	}
	m.queues[config.Name] = &config
	m.status[config.Name] = &QueueStatus{
		State:       "active",
		LastUpdated: time.Now(),
	}
	m.metrics[config.Name] = &QueueMetrics{
		BacklogSize:    0,
		ProcessingRate: 0,
		ErrorRate:      0,
		AverageLatency: 0,
		LastUpdated:    time.Now(),
	}
	return nil
}

func (m *MockAdminAPIClient) UpdateQueue(ctx context.Context, name string, config QueueConfig) error {
	if err, exists := m.errors["update"]; exists {
		return err
	}
	m.queues[name] = &config
	return nil
}

func (m *MockAdminAPIClient) DeleteQueue(ctx context.Context, name string) error {
	if err, exists := m.errors["delete"]; exists {
		return err
	}
	delete(m.queues, name)
	delete(m.status, name)
	delete(m.metrics, name)
	return nil
}

func (m *MockAdminAPIClient) GetQueueStatus(ctx context.Context, name string) (*QueueStatus, error) {
	if err, exists := m.errors["status"]; exists {
		return nil, err
	}
	if status, exists := m.status[name]; exists {
		return status, nil
	}
	return nil, fmt.Errorf("queue not found")
}

func (m *MockAdminAPIClient) GetQueueMetrics(ctx context.Context, name string) (*QueueMetrics, error) {
	if err, exists := m.errors["metrics"]; exists {
		return nil, err
	}
	if metrics, exists := m.metrics[name]; exists {
		return metrics, nil
	}
	return &QueueMetrics{}, nil
}

// SetError allows setting errors for testing failure scenarios
func (m *MockAdminAPIClient) SetError(operation string, err error) {
	m.errors[operation] = err
}

func TestQueueController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Queue Controller Suite")
}

var _ = Describe("QueueController", func() {
	var (
		ctx           context.Context
		k8sClient     client.Client
		reconciler    *QueueReconciler
		mockAdminAPI  *MockAdminAPIClient
		scheme        *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(queuev1.AddToScheme(scheme)).To(Succeed())
		Expect(corev1.AddToScheme(scheme)).To(Succeed())

		k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		mockAdminAPI = NewMockAdminAPIClient()

		reconciler = &QueueReconciler{
			Client:         k8sClient,
			Scheme:         scheme,
			AdminAPIClient: mockAdminAPI,
		}
	})

	Describe("Reconciling a Queue", func() {
		var queue *queuev1.Queue

		BeforeEach(func() {
			queue = &queuev1.Queue{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-queue",
					Namespace: "default",
				},
				Spec: queuev1.QueueSpec{
					Name:     "test-queue",
					Priority: "medium",
					RateLimit: &queuev1.RateLimitSpec{
						RequestsPerSecond: 100,
						BurstCapacity:     200,
						Enabled:           true,
					},
				},
			}
		})

		Context("When creating a new Queue", func() {
			It("Should create the queue successfully", func() {
				Expect(k8sClient.Create(ctx, queue)).To(Succeed())

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      queue.Name,
						Namespace: queue.Namespace,
					},
				}

				result, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RequeueAfter).To(Equal(30 * time.Second))

				// Verify queue was created in mock API
				Expect(mockAdminAPI.queues).To(HaveKey("test-queue"))
				Expect(mockAdminAPI.queues["test-queue"].Priority).To(Equal("medium"))

				// Verify status was updated
				var updatedQueue queuev1.Queue
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: queue.Name, Namespace: queue.Namespace}, &updatedQueue)).To(Succeed())
				Expect(updatedQueue.Status.Phase).To(Equal(queuev1.QueuePhaseActive))
			})

			It("Should handle creation failures", func() {
				mockAdminAPI.SetError("create", fmt.Errorf("API error"))

				Expect(k8sClient.Create(ctx, queue)).To(Succeed())

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      queue.Name,
						Namespace: queue.Namespace,
					},
				}

				result, err := reconciler.Reconcile(ctx, req)
				Expect(err).To(HaveOccurred())
				Expect(result.RequeueAfter).To(Equal(time.Minute))

				// Verify status shows failure
				var updatedQueue queuev1.Queue
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: queue.Name, Namespace: queue.Namespace}, &updatedQueue)).To(Succeed())
				Expect(updatedQueue.Status.Phase).To(Equal(queuev1.QueuePhaseFailed))
			})
		})

		Context("When updating an existing Queue", func() {
			It("Should update the queue configuration", func() {
				// Create initial queue
				Expect(k8sClient.Create(ctx, queue)).To(Succeed())

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      queue.Name,
						Namespace: queue.Namespace,
					},
				}

				// First reconcile to create
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				// Update queue spec
				var updatedQueue queuev1.Queue
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: queue.Name, Namespace: queue.Namespace}, &updatedQueue)).To(Succeed())
				updatedQueue.Spec.RateLimit.RequestsPerSecond = 200
				Expect(k8sClient.Update(ctx, &updatedQueue)).To(Succeed())

				// Second reconcile to update
				_, err = reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				// Verify update was applied
				Expect(mockAdminAPI.queues["test-queue"].RateLimit.RequestsPerSecond).To(Equal(200.0))
			})
		})

		Context("When deleting a Queue", func() {
			It("Should perform cleanup and remove finalizer", func() {
				Expect(k8sClient.Create(ctx, queue)).To(Succeed())

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      queue.Name,
						Namespace: queue.Namespace,
					},
				}

				// First reconcile to create and add finalizer
				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				// Verify finalizer was added
				var queueWithFinalizer queuev1.Queue
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: queue.Name, Namespace: queue.Namespace}, &queueWithFinalizer)).To(Succeed())
				Expect(queueWithFinalizer.Finalizers).To(ContainElement(QueueFinalizerName))

				// Delete the queue
				Expect(k8sClient.Delete(ctx, &queueWithFinalizer)).To(Succeed())

				// Reconcile deletion
				_, err = reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				// Verify queue was deleted from mock API
				Expect(mockAdminAPI.queues).NotTo(HaveKey("test-queue"))
			})
		})

		Context("When handling Redis configuration", func() {
			It("Should handle password secrets", func() {
				// Create a secret
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "redis-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"password": []byte("secret-password"),
					},
				}
				Expect(k8sClient.Create(ctx, secret)).To(Succeed())

				// Create queue with Redis config
				queue.Spec.Redis = &queuev1.RedisSpec{
					Addresses: []string{"redis:6379"},
					Database:  1,
					PasswordSecret: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "redis-secret",
						},
						Key: "password",
					},
				}

				Expect(k8sClient.Create(ctx, queue)).To(Succeed())

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      queue.Name,
						Namespace: queue.Namespace,
					},
				}

				_, err := reconciler.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())

				// Verify Redis config includes password
				queueConfig := mockAdminAPI.queues["test-queue"]
				Expect(queueConfig.Redis).NotTo(BeNil())
				Expect(queueConfig.Redis.Password).To(Equal("secret-password"))
			})
		})
	})

	Describe("buildQueueConfig", func() {
		var queue *queuev1.Queue

		BeforeEach(func() {
			queue = &queuev1.Queue{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-queue",
					Namespace: "default",
				},
				Spec: queuev1.QueueSpec{
					Name:     "test-queue",
					Priority: "high",
				},
			}
		})

		It("Should build basic configuration", func() {
			config, err := reconciler.buildQueueConfig(ctx, queue)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Name).To(Equal("test-queue"))
			Expect(config.Priority).To(Equal("high"))
		})

		It("Should handle rate limit configuration", func() {
			queue.Spec.RateLimit = &queuev1.RateLimitSpec{
				RequestsPerSecond: 150,
				BurstCapacity:     300,
				Enabled:           true,
			}

			config, err := reconciler.buildQueueConfig(ctx, queue)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.RateLimit).NotTo(BeNil())
			Expect(config.RateLimit.RequestsPerSecond).To(Equal(150.0))
			Expect(config.RateLimit.BurstCapacity).To(Equal(int32(300)))
			Expect(config.RateLimit.Enabled).To(BeTrue())
		})

		It("Should handle DLQ configuration", func() {
			queue.Spec.DeadLetterQueue = &queuev1.DeadLetterQueueSpec{
				Enabled:    true,
				MaxRetries: 5,
				RetryBackoff: &queuev1.RetryBackoffSpec{
					InitialDelay: metav1.Duration{Duration: 2 * time.Second},
					MaxDelay:     metav1.Duration{Duration: 60 * time.Second},
					Multiplier:   2.5,
				},
			}

			config, err := reconciler.buildQueueConfig(ctx, queue)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.DeadLetterQueue).NotTo(BeNil())
			Expect(config.DeadLetterQueue.Enabled).To(BeTrue())
			Expect(config.DeadLetterQueue.MaxRetries).To(Equal(int32(5)))
			Expect(config.DeadLetterQueue.RetryBackoff.InitialDelay).To(Equal(2 * time.Second))
			Expect(config.DeadLetterQueue.RetryBackoff.Multiplier).To(Equal(2.5))
		})

		It("Should handle retention configuration", func() {
			queue.Spec.Retention = &queuev1.RetentionSpec{
				CompletedJobs: metav1.Duration{Duration: 48 * time.Hour},
				FailedJobs:    metav1.Duration{Duration: 96 * time.Hour},
				MaxJobs:       50000,
			}

			config, err := reconciler.buildQueueConfig(ctx, queue)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Retention).NotTo(BeNil())
			Expect(config.Retention.CompletedJobs).To(Equal(48 * time.Hour))
			Expect(config.Retention.FailedJobs).To(Equal(96 * time.Hour))
			Expect(config.Retention.MaxJobs).To(Equal(int32(50000)))
		})
	})

	Describe("Status Updates", func() {
		var queue *queuev1.Queue

		BeforeEach(func() {
			queue = &queuev1.Queue{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-queue",
					Namespace: "default",
				},
				Spec: queuev1.QueueSpec{
					Name:     "test-queue",
					Priority: "medium",
				},
			}
			Expect(k8sClient.Create(ctx, queue)).To(Succeed())
		})

		It("Should update status with metrics", func() {
			// Set up mock metrics
			mockAdminAPI.metrics["test-queue"] = &QueueMetrics{
				BacklogSize:    100,
				ProcessingRate: 50.5,
				ErrorRate:      0.02,
				AverageLatency: 125.5,
				LastUpdated:    time.Now(),
			}

			metrics := &queuev1.QueueMetrics{
				BacklogSize:    100,
				ProcessingRate: 50.5,
				ErrorRate:      0.02,
				AverageLatency: 125.5,
				LastUpdated:    metav1.NewTime(time.Now()),
			}

			reconciler.updateQueueStatus(ctx, queue, queuev1.QueuePhaseActive, "Queue is active", metrics)

			// Verify status was updated
			var updatedQueue queuev1.Queue
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: queue.Name, Namespace: queue.Namespace}, &updatedQueue)).To(Succeed())
			Expect(updatedQueue.Status.Phase).To(Equal(queuev1.QueuePhaseActive))
			Expect(updatedQueue.Status.Metrics).NotTo(BeNil())
			Expect(updatedQueue.Status.Metrics.BacklogSize).To(Equal(int64(100)))
			Expect(updatedQueue.Status.Metrics.ProcessingRate).To(Equal(50.5))
		})

		It("Should update conditions correctly", func() {
			reconciler.updateQueueStatus(ctx, queue, queuev1.QueuePhaseActive, "Queue is ready", nil)

			var updatedQueue queuev1.Queue
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: queue.Name, Namespace: queue.Namespace}, &updatedQueue)).To(Succeed())
			Expect(updatedQueue.Status.Conditions).To(HaveLen(1))
			Expect(updatedQueue.Status.Conditions[0].Type).To(Equal("Ready"))
			Expect(updatedQueue.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))
			Expect(updatedQueue.Status.Conditions[0].Reason).To(Equal("QueueReady"))
		})
	})
})
