package main

import (
	"context"
	"crypto/tls"
	"flag"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	queuev1 "github.com/flyingrobots/go-redis-work-queue/internal/kubernetes-operator/apis/v1"
	"github.com/flyingrobots/go-redis-work-queue/internal/kubernetes-operator/controllers"
	"github.com/flyingrobots/go-redis-work-queue/internal/kubernetes-operator/webhooks"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(queuev1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	var webhookPort int
	var adminAPIEndpoint string
	var metricsEndpoint string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", false,
		"If set the metrics endpoint is served securely")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	flag.IntVar(&webhookPort, "webhook-port", 9443, "The port that the webhook server serves at.")
	flag.StringVar(&adminAPIEndpoint, "admin-api-endpoint", "http://localhost:8080",
		"The endpoint URL for the queue system Admin API")
	flag.StringVar(&metricsEndpoint, "metrics-endpoint", "http://localhost:9090",
		"The endpoint URL for the Prometheus metrics server")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	tlsOpts := []func(*tls.Config){}
	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	metricsOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
	}
	if secureMetrics && !enableHTTP2 {
		metricsOptions.TLSOpts = append(metricsOptions.TLSOpts, disableHTTP2)
	}

	webhookServer := webhook.NewServer(webhook.Options{
		Port:    webhookPort,
		TLSOpts: tlsOpts,
	})

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsOptions,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "kubernetes-operator.example.com",
		WebhookServer:          webhookServer,
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Create Admin API client
	adminAPIClient, err := NewAdminAPIClient(adminAPIEndpoint)
	if err != nil {
		setupLog.Error(err, "unable to create Admin API client")
		os.Exit(1)
	}

	// Create metrics client
	metricsClient, err := NewMetricsClient(metricsEndpoint)
	if err != nil {
		setupLog.Error(err, "unable to create metrics client")
		os.Exit(1)
	}

	// Set up controllers
	if err = (&controllers.QueueReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		AdminAPIClient: adminAPIClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Queue")
		os.Exit(1)
	}

	if err = (&controllers.WorkerPoolReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		AdminAPIClient: adminAPIClient,
		MetricsClient:  metricsClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "WorkerPool")
		os.Exit(1)
	}

	// Set up webhooks
	if err = (&webhooks.QueueWebhook{}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Queue")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// AdminAPIClient implementation
type AdminAPIClient struct {
	baseURL string
	// HTTP client would be here in real implementation
}

func NewAdminAPIClient(endpoint string) (*AdminAPIClient, error) {
	return &AdminAPIClient{
		baseURL: endpoint,
	}, nil
}

func (c *AdminAPIClient) CreateQueue(ctx context.Context, config controllers.QueueConfig) error {
	// Implementation would make HTTP calls to Admin API
	setupLog.Info("Creating queue via Admin API", "queue", config.Name)
	return nil
}

func (c *AdminAPIClient) UpdateQueue(ctx context.Context, name string, config controllers.QueueConfig) error {
	// Implementation would make HTTP calls to Admin API
	setupLog.Info("Updating queue via Admin API", "queue", name)
	return nil
}

func (c *AdminAPIClient) DeleteQueue(ctx context.Context, name string) error {
	// Implementation would make HTTP calls to Admin API
	setupLog.Info("Deleting queue via Admin API", "queue", name)
	return nil
}

func (c *AdminAPIClient) GetQueueMetrics(ctx context.Context, name string) (*controllers.QueueMetrics, error) {
	// Implementation would make HTTP calls to Admin API
	return &controllers.QueueMetrics{
		BacklogSize:    0,
		ProcessingRate: 0,
		ErrorRate:      0,
		AverageLatency: 0,
		LastUpdated:    time.Now(),
	}, nil
}

func (c *AdminAPIClient) GetQueueStatus(ctx context.Context, name string) (*controllers.QueueStatus, error) {
	// Implementation would make HTTP calls to Admin API
	return &controllers.QueueStatus{
		State:       "active",
		LastUpdated: time.Now(),
	}, nil
}

// MetricsClient implementation
type MetricsClient struct {
	baseURL string
	// Prometheus client would be here in real implementation
}

func NewMetricsClient(endpoint string) (*MetricsClient, error) {
	return &MetricsClient{
		baseURL: endpoint,
	}, nil
}

func (c *MetricsClient) GetQueueBacklog(ctx context.Context, queueName string) (int64, error) {
	// Implementation would query Prometheus for queue backlog metrics
	return 0, nil
}

func (c *MetricsClient) GetQueueLatency(ctx context.Context, queueName string, percentile float64) (float64, error) {
	// Implementation would query Prometheus for latency percentiles
	return 100.0, nil
}

func (c *MetricsClient) GetWorkerMetrics(ctx context.Context, namespace, workerPoolName string) (*controllers.WorkerMetrics, error) {
	// Implementation would query Prometheus for worker metrics
	return &controllers.WorkerMetrics{
		ProcessingRate: 10.0,
		AverageLatency: 100.0,
		ActiveWorkers:  1,
		LastUpdated:    time.Now(),
	}, nil
}
