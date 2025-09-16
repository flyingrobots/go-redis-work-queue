//go:build integration

package main

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	queuev1 "github.com/flyingrobots/go-redis-work-queue/internal/kubernetes-operator/apis/v1"
)

var (
	k8sClient client.Client
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Integration Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{"config/crd/bases"},
		ErrorIfCRDPathMissing: false,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = queuev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Queue Controller Integration", func() {
	const (
		QueueName      = "test-queue"
		QueueNamespace = "default"
		timeout        = time.Second * 30
		interval       = time.Millisecond * 250
	)

	Context("When creating a Queue", func() {
		It("Should create and manage the queue successfully", func() {
			By("Creating a new Queue")
			queue := &queuev1.Queue{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "queue.example.com/v1",
					Kind:       "Queue",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      QueueName,
					Namespace: QueueNamespace,
				},
				Spec: queuev1.QueueSpec{
					Name:     QueueName,
					Priority: "high",
					RateLimit: &queuev1.RateLimitSpec{
						RequestsPerSecond: 100,
						BurstCapacity:     200,
						Enabled:           true,
					},
					DeadLetterQueue: &queuev1.DeadLetterQueueSpec{
						Enabled:    true,
						MaxRetries: 3,
						RetryBackoff: &queuev1.RetryBackoffSpec{
							InitialDelay: metav1.Duration{Duration: time.Second},
							MaxDelay:     metav1.Duration{Duration: 30 * time.Second},
							Multiplier:   2.0,
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, queue)).Should(Succeed())

			queueLookupKey := types.NamespacedName{Name: QueueName, Namespace: QueueNamespace}
			createdQueue := &queuev1.Queue{}

			// We'll need to retry getting this newly created Queue, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, queueLookupKey, createdQueue)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Verify the queue spec was set correctly
			Expect(createdQueue.Spec.Name).Should(Equal(QueueName))
			Expect(createdQueue.Spec.Priority).Should(Equal("high"))
			Expect(createdQueue.Spec.RateLimit.RequestsPerSecond).Should(Equal(100.0))

			By("Checking the Queue status")
			// In a real integration test, we would wait for the controller to update the status
			// For now, we'll just verify the resource exists
			Expect(createdQueue.Status.Phase).Should(BeEmpty()) // Initially empty until controller processes

			By("Cleaning up the Queue")
			Eventually(func() error {
				f := &queuev1.Queue{}
				k8sClient.Get(ctx, queueLookupKey, f)
				return k8sClient.Delete(ctx, f)
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				f := &queuev1.Queue{}
				return k8sClient.Get(ctx, queueLookupKey, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})

var _ = Describe("WorkerPool Controller Integration", func() {
	const (
		WorkerPoolName      = "test-worker-pool"
		WorkerPoolNamespace = "default"
		timeout             = time.Second * 30
		interval            = time.Millisecond * 250
	)

	Context("When creating a WorkerPool", func() {
		It("Should create and manage the worker pool successfully", func() {
			By("Creating a new WorkerPool")
			replicas := int32(2)
			minReplicas := int32(1)
			workerPool := &queuev1.WorkerPool{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "queue.example.com/v1",
					Kind:       "WorkerPool",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      WorkerPoolName,
					Namespace: WorkerPoolNamespace,
				},
				Spec: queuev1.WorkerPoolSpec{
					Replicas: &replicas,
					QueueSelector: queuev1.QueueSelector{
						Queue:      "test-queue",
						Priorities: []string{"high", "medium"},
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
					AutoScaling: &queuev1.AutoScalingSpec{
						MinReplicas:            &minReplicas,
						MaxReplicas:            10,
						TargetBacklogPerWorker: 50,
						ScaleUpCooldown:        metav1.Duration{Duration: 3 * time.Minute},
						ScaleDownCooldown:      metav1.Duration{Duration: 5 * time.Minute},
					},
				},
			}

			Expect(k8sClient.Create(ctx, workerPool)).Should(Succeed())

			workerPoolLookupKey := types.NamespacedName{Name: WorkerPoolName, Namespace: WorkerPoolNamespace}
			createdWorkerPool := &queuev1.WorkerPool{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, workerPoolLookupKey, createdWorkerPool)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Verify the worker pool spec was set correctly
			Expect(createdWorkerPool.Spec.QueueSelector.Queue).Should(Equal("test-queue"))
			Expect(createdWorkerPool.Spec.Template.Spec.Image).Should(Equal("worker:latest"))
			Expect(createdWorkerPool.Spec.AutoScaling.MaxReplicas).Should(Equal(int32(10)))

			By("Verifying WorkerPool validation")
			// Test validation by trying to create an invalid WorkerPool
			invalidWorkerPool := workerPool.DeepCopy()
			invalidWorkerPool.Name = "invalid-worker-pool"
			invalidWorkerPool.Spec.Template.Spec.Concurrency = 0 // Invalid concurrency

			err := k8sClient.Create(ctx, invalidWorkerPool)
			Expect(err).Should(HaveOccurred()) // Should fail validation

			By("Testing WorkerPool updates")
			// Update the worker pool
			Eventually(func() error {
				if err := k8sClient.Get(ctx, workerPoolLookupKey, createdWorkerPool); err != nil {
					return err
				}
				createdWorkerPool.Spec.Template.Spec.Concurrency = 20
				return k8sClient.Update(ctx, createdWorkerPool)
			}, timeout, interval).Should(Succeed())

			// Verify the update
			Eventually(func() int32 {
				k8sClient.Get(ctx, workerPoolLookupKey, createdWorkerPool)
				return createdWorkerPool.Spec.Template.Spec.Concurrency
			}, timeout, interval).Should(Equal(int32(20)))

			By("Cleaning up the WorkerPool")
			Eventually(func() error {
				f := &queuev1.WorkerPool{}
				k8sClient.Get(ctx, workerPoolLookupKey, f)
				return k8sClient.Delete(ctx, f)
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				f := &queuev1.WorkerPool{}
				return k8sClient.Get(ctx, workerPoolLookupKey, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})

var _ = Describe("Policy Controller Integration", func() {
	const (
		PolicyName      = "test-policy"
		PolicyNamespace = "default"
		timeout         = time.Second * 30
		interval        = time.Millisecond * 250
	)

	Context("When creating a Policy", func() {
		It("Should create and manage the policy successfully", func() {
			By("Creating a new Policy")
			policy := &queuev1.Policy{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "queue.example.com/v1",
					Kind:       "Policy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      PolicyName,
					Namespace: PolicyNamespace,
				},
				Spec: queuev1.PolicySpec{
					CircuitBreaker: &queuev1.CircuitBreakerSpec{
						FailureThreshold:  5,
						RecoveryThreshold: 3,
						Timeout:           metav1.Duration{Duration: 60 * time.Second},
					},
					RetryDefaults: &queuev1.RetryBackoffSpec{
						InitialDelay: metav1.Duration{Duration: time.Second},
						MaxDelay:     metav1.Duration{Duration: 30 * time.Second},
						Multiplier:   2.0,
					},
					SecurityPolicy: &queuev1.SecurityPolicySpec{
						RequiredServiceAccount: "queue-worker",
						PodSecurityStandards: &queuev1.PodSecurityStandardsSpec{
							Enforce:                  "baseline",
							AllowPrivilegeEscalation: false,
							RunAsNonRoot:             true,
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, policy)).Should(Succeed())

			policyLookupKey := types.NamespacedName{Name: PolicyName, Namespace: PolicyNamespace}
			createdPolicy := &queuev1.Policy{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, policyLookupKey, createdPolicy)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Verify the policy spec was set correctly
			Expect(createdPolicy.Spec.CircuitBreaker.FailureThreshold).Should(Equal(int32(5)))
			Expect(createdPolicy.Spec.SecurityPolicy.RequiredServiceAccount).Should(Equal("queue-worker"))

			By("Cleaning up the Policy")
			Eventually(func() error {
				f := &queuev1.Policy{}
				k8sClient.Get(ctx, policyLookupKey, f)
				return k8sClient.Delete(ctx, f)
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				f := &queuev1.Policy{}
				return k8sClient.Get(ctx, policyLookupKey, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})

// End-to-end scenario test
var _ = Describe("End-to-End Queue System", func() {
	const (
		QueueName           = "e2e-queue"
		WorkerPoolName      = "e2e-workers"
		PolicyName          = "e2e-policy"
		Namespace           = "default"
		timeout             = time.Second * 60
		interval            = time.Millisecond * 500
	)

	Context("When setting up a complete queue system", func() {
		It("Should create and coordinate all components", func() {
			By("Creating a Policy first")
			policy := &queuev1.Policy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      PolicyName,
					Namespace: Namespace,
				},
				Spec: queuev1.PolicySpec{
					RetryDefaults: &queuev1.RetryBackoffSpec{
						InitialDelay: metav1.Duration{Duration: time.Second},
						MaxDelay:     metav1.Duration{Duration: 30 * time.Second},
						Multiplier:   2.0,
					},
				},
			}
			Expect(k8sClient.Create(ctx, policy)).Should(Succeed())

			By("Creating a Queue")
			queue := &queuev1.Queue{
				ObjectMeta: metav1.ObjectMeta{
					Name:      QueueName,
					Namespace: Namespace,
				},
				Spec: queuev1.QueueSpec{
					Name:     QueueName,
					Priority: "medium",
					RateLimit: &queuev1.RateLimitSpec{
						RequestsPerSecond: 50,
						BurstCapacity:     100,
						Enabled:           true,
					},
				},
			}
			Expect(k8sClient.Create(ctx, queue)).Should(Succeed())

			By("Creating a WorkerPool")
			replicas := int32(1)
			workerPool := &queuev1.WorkerPool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      WorkerPoolName,
					Namespace: Namespace,
				},
				Spec: queuev1.WorkerPoolSpec{
					Replicas: &replicas,
					QueueSelector: queuev1.QueueSelector{
						Queue: QueueName,
					},
					Template: queuev1.WorkerPodTemplate{
						Spec: queuev1.WorkerPodSpec{
							Image:       "worker:e2e",
							Concurrency: 5,
							MaxInFlight: 50,
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, workerPool)).Should(Succeed())

			By("Verifying all resources are created")
			// Check Queue exists
			queueKey := types.NamespacedName{Name: QueueName, Namespace: Namespace}
			createdQueue := &queuev1.Queue{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, queueKey, createdQueue)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Check WorkerPool exists
			workerPoolKey := types.NamespacedName{Name: WorkerPoolName, Namespace: Namespace}
			createdWorkerPool := &queuev1.WorkerPool{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, workerPoolKey, createdWorkerPool)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Check Policy exists
			policyKey := types.NamespacedName{Name: PolicyName, Namespace: Namespace}
			createdPolicy := &queuev1.Policy{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, policyKey, createdPolicy)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying resource relationships")
			// WorkerPool should reference the correct Queue
			Expect(createdWorkerPool.Spec.QueueSelector.Queue).Should(Equal(QueueName))

			By("Cleaning up all resources")
			Expect(k8sClient.Delete(ctx, createdWorkerPool)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, createdQueue)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, createdPolicy)).Should(Succeed())

			// Verify cleanup
			Eventually(func() error {
				return k8sClient.Get(ctx, workerPoolKey, &queuev1.WorkerPool{})
			}, timeout, interval).ShouldNot(Succeed())

			Eventually(func() error {
				return k8sClient.Get(ctx, queueKey, &queuev1.Queue{})
			}, timeout, interval).ShouldNot(Succeed())

			Eventually(func() error {
				return k8sClient.Get(ctx, policyKey, &queuev1.Policy{})
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
