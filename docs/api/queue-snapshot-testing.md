# Queue Snapshot Testing API Documentation

## Overview

The Queue Snapshot Testing system provides a comprehensive solution for capturing, comparing, and asserting queue states. It enables regression testing, state debugging, and scenario replay by creating versioned snapshots of complete queue system states.

## Core Components

### SnapshotManager

The main orchestrator for all snapshot operations.

```go
manager, err := NewSnapshotManager(config, redisClient, logger)
```

#### Configuration

```go
config := &SnapshotConfig{
    // Storage settings
    StoragePath:     "./snapshots",      // Directory for snapshot files
    MaxSnapshots:    100,                 // Maximum snapshots to retain
    RetentionDays:   30,                  // Auto-delete after N days
    CompressLevel:   gzip.BestSpeed,      // Compression level (0-9)

    // Diff settings
    IgnoreTimestamps: true,               // Ignore timestamp differences
    IgnoreIDs:        true,               // Ignore job/worker ID differences
    IgnoreWorkerIDs:  true,               // Ignore worker IDs specifically
    CustomIgnores:    []string{"temp_"},  // Custom patterns to ignore

    // Performance settings
    MaxJobsPerSnapshot: 10000,            // Limit jobs per snapshot
    SampleRate:         1.0,               // Sampling rate (0.0-1.0)
    TimeoutSeconds:     30,                // Capture timeout
}
```

### Snapshot Operations

#### Capture Snapshot

Captures the current state of all queues, jobs, workers, and metrics.

```go
snapshot, err := manager.CaptureSnapshot(ctx,
    "pre-deployment",           // Name
    "State before v2.0 deploy",  // Description
    []string{"production", "v1"} // Tags
)
```

#### Load Snapshot

Loads a previously captured snapshot by ID.

```go
snapshot, err := manager.LoadSnapshot("snapshot-id-123")
```

#### Restore Snapshot

Restores the queue system to a previous state.

```go
err := manager.RestoreSnapshot(ctx, "snapshot-id-123")
```

**Warning:** This operation clears current state before restoration.

#### Compare Snapshots

Performs intelligent comparison between two snapshots.

```go
diff, err := manager.CompareSnapshots("before-id", "after-id")

// Analyze results
fmt.Printf("Total changes: %d\n", diff.TotalChanges)
fmt.Printf("Added: %d, Removed: %d, Modified: %d\n",
    diff.Added, diff.Removed, diff.Modified)

// Check semantic changes
for _, change := range diff.SemanticChanges {
    fmt.Printf("Detected: %s (severity: %s)\n",
        change.Description, change.Severity)
}
```

#### Assert Snapshot

Validates that current state matches an expected snapshot.

```go
result, err := manager.AssertSnapshot(ctx, "expected-snapshot-id")

if !result.Passed {
    fmt.Printf("Assertion failed: %s\n", result.Message)
    for _, diff := range result.Differences {
        fmt.Printf("- %s: %v -> %v\n",
            diff.Path, diff.OldValue, diff.NewValue)
    }
}
```

### Test Framework Integration

The `TestHelper` provides Jest-style snapshot testing for Go tests.

```go
func TestQueueBehavior(t *testing.T) {
    helper := NewTestHelper(t, redisClient)
    defer helper.Cleanup()

    // Setup your test scenario
    processJobs()

    // Assert state matches snapshot
    helper.AssertSnapshot(t, "processed-state")
}
```

Run with `UPDATE_SNAPSHOTS=true` to update snapshots:

```bash
UPDATE_SNAPSHOTS=true go test ./...
```

### Smart Diffing

The differ intelligently compares snapshots:

- **Semantic Analysis**: Detects high-level changes like queue overload or worker scaling
- **Configurable Ignores**: Skip timestamps, IDs, or custom patterns
- **Impact Assessment**: Rates changes as low/medium/high impact
- **Movement Detection**: Identifies jobs moved between queues

```go
differ := NewDiffer(&SnapshotConfig{
    IgnoreTimestamps: true,
    IgnoreIDs:        true,
    CustomIgnores:    []string{"session_", "temp_"},
})

diff, err := differ.Compare(snapshot1, snapshot2)
```

## Data Types

### Snapshot

Complete system state at a point in time:

```go
type Snapshot struct {
    ID          string              // Unique identifier
    Name        string              // Human-readable name
    Description string              // Detailed description
    Version     string              // Format version
    CreatedAt   time.Time           // Creation timestamp
    CreatedBy   string              // User who created it
    Tags        []string            // Categorization tags

    // State data
    Queues      []QueueState        // All queue states
    Jobs        []JobState          // Captured jobs
    Workers     []WorkerState       // Worker states
    Metrics     map[string]interface{} // System metrics

    // Metadata
    Context     map[string]interface{} // Additional context
    Environment string              // Environment name
    Checksum    string              // Data integrity checksum
    Compressed  bool                // Whether compressed
    SizeBytes   int64               // Storage size
}
```

### DiffResult

Comparison results between snapshots:

```go
type DiffResult struct {
    LeftID      string              // First snapshot ID
    RightID     string              // Second snapshot ID
    Timestamp   time.Time           // Comparison time

    // Summary
    TotalChanges int                // Total number of changes
    Added        int                // Items added
    Removed      int                // Items removed
    Modified     int                // Items modified

    // Detailed changes
    QueueChanges  []Change          // Queue-level changes
    JobChanges    []Change          // Job-level changes
    WorkerChanges []Change          // Worker changes
    MetricChanges []Change          // Metric changes

    // High-level analysis
    SemanticChanges []SemanticChange // Semantic interpretations
}
```

## Usage Examples

### Basic Snapshot Testing

```go
// Capture baseline
baseline, _ := manager.CaptureSnapshot(ctx, "baseline",
    "Initial state", []string{"test"})

// Run your operations
runQueueOperations()

// Capture result
result, _ := manager.CaptureSnapshot(ctx, "after-ops",
    "After operations", []string{"test"})

// Compare
diff, _ := manager.CompareSnapshots(baseline.ID, result.ID)

// Verify expectations
assert.Equal(t, 10, diff.TotalChanges)
assert.Greater(t, diff.Added, 0)
```

### Regression Testing

```go
func TestQueueRegression(t *testing.T) {
    helper := NewTestHelper(t, redis)

    // Load known good state
    helper.RestoreSnapshot(t, "golden-state")

    // Run system under test
    err := processQueueBatch()
    require.NoError(t, err)

    // Assert final state matches expected
    helper.AssertSnapshot(t, "expected-final-state")
}
```

### Debugging Production Issues

```go
// Capture production state
prodSnapshot, _ := manager.CaptureSnapshot(ctx, "prod-issue",
    "Production issue state", []string{"debug", "production"})

// In development, restore and debug
devManager.RestoreSnapshot(ctx, prodSnapshot.ID)

// Reproduce and fix issue
debugAndFix()

// Capture fixed state
fixedSnapshot, _ := devManager.CaptureSnapshot(ctx, "fixed",
    "After fix", []string{"fixed"})

// Verify fix
diff, _ := manager.CompareSnapshots(prodSnapshot.ID, fixedSnapshot.ID)
```

### CI/CD Integration

```go
// Pre-deployment snapshot
preDeploySnap, _ := manager.CaptureSnapshot(ctx, "pre-deploy",
    fmt.Sprintf("Before deploying %s", version),
    []string{"deployment", version})

// Deploy
deployNewVersion()

// Post-deployment snapshot
postDeploySnap, _ := manager.CaptureSnapshot(ctx, "post-deploy",
    fmt.Sprintf("After deploying %s", version),
    []string{"deployment", version})

// Analyze impact
diff, _ := manager.CompareSnapshots(preDeploySnap.ID, postDeploySnap.ID)

// Check for unexpected changes
if diff.TotalChanges > expectedChanges {
    // Rollback if needed
    manager.RestoreSnapshot(ctx, preDeploySnap.ID)
    return fmt.Errorf("deployment caused unexpected changes")
}
```

### Test Fixtures

```go
fixtures := NewFixtures(manager)

// Load predefined scenarios
fixtures.LoadScenario(ctx, "simple")   // Basic queue with jobs
fixtures.LoadScenario(ctx, "complex")  // Multiple queues, workers
fixtures.LoadScenario(ctx, "error")    // Error conditions
fixtures.LoadScenario(ctx, "empty")    // Clean slate

// Custom scenario
func loadCustomScenario(ctx context.Context) error {
    // Setup specific test conditions
    for i := 0; i < 100; i++ {
        redis.RPush(ctx, "queue:test", fmt.Sprintf("job-%d", i))
    }
    return nil
}
```

## Performance Considerations

### Storage Optimization

- **Compression**: Enable compression for large snapshots (reduces size by 60-80%)
- **Sampling**: Use `SampleRate` < 1.0 for statistical sampling
- **Job Limits**: Set `MaxJobsPerSnapshot` to cap captured jobs
- **Retention**: Configure `RetentionDays` for automatic cleanup

### Capture Performance

Typical capture times:
- Small (< 100 jobs): < 100ms
- Medium (1000 jobs): 200-500ms
- Large (10000 jobs): 1-2 seconds
- Very Large (100000 jobs): 5-10 seconds

### Comparison Performance

Diff operations are O(n) where n is the number of items:
- Small snapshots: < 50ms
- Large snapshots: 100-500ms
- Semantic analysis adds ~10-20% overhead

## Best Practices

### Snapshot Naming

Use descriptive, consistent naming:
```go
// Good
"pre-v2.0-deployment"
"after-batch-processing"
"regression-test-baseline"

// Bad
"snapshot1"
"test"
"temp"
```

### Tagging Strategy

Use tags for categorization:
```go
tags := []string{
    "environment:production",
    "version:2.0.1",
    "test:regression",
    "feature:bulk-import",
}
```

### Ignore Configuration

Configure ignores based on test needs:
```go
// For functional tests
config.IgnoreTimestamps = true
config.IgnoreIDs = true

// For exact state matching
config.IgnoreTimestamps = false
config.IgnoreIDs = false

// Custom patterns
config.CustomIgnores = []string{
    "session_",      // Ignore session IDs
    "temp_",         // Ignore temporary data
    "_timestamp",    // Ignore timestamp fields
}
```

### Version Control

Store snapshots in git for team sharing:
```bash
# Add snapshot directory to git
git add testdata/snapshots/

# Use git LFS for large snapshots
git lfs track "*.snapshot.gz"
```

## Troubleshooting

### Snapshot Assertion Failures

1. Check if snapshot exists:
```go
if !storage.Exists(snapshotID) {
    // Run with UPDATE_SNAPSHOTS=true to create
}
```

2. Review differences:
```go
for _, diff := range result.Differences {
    log.Printf("Path: %s, Type: %s, Impact: %s",
        diff.Path, diff.Type, diff.Impact)
}
```

3. Update if legitimate change:
```bash
UPDATE_SNAPSHOTS=true go test
```

### Storage Issues

1. Check disk space:
```go
info, _ := os.Stat(config.StoragePath)
fmt.Printf("Storage size: %d bytes\n", info.Size())
```

2. Clean old snapshots:
```go
filter := &SnapshotFilter{
    CreatedBefore: time.Now().Add(-30 * 24 * time.Hour),
}
old, _ := storage.List(filter)
for _, snap := range old {
    storage.Delete(snap.ID)
}
```

### Performance Problems

1. Reduce captured data:
```go
config.MaxJobsPerSnapshot = 1000
config.SampleRate = 0.1  // Sample 10%
```

2. Enable compression:
```go
config.CompressLevel = gzip.BestSpeed
```

3. Use parallel capture for multiple queues:
```go
var wg sync.WaitGroup
for _, queue := range queues {
    wg.Add(1)
    go func(q string) {
        defer wg.Done()
        captureQueue(q)
    }(queue)
}
wg.Wait()
```

## Security Considerations

### Sensitive Data

- Snapshots may contain sensitive job payloads
- Store snapshots securely with appropriate permissions
- Consider encryption for production snapshots
- Sanitize snapshots before sharing

### Access Control

```go
// Restrict snapshot operations
if !user.HasPermission("snapshot:write") {
    return ErrUnauthorized
}

// Audit snapshot operations
audit.Log("snapshot.captured", map[string]interface{}{
    "user":     user.ID,
    "snapshot": snapshot.ID,
    "action":   "capture",
})
```

### Data Retention

- Implement automatic cleanup for old snapshots
- Comply with data retention policies
- Consider GDPR/privacy requirements

## Migration Guide

### From Manual Testing

Before:
```go
// Manual state verification
jobs := getJobs()
assert.Equal(t, 10, len(jobs))
assert.Equal(t, "pending", jobs[0].Status)
```

After:
```go
// Snapshot-based verification
helper.AssertSnapshot(t, "expected-state")
```

### From Custom Comparison

Before:
```go
// Custom comparison logic
differences := compareStates(before, after)
```

After:
```go
// Built-in intelligent diffing
diff, _ := manager.CompareSnapshots(beforeID, afterID)
```