// Copyright 2025 James Ross
package archives

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"go.uber.org/zap"
)

func TestNewManager(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := &Config{
		RedisAddr: mr.Addr(),
		RedisDB:   0,
		Archive: ArchiveConfig{
			Enabled:      true,
			SamplingRate: 1.0,
			BatchSize:    100,
			ClickHouse: ClickHouseConfig{
				Enabled: false, // Disable for basic test
			},
			S3: S3Config{
				Enabled: false, // Disable for basic test
			},
			Retention: RetentionConfig{
				RedisStreamTTL: 24 * time.Hour,
				ArchiveWindow:  30 * 24 * time.Hour,
				DeleteAfter:    365 * 24 * time.Hour,
				GDPRCompliant:  true,
			},
			PayloadHandling: PayloadHandlingConfig{
				IncludePayload: true,
				MaxPayloadSize: 1024 * 1024,
				HashOnly:       false,
			},
		},
	}

	mgr, err := NewManager(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer mgr.Close()

	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}
}

func TestArchiveJob(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := &Config{
		RedisAddr: mr.Addr(),
		RedisDB:   0,
		Archive: ArchiveConfig{
			Enabled:        true,
			SamplingRate:   1.0,
			RedisStreamKey: "test:archive:stream",
			ClickHouse: ClickHouseConfig{
				Enabled: false,
			},
			S3: S3Config{
				Enabled: false,
			},
			PayloadHandling: PayloadHandlingConfig{
				IncludePayload: true,
				MaxPayloadSize: 1024,
				HashOnly:       false,
			},
		},
	}

	mgr, err := NewManager(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer mgr.Close()

	ctx := context.Background()

	job := ArchiveJob{
		JobID:       "test-job-123",
		Queue:       "test-queue",
		Priority:    1,
		EnqueuedAt:  time.Now().Add(-5 * time.Minute),
		StartedAt:   &time.Time{},
		CompletedAt: time.Now(),
		Outcome:     OutcomeSuccess,
		RetryCount:  0,
		WorkerID:    "worker-1",
		PayloadSize: 512,
		TraceID:     "trace-123",
		PayloadSnapshot: []byte(`{"message": "test payload"}`),
		Tags: map[string]string{
			"test": "true",
		},
	}

	*job.StartedAt = time.Now().Add(-2 * time.Minute)

	err = mgr.ArchiveJob(ctx, job)
	if err != nil {
		t.Fatalf("Failed to archive job: %v", err)
	}

	// Verify job was added to stream
	length, err := mr.XLen("test:archive:stream")
	if err != nil {
		t.Fatalf("Failed to get stream length: %v", err)
	}

	if length != 1 {
		t.Errorf("Expected 1 message in stream, got %d", length)
	}
}

func TestSchemaValidation(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := &Config{
		RedisAddr: mr.Addr(),
		RedisDB:   0,
		Archive: ArchiveConfig{
			Enabled:      true,
			SamplingRate: 1.0,
		},
	}

	mgr, err := NewManager(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer mgr.Close()

	ctx := context.Background()

	tests := []struct {
		name    string
		job     ArchiveJob
		wantErr bool
	}{
		{
			name: "valid job",
			job: ArchiveJob{
				JobID:       "valid-job",
				Queue:       "test-queue",
				CompletedAt: time.Now(),
				Outcome:     OutcomeSuccess,
				WorkerID:    "worker-1",
			},
			wantErr: false,
		},
		{
			name: "missing job ID",
			job: ArchiveJob{
				Queue:       "test-queue",
				CompletedAt: time.Now(),
				Outcome:     OutcomeSuccess,
			},
			wantErr: true,
		},
		{
			name: "missing queue",
			job: ArchiveJob{
				JobID:       "test-job",
				CompletedAt: time.Now(),
				Outcome:     OutcomeSuccess,
			},
			wantErr: true,
		},
		{
			name: "missing outcome",
			job: ArchiveJob{
				JobID:       "test-job",
				Queue:       "test-queue",
				CompletedAt: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.ArchiveJob(ctx, tt.job)
			if (err != nil) != tt.wantErr {
				t.Errorf("ArchiveJob() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSchemaManager(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := &Config{
		RedisAddr: mr.Addr(),
		RedisDB:   0,
		Archive:   ArchiveConfig{},
	}

	mgr, err := NewManager(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer mgr.Close()

	ctx := context.Background()

	// Test get current version
	version, err := mgr.GetSchemaVersion(ctx)
	if err != nil {
		t.Fatalf("Failed to get schema version: %v", err)
	}

	if version < 1 {
		t.Errorf("Expected version >= 1, got %d", version)
	}

	// Test get schema evolution
	evolution, err := mgr.GetSchemaEvolution(ctx)
	if err != nil {
		t.Fatalf("Failed to get schema evolution: %v", err)
	}

	if len(evolution) == 0 {
		t.Error("Expected at least one schema evolution")
	}
}

func TestPayloadHandling(t *testing.T) {
	tests := []struct {
		name    string
		config  PayloadHandlingConfig
		payload []byte
		wantLen int
		wantHash bool
	}{
		{
			name: "include payload",
			config: PayloadHandlingConfig{
				IncludePayload: true,
				MaxPayloadSize: 1024,
			},
			payload: []byte("test payload"),
			wantLen: 12,
			wantHash: true,
		},
		{
			name: "hash only",
			config: PayloadHandlingConfig{
				IncludePayload: true,
				HashOnly:       true,
			},
			payload: []byte("test payload"),
			wantLen: 0,
			wantHash: true,
		},
		{
			name: "size limit",
			config: PayloadHandlingConfig{
				IncludePayload: true,
				MaxPayloadSize: 5,
			},
			payload: []byte("test payload"),
			wantLen: 5,
			wantHash: true,
		},
		{
			name: "exclude payload",
			config: PayloadHandlingConfig{
				IncludePayload: false,
			},
			payload: []byte("test payload"),
			wantLen: 0,
			wantHash: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := miniredis.RunT(t)
			defer mr.Close()

			config := &Config{
				RedisAddr: mr.Addr(),
				RedisDB:   0,
				Archive: ArchiveConfig{
					Enabled:         true,
					SamplingRate:    1.0,
					PayloadHandling: tt.config,
				},
			}

			mgr, err := NewManager(config, zap.NewNop())
			if err != nil {
				t.Fatalf("Failed to create manager: %v", err)
			}
			defer mgr.Close()

			job := ArchiveJob{
				JobID:           "test-job",
				Queue:           "test-queue",
				CompletedAt:     time.Now(),
				Outcome:         OutcomeSuccess,
				PayloadSnapshot: tt.payload,
			}

			// Process the payload
			m := mgr.(*manager)
			err = m.processPayload(&job)
			if err != nil {
				t.Fatalf("Failed to process payload: %v", err)
			}

			if len(job.PayloadSnapshot) != tt.wantLen {
				t.Errorf("Expected payload length %d, got %d", tt.wantLen, len(job.PayloadSnapshot))
			}

			hasHash := job.PayloadHash != ""
			if hasHash != tt.wantHash {
				t.Errorf("Expected hash presence %v, got %v", tt.wantHash, hasHash)
			}
		})
	}
}

func TestSearchJobs(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := &Config{
		RedisAddr: mr.Addr(),
		RedisDB:   0,
		Archive: ArchiveConfig{
			Enabled:        true,
			SamplingRate:   1.0,
			RedisStreamKey: "test:archive:stream",
		},
	}

	mgr, err := NewManager(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer mgr.Close()

	ctx := context.Background()

	// Archive some test jobs
	jobs := []ArchiveJob{
		{
			JobID:       "job-1",
			Queue:       "queue-a",
			CompletedAt: time.Now(),
			Outcome:     OutcomeSuccess,
			WorkerID:    "worker-1",
		},
		{
			JobID:       "job-2",
			Queue:       "queue-b",
			CompletedAt: time.Now(),
			Outcome:     OutcomeFailed,
			WorkerID:    "worker-2",
		},
		{
			JobID:       "job-3",
			Queue:       "queue-a",
			CompletedAt: time.Now(),
			Outcome:     OutcomeSuccess,
			WorkerID:    "worker-1",
		},
	}

	for _, job := range jobs {
		err := mgr.ArchiveJob(ctx, job)
		if err != nil {
			t.Fatalf("Failed to archive job %s: %v", job.JobID, err)
		}
	}

	tests := []struct {
		name      string
		query     SearchQuery
		wantCount int
	}{
		{
			name: "search by queue",
			query: SearchQuery{
				Queue: "queue-a",
				Limit: 10,
			},
			wantCount: 2,
		},
		{
			name: "search by outcome",
			query: SearchQuery{
				Outcome: OutcomeSuccess,
				Limit:   10,
			},
			wantCount: 2,
		},
		{
			name: "search by worker",
			query: SearchQuery{
				WorkerID: "worker-1",
				Limit:    10,
			},
			wantCount: 2,
		},
		{
			name: "search by job ID",
			query: SearchQuery{
				JobIDs: []string{"job-2"},
				Limit:  10,
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := mgr.SearchJobs(ctx, tt.query)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			if len(results) != tt.wantCount {
				t.Errorf("Expected %d results, got %d", tt.wantCount, len(results))
			}
		})
	}
}

func TestQueryTemplates(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := &Config{
		RedisAddr: mr.Addr(),
		RedisDB:   0,
		Archive:   ArchiveConfig{},
	}

	mgr, err := NewManager(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer mgr.Close()

	ctx := context.Background()

	// Test get default templates
	templates, err := mgr.GetQueryTemplates(ctx)
	if err != nil {
		t.Fatalf("Failed to get templates: %v", err)
	}

	if len(templates) == 0 {
		t.Error("Expected at least some default templates")
	}

	// Test add custom template
	customTemplate := QueryTemplate{
		Name:        "test_template",
		Description: "Test template",
		SQL:         "SELECT * FROM archives WHERE queue = ?",
		Parameters: []QueryParameter{
			{
				Name:        "queue",
				Type:        "string",
				Description: "Queue name",
				Required:    true,
			},
		},
		Tags: []string{"test"},
	}

	err = mgr.AddQueryTemplate(ctx, customTemplate)
	if err != nil {
		t.Fatalf("Failed to add template: %v", err)
	}

	// Verify template was added
	templates, err = mgr.GetQueryTemplates(ctx)
	if err != nil {
		t.Fatalf("Failed to get templates after adding: %v", err)
	}

	found := false
	for _, template := range templates {
		if template.Name == "test_template" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Custom template not found in templates list")
	}
}

func TestHealthCheck(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := &Config{
		RedisAddr: mr.Addr(),
		RedisDB:   0,
		Archive:   ArchiveConfig{},
	}

	mgr, err := NewManager(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer mgr.Close()

	ctx := context.Background()

	health, err := mgr.GetHealth(ctx)
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	if health["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", health["status"])
	}

	if health["redis"] != "connected" {
		t.Errorf("Expected Redis status 'connected', got %v", health["redis"])
	}
}