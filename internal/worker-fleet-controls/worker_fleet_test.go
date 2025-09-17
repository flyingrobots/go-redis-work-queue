package workerfleetcontrols

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) *redis.Client {
	redisAddr := os.Getenv("TEST_REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   2, // Use different DB for tests
	})

	ctx := context.Background()
	err := client.Ping(ctx).Err()
	if err != nil {
		t.Skipf("Redis not available for testing: %v", err)
	}

	client.FlushDB(ctx)

	t.Cleanup(func() {
		client.FlushDB(ctx)
		client.Close()
	})

	return client
}

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))
}

func createTestConfig() Config {
	return Config{
		HeartbeatTimeout:     30 * time.Second,
		DefaultDrainTimeout:  5 * time.Minute,
		MaxConcurrentActions: 10,
		RequireConfirmation:  true,
		SafetyChecksEnabled:  true,
		MinHealthyWorkers:    1,
		MaxDrainPercentage:   80.0,
		AuditLogRetention:    7 * 24 * time.Hour,
		EnableMetrics:        true,
		MetricsPrefix:        "worker_fleet",
	}
}

func createTestWorker(id, hostname string) *Worker {
	return &Worker{
		ID:            id,
		State:         WorkerStateRunning,
		LastHeartbeat: time.Now(),
		StartedAt:     time.Now().Add(-time.Hour),
		Version:       "1.0.0",
		Hostname:      hostname,
		PID:           12345,
		CurrentJob:    nil,
		Capabilities:  []string{"golang", "redis"},
		Metadata: map[string]interface{}{
			"test": true,
		},
		Stats: WorkerStats{
			JobsProcessed:  100,
			JobsSuccessful: 95,
			JobsFailed:     5,
			TotalRuntime:   time.Hour,
			AverageJobTime: 30 * time.Second,
			MemoryUsage:    512 * 1024 * 1024, // 512MB
			CPUUsage:       25.5,
			GoroutineCount: 10,
		},
		Config: WorkerConfig{
			MaxConcurrentJobs: 5,
			Queues:            []string{"default", "priority"},
			JobTypes:          []string{"email", "process"},
			HeartbeatInterval: 30 * time.Second,
			GracefulTimeout:   60 * time.Second,
			EnableProfiling:   false,
		},
		Labels: map[string]string{
			"env":  "test",
			"role": "worker",
		},
		Health: WorkerHealth{
			Status:        HealthStatusHealthy,
			LastCheck:     time.Now(),
			Checks:        make(map[string]HealthCheck),
			ErrorCount:    0,
			RecoveryCount: 0,
		},
	}
}

func TestWorkerRegistry_RegisterAndGetWorker(t *testing.T) {
	client := setupTestRedis(t)
	logger := createTestLogger()
	config := createTestConfig()

	registry := NewRedisWorkerRegistry(client, config, logger)

	worker := createTestWorker("worker-1", "test-host-1")

	err := registry.RegisterWorker(worker)
	require.NoError(t, err)

	retrieved, err := registry.GetWorker("worker-1")
	require.NoError(t, err)

	assert.Equal(t, worker.ID, retrieved.ID)
	assert.Equal(t, worker.State, retrieved.State)
	assert.Equal(t, worker.Hostname, retrieved.Hostname)
	assert.Equal(t, worker.Version, retrieved.Version)
}

func TestWorkerRegistry_ListWorkers(t *testing.T) {
	client := setupTestRedis(t)
	logger := createTestLogger()
	config := createTestConfig()

	registry := NewRedisWorkerRegistry(client, config, logger)

	workers := []*Worker{
		createTestWorker("worker-1", "host-1"),
		createTestWorker("worker-2", "host-2"),
		createTestWorker("worker-3", "host-1"),
	}

	workers[1].State = WorkerStatePaused
	workers[2].Health.Status = HealthStatusDegraded

	for _, worker := range workers {
		err := registry.RegisterWorker(worker)
		require.NoError(t, err)
	}

	t.Run("list all workers", func(t *testing.T) {
		request := WorkerListRequest{
			Pagination: Pagination{
				Page:     1,
				PageSize: 10,
			},
		}

		response, err := registry.ListWorkers(request)
		require.NoError(t, err)

		assert.Equal(t, 3, response.TotalCount)
		assert.Equal(t, 3, len(response.Workers))
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 1, response.TotalPages)
		assert.False(t, response.HasNext)
		assert.False(t, response.HasPrevious)
	})

	t.Run("filter by state", func(t *testing.T) {
		request := WorkerListRequest{
			Filter: WorkerFilter{
				States: []WorkerState{WorkerStateRunning},
			},
			Pagination: Pagination{
				Page:     1,
				PageSize: 10,
			},
		}

		response, err := registry.ListWorkers(request)
		require.NoError(t, err)

		assert.Equal(t, 2, response.TotalCount)
		for _, worker := range response.Workers {
			assert.Equal(t, WorkerStateRunning, worker.State)
		}
	})

	t.Run("filter by hostname", func(t *testing.T) {
		request := WorkerListRequest{
			Filter: WorkerFilter{
				Hostname: "host-1",
			},
			Pagination: Pagination{
				Page:     1,
				PageSize: 10,
			},
		}

		response, err := registry.ListWorkers(request)
		require.NoError(t, err)

		assert.Equal(t, 2, response.TotalCount)
		for _, worker := range response.Workers {
			assert.Equal(t, "host-1", worker.Hostname)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		request := WorkerListRequest{
			Pagination: Pagination{
				Page:     1,
				PageSize: 2,
			},
		}

		response, err := registry.ListWorkers(request)
		require.NoError(t, err)

		assert.Equal(t, 3, response.TotalCount)
		assert.Equal(t, 2, len(response.Workers))
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 2, response.TotalPages)
		assert.True(t, response.HasNext)
		assert.False(t, response.HasPrevious)

		request.Pagination.Page = 2
		response, err = registry.ListWorkers(request)
		require.NoError(t, err)

		assert.Equal(t, 1, len(response.Workers))
		assert.Equal(t, 2, response.Page)
		assert.False(t, response.HasNext)
		assert.True(t, response.HasPrevious)
	})
}

func TestWorkerRegistry_UpdateHeartbeat(t *testing.T) {
	client := setupTestRedis(t)
	logger := createTestLogger()
	config := createTestConfig()

	registry := NewRedisWorkerRegistry(client, config, logger)

	worker := createTestWorker("worker-1", "host-1")
	err := registry.RegisterWorker(worker)
	require.NoError(t, err)

	newHeartbeat := time.Now()
	currentJob := &ActiveJob{
		ID:        "job-123",
		Type:      "email",
		Queue:     "priority",
		StartedAt: time.Now().Add(-5 * time.Minute),
		Progress: &JobProgress{
			Percentage: 75.0,
			Stage:      "processing",
			Message:    "Processing email template",
			UpdatedAt:  time.Now(),
		},
	}

	err = registry.UpdateHeartbeat("worker-1", newHeartbeat, currentJob)
	require.NoError(t, err)

	updated, err := registry.GetWorker("worker-1")
	require.NoError(t, err)

	assert.True(t, updated.LastHeartbeat.Equal(newHeartbeat))
	assert.NotNil(t, updated.CurrentJob)
	assert.Equal(t, "job-123", updated.CurrentJob.ID)
	assert.Equal(t, 75.0, updated.CurrentJob.Progress.Percentage)
}

func TestWorkerController_PauseResumeWorkers(t *testing.T) {
	client := setupTestRedis(t)
	logger := createTestLogger()
	config := createTestConfig()
	config.SafetyChecksEnabled = false // Disable for testing

	manager := NewWorkerFleetManager(client, config, logger)

	worker1 := createTestWorker("worker-1", "host-1")
	worker2 := createTestWorker("worker-2", "host-2")

	err := manager.Registry().RegisterWorker(worker1)
	require.NoError(t, err)
	err = manager.Registry().RegisterWorker(worker2)
	require.NoError(t, err)

	t.Run("pause workers", func(t *testing.T) {
		response, err := manager.Controller().PauseWorkers([]string{"worker-1", "worker-2"}, "Test pause")
		require.NoError(t, err)

		assert.Equal(t, 2, response.TotalRequested)
		assert.Equal(t, WorkerActionPause, response.Action)
		assert.Equal(t, ActionStatusInProgress, response.Status)

		time.Sleep(100 * time.Millisecond)

		worker1Updated, err := manager.Registry().GetWorker("worker-1")
		require.NoError(t, err)
		assert.Equal(t, WorkerStatePaused, worker1Updated.State)

		worker2Updated, err := manager.Registry().GetWorker("worker-2")
		require.NoError(t, err)
		assert.Equal(t, WorkerStatePaused, worker2Updated.State)
	})

	t.Run("resume workers", func(t *testing.T) {
		response, err := manager.Controller().ResumeWorkers([]string{"worker-1", "worker-2"}, "Test resume")
		require.NoError(t, err)

		assert.Equal(t, 2, response.TotalRequested)
		assert.Equal(t, WorkerActionResume, response.Action)

		time.Sleep(100 * time.Millisecond)

		worker1Updated, err := manager.Registry().GetWorker("worker-1")
		require.NoError(t, err)
		assert.Equal(t, WorkerStateRunning, worker1Updated.State)

		worker2Updated, err := manager.Registry().GetWorker("worker-2")
		require.NoError(t, err)
		assert.Equal(t, WorkerStateRunning, worker2Updated.State)
	})
}

func TestWorkerController_DrainWorkers(t *testing.T) {
	client := setupTestRedis(t)
	logger := createTestLogger()
	config := createTestConfig()
	config.SafetyChecksEnabled = false // Disable for testing

	manager := NewWorkerFleetManager(client, config, logger)

	worker := createTestWorker("worker-1", "host-1")
	err := manager.Registry().RegisterWorker(worker)
	require.NoError(t, err)

	response, err := manager.Controller().DrainWorkers([]string{"worker-1"}, 2*time.Minute, "Test drain")
	require.NoError(t, err)

	assert.Equal(t, 1, response.TotalRequested)
	assert.Equal(t, WorkerActionDrain, response.Action)

	time.Sleep(100 * time.Millisecond)

	workerUpdated, err := manager.Registry().GetWorker("worker-1")
	require.NoError(t, err)
	assert.Equal(t, WorkerStateDraining, workerUpdated.State)
}

func TestSafetyChecker_ValidateAction(t *testing.T) {
	client := setupTestRedis(t)
	logger := createTestLogger()
	config := createTestConfig()

	registry := NewRedisWorkerRegistry(client, config, logger)
	safetyChecker := NewSafetyChecker(registry, config, logger)

	workers := []*Worker{
		createTestWorker("worker-1", "host-1"),
		createTestWorker("worker-2", "host-2"),
		createTestWorker("worker-3", "host-3"),
	}

	for _, worker := range workers {
		err := registry.RegisterWorker(worker)
		require.NoError(t, err)
	}

	t.Run("allow safe drain", func(t *testing.T) {
		request := WorkerActionRequest{
			WorkerIDs: []string{"worker-1"},
			Action:    WorkerActionDrain,
		}

		err := safetyChecker.ValidateAction(request)
		assert.NoError(t, err)
	})

	t.Run("prevent unsafe drain", func(t *testing.T) {
		request := WorkerActionRequest{
			WorkerIDs: []string{"worker-1", "worker-2", "worker-3"},
			Action:    WorkerActionDrain,
		}

		err := safetyChecker.ValidateAction(request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "percentage")
	})

	t.Run("allow force drain", func(t *testing.T) {
		request := WorkerActionRequest{
			WorkerIDs: []string{"worker-1", "worker-2", "worker-3"},
			Action:    WorkerActionDrain,
			Force:     true,
		}

		err := safetyChecker.ValidateAction(request)
		assert.NoError(t, err)
	})
}

func TestSafetyChecker_Confirmation(t *testing.T) {
	client := setupTestRedis(t)
	logger := createTestLogger()
	config := createTestConfig()

	registry := NewRedisWorkerRegistry(client, config, logger)
	safetyChecker := NewSafetyChecker(registry, config, logger)

	workers := []*Worker{
		createTestWorker("worker-1", "host-1"),
		createTestWorker("worker-2", "host-2"),
		createTestWorker("worker-3", "host-3"),
		createTestWorker("worker-4", "host-4"),
		createTestWorker("worker-5", "host-5"),
		createTestWorker("worker-6", "host-6"),
	}

	for _, worker := range workers {
		err := registry.RegisterWorker(worker)
		require.NoError(t, err)
	}

	workerIDs := []string{"worker-1", "worker-2", "worker-3", "worker-4", "worker-5"}

	t.Run("requires confirmation for large drain", func(t *testing.T) {
		requires := safetyChecker.RequiresConfirmation(WorkerActionDrain, workerIDs)
		assert.True(t, requires)
	})

	t.Run("generate confirmation prompt", func(t *testing.T) {
		prompt := safetyChecker.GenerateConfirmationPrompt(WorkerActionDrain, workerIDs)
		assert.Contains(t, prompt, "CONFIRM")
		assert.Contains(t, prompt, "5 workers")
	})

	t.Run("validate confirmation", func(t *testing.T) {
		err := safetyChecker.ValidateConfirmation(WorkerActionDrain, workerIDs, "CONFIRM")
		assert.NoError(t, err)

		err = safetyChecker.ValidateConfirmation(WorkerActionDrain, workerIDs, "confirm")
		assert.NoError(t, err)

		err = safetyChecker.ValidateConfirmation(WorkerActionDrain, workerIDs, "wrong")
		assert.Error(t, err)
	})
}

func TestAuditLogger_LogAndRetrieve(t *testing.T) {
	client := setupTestRedis(t)
	logger := createTestLogger()
	config := createTestConfig()

	auditLogger := NewRedisAuditLogger(client, config, logger)

	log1 := AuditLog{
		ID:        "audit-1",
		Timestamp: time.Now(),
		Action:    WorkerActionPause,
		WorkerIDs: []string{"worker-1", "worker-2"},
		UserID:    "user-123",
		Reason:    "Maintenance",
		Success:   true,
		Duration:  2 * time.Second,
		Metadata: map[string]interface{}{
			"test": true,
		},
	}

	log2 := AuditLog{
		ID:        "audit-2",
		Timestamp: time.Now().Add(time.Minute),
		Action:    WorkerActionDrain,
		WorkerIDs: []string{"worker-3"},
		UserID:    "user-456",
		Reason:    "Deploy",
		Success:   false,
		Error:     "Worker not found",
		Duration:  1 * time.Second,
	}

	err := auditLogger.LogAction(log1)
	require.NoError(t, err)

	err = auditLogger.LogAction(log2)
	require.NoError(t, err)

	t.Run("get all logs", func(t *testing.T) {
		filter := AuditLogFilter{
			Limit: 10,
		}

		logs, err := auditLogger.GetAuditLogs(filter)
		require.NoError(t, err)

		assert.Len(t, logs, 2)
	})

	t.Run("filter by action", func(t *testing.T) {
		filter := AuditLogFilter{
			Actions: []WorkerAction{WorkerActionPause},
			Limit:   10,
		}

		logs, err := auditLogger.GetAuditLogs(filter)
		require.NoError(t, err)

		assert.Len(t, logs, 1)
		assert.Equal(t, WorkerActionPause, logs[0].Action)
	})

	t.Run("filter by success", func(t *testing.T) {
		success := false
		filter := AuditLogFilter{
			Success: &success,
			Limit:   10,
		}

		logs, err := auditLogger.GetAuditLogs(filter)
		require.NoError(t, err)

		assert.Len(t, logs, 1)
		assert.False(t, logs[0].Success)
	})

	t.Run("get logs by worker", func(t *testing.T) {
		logs, err := auditLogger.GetAuditLogsByWorker("worker-1", 10)
		require.NoError(t, err)

		assert.Len(t, logs, 1)
		assert.Contains(t, logs[0].WorkerIDs, "worker-1")
	})

	t.Run("get logs by user", func(t *testing.T) {
		logs, err := auditLogger.GetAuditLogsByUser("user-123", 10)
		require.NoError(t, err)

		assert.Len(t, logs, 1)
		assert.Equal(t, "user-123", logs[0].UserID)
	})
}

func TestSignalHandler_SendReceive(t *testing.T) {
	client := setupTestRedis(t)
	logger := createTestLogger()
	config := createTestConfig()

	signalHandler := NewRedisSignalHandler(client, config, logger)

	workerID := "worker-1"

	signal1 := WorkerSignal{
		Type:      SignalTypePause,
		Timestamp: time.Now(),
		Source:    "test",
	}

	signal2 := WorkerSignal{
		Type:      SignalTypeResume,
		Timestamp: time.Now(),
		Source:    "test",
	}

	err := signalHandler.SendSignal(workerID, signal1)
	require.NoError(t, err)

	err = signalHandler.SendSignal(workerID, signal2)
	require.NoError(t, err)

	signals, err := signalHandler.ReceiveSignals(workerID)
	require.NoError(t, err)

	receivedSignals := make([]WorkerSignal, 0)
	timeout := time.After(2 * time.Second)

	for i := 0; i < 2; i++ {
		select {
		case signal := <-signals:
			receivedSignals = append(receivedSignals, signal)
		case <-timeout:
			t.Fatal("Timeout waiting for signals")
		}
	}

	assert.Len(t, receivedSignals, 2)

	assert.Equal(t, SignalTypeResume, receivedSignals[0].Type)
	assert.Equal(t, SignalTypePause, receivedSignals[1].Type)
}

func TestWorkerFleetManager_Integration(t *testing.T) {
	client := setupTestRedis(t)
	logger := createTestLogger()
	config := createTestConfig()
	config.SafetyChecksEnabled = false

	manager := NewWorkerFleetManager(client, config, logger)

	err := manager.Start()
	require.NoError(t, err)
	defer manager.Stop()

	workers := []*Worker{
		createTestWorker("worker-1", "host-1"),
		createTestWorker("worker-2", "host-2"),
		createTestWorker("worker-3", "host-3"),
	}

	for _, worker := range workers {
		err := manager.Registry().RegisterWorker(worker)
		require.NoError(t, err)
	}

	summary, err := manager.Registry().GetFleetSummary()
	require.NoError(t, err)

	assert.Equal(t, 3, summary.TotalWorkers)
	assert.Equal(t, 3, summary.StateDistribution[WorkerStateRunning])
	assert.Equal(t, 3, summary.HealthDistribution[HealthStatusHealthy])

	_, err = manager.Controller().PauseWorkers([]string{"worker-1"}, "Integration test")
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	worker1, err := manager.Registry().GetWorker("worker-1")
	require.NoError(t, err)
	assert.Equal(t, WorkerStatePaused, worker1.State)

	err = manager.Health()
	assert.NoError(t, err)
}
