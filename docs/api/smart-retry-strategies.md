# Smart Retry Strategies API

The Smart Retry Strategies module provides intelligent retry recommendations based on historical patterns, Bayesian learning, and optional machine learning models. It adapts retry timing and policy based on success patterns while maintaining safety guardrails.

## Overview

The system consists of three recommendation layers:

1. **Rule-based policies** - Baseline heuristics for common error patterns
2. **Bayesian recommender** - Data-driven recommendations using Beta distributions
3. **Optional ML models** - Advanced predictions using logistic regression or gradient boosting

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                Smart Retry Strategies Manager                │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Rule-Based   │  │ Bayesian     │  │ ML Models    │     │
│  │ Policies     │  │ Recommender  │  │ (Optional)   │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│       ↓                   ↓                   ↓             │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │            Policy Guardrails & Safety Limits           │ │
│  └─────────────────────────────────────────────────────────┘ │
│       ↓                                                     │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │         Data Collection & Feature Extraction           │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Core Concepts

### Attempt History
Each job attempt is recorded with comprehensive metadata:

```json
{
  "job_id": "task_123",
  "job_type": "email_send",
  "attempt_number": 2,
  "error_class": "rate_limit",
  "error_code": "429",
  "status": "failed",
  "queue": "default",
  "tenant": "customer_a",
  "payload_size": 1024,
  "time_of_day": 14,
  "worker_version": "v1.2.3",
  "health": {
    "cpu_usage": 0.75,
    "memory_usage": 0.60
  },
  "delay_ms": 5000,
  "success": false,
  "timestamp": "2025-01-15T14:30:00Z",
  "processing_time": "00:05.250"
}
```

### Retry Features
Features extracted for recommendation generation:

```json
{
  "job_type": "email_send",
  "error_class": "rate_limit",
  "error_code": "429",
  "attempt_number": 2,
  "queue": "default",
  "tenant": "customer_a",
  "payload_size": 1024,
  "time_of_day": 14,
  "worker_version": "v1.2.3",
  "health": {
    "downstream_latency": 250.0
  },
  "since_last_failure": "00:02:30",
  "recent_failures": 3,
  "avg_processing_time": "00:04.500"
}
```

### Retry Recommendation
Structured recommendation with confidence and rationale:

```json
{
  "should_retry": true,
  "delay_ms": 8000,
  "max_attempts": 5,
  "confidence": 0.85,
  "rationale": "Bayesian model predicts 85% success with 8s delay",
  "method": "bayesian",
  "estimated_success": 0.85,
  "next_evaluation": "2025-01-15T14:30:08Z",
  "policy_guardrails": []
}
```

## API Endpoints

### Recommendation Generation

#### Get Retry Recommendation
```http
POST /api/v1/retry/recommendation
Content-Type: application/json

{
  "job_type": "email_send",
  "error_class": "rate_limit",
  "error_code": "429",
  "attempt_number": 2,
  "queue": "default",
  "payload_size": 1024,
  "time_of_day": 14,
  "since_last_failure": "00:02:30"
}
```

**Response:**
```json
{
  "should_retry": true,
  "delay_ms": 8000,
  "max_attempts": 5,
  "confidence": 0.85,
  "rationale": "Bayesian model predicts 85% success",
  "method": "bayesian",
  "estimated_success": 0.85,
  "next_evaluation": "2025-01-15T14:30:08Z"
}
```

#### Preview Retry Schedule
```http
POST /api/v1/retry/preview
Content-Type: application/json

{
  "features": {
    "job_type": "email_send",
    "error_class": "rate_limit",
    "attempt_number": 1
  },
  "max_attempts": 5
}
```

**Response:**
```json
{
  "job_id": "email_send_preview",
  "current_attempt": 1,
  "features": { ... },
  "recommendations": [
    {
      "should_retry": true,
      "delay_ms": 2000,
      "confidence": 0.80,
      "method": "rules"
    }
  ],
  "timeline": [
    {
      "attempt_number": 2,
      "scheduled_time": "2025-01-15T14:30:02Z",
      "estimated_success": 0.70,
      "delay_ms": 2000,
      "method": "rules",
      "rationale": "Rate limit policy matched"
    }
  ],
  "generated_at": "2025-01-15T14:30:00Z"
}
```

### Data Collection

#### Record Attempt
```http
POST /api/v1/retry/attempt
Content-Type: application/json

{
  "job_id": "task_123",
  "job_type": "email_send",
  "attempt_number": 2,
  "error_class": "timeout",
  "status": "failed",
  "success": false,
  "delay_ms": 4000,
  "timestamp": "2025-01-15T14:30:00Z"
}
```

#### Get Statistics
```http
GET /api/v1/retry/stats?job_type=email_send&error_class=timeout&window=24h
```

**Response:**
```json
{
  "job_type": "email_send",
  "error_class": "timeout",
  "total_attempts": 150,
  "successful_retries": 120,
  "failed_retries": 30,
  "avg_delay_ms": 4500.0,
  "success_rate": 0.80,
  "last_updated": "2025-01-15T14:30:00Z",
  "window_start": "2025-01-14T14:30:00Z",
  "window_end": "2025-01-15T14:30:00Z"
}
```

### Policy Management

#### List Policies
```http
GET /api/v1/retry/policies
```

**Response:**
```json
{
  "policies": [
    {
      "name": "rate_limit",
      "error_patterns": ["429", "rate_limit"],
      "max_attempts": 5,
      "base_delay_ms": 5000,
      "max_delay_ms": 60000,
      "backoff_multiplier": 2.0,
      "jitter_percent": 25.0,
      "priority": 100
    }
  ]
}
```

#### Add Policy
```http
POST /api/v1/retry/policies
Content-Type: application/json

{
  "name": "custom_timeout",
  "error_patterns": ["timeout", "connection_timeout"],
  "job_type_patterns": ["critical_.*"],
  "max_attempts": 3,
  "base_delay_ms": 2000,
  "max_delay_ms": 30000,
  "backoff_multiplier": 2.0,
  "jitter_percent": 20.0,
  "priority": 90
}
```

#### Remove Policy
```http
DELETE /api/v1/retry/policies/custom_timeout
```

### Model Management

#### Update Bayesian Model
```http
POST /api/v1/retry/bayesian/update
Content-Type: application/json

{
  "job_type": "email_send",
  "error_class": "timeout"
}
```

#### Train ML Model
```http
POST /api/v1/retry/ml/train
Content-Type: application/json

{
  "model_type": "logistic",
  "features": [
    "attempt_number",
    "payload_size",
    "time_of_day",
    "since_last_failure_ms"
  ],
  "training_period": "720h",
  "validation_set": 0.2,
  "cross_validation": 5,
  "hyperparameters": {
    "learning_rate": 0.01,
    "regularization": 0.1
  }
}
```

**Response:**
```json
{
  "version": "v1642680000",
  "model_type": "logistic",
  "features": ["attempt_number", "payload_size"],
  "trained_at": "2025-01-15T14:30:00Z",
  "accuracy": 0.87,
  "f1_score": 0.84,
  "validation_set": "20% of 5000 samples",
  "enabled": false,
  "metadata": {
    "training_samples": 4000,
    "validation_samples": 1000,
    "feature_count": 4
  }
}
```

#### Deploy ML Model
```http
POST /api/v1/retry/ml/deploy
Content-Type: application/json

{
  "model": {
    "version": "v1642680000",
    "model_type": "logistic"
  },
  "canary_percent": 10.0
}
```

#### Rollback ML Model
```http
POST /api/v1/retry/ml/rollback
```

### Configuration

#### Get Strategy
```http
GET /api/v1/retry/strategy
```

**Response:**
```json
{
  "name": "production",
  "enabled": true,
  "policies": [...],
  "bayesian_threshold": 0.70,
  "ml_enabled": true,
  "guardrails": {
    "max_attempts": 10,
    "max_delay_ms": 300000,
    "max_budget_percent": 20.0,
    "per_tenant_limits": true,
    "emergency_stop": false,
    "explainability_required": true
  },
  "data_collection": {
    "enabled": true,
    "sample_rate": 1.0,
    "retention_days": 30,
    "aggregation_interval": "5m",
    "feature_extraction": true
  }
}
```

#### Update Guardrails
```http
PUT /api/v1/retry/guardrails
Content-Type: application/json

{
  "max_attempts": 8,
  "max_delay_ms": 180000,
  "max_budget_percent": 15.0,
  "per_tenant_limits": true,
  "emergency_stop": false,
  "explainability_required": true
}
```

## Default Policies

The system ships with sensible default policies:

### Rate Limiting Policy
- **Patterns:** `429`, `rate_limit`, `too_many_requests`
- **Strategy:** Exponential backoff with high jitter
- **Max Attempts:** 5
- **Base Delay:** 5 seconds
- **Max Delay:** 60 seconds
- **Jitter:** 25%

### Service Unavailable Policy
- **Patterns:** `503`, `service_unavailable`, `timeout`
- **Strategy:** Moderate exponential backoff
- **Max Attempts:** 3
- **Base Delay:** 2 seconds
- **Max Delay:** 30 seconds
- **Jitter:** 20%

### Validation Error Policy
- **Patterns:** `400`, `validation`, `invalid_input`
- **Strategy:** No retry (fail fast)
- **Max Attempts:** 1
- **Stop on Validation:** true

### Default Fallback Policy
- **Patterns:** (matches all)
- **Strategy:** Standard exponential backoff
- **Max Attempts:** 3
- **Base Delay:** 1 second
- **Max Delay:** 30 seconds
- **Jitter:** 15%

## Bayesian Learning

The Bayesian recommender uses Beta distributions to model success probability over delay buckets:

### Delay Buckets
- 0-1s, 1-5s, 5-15s, 15-30s, 30-60s, 1-5m, 5-15m, 15m+

### Model Updates
- Automatic updates when sufficient new data is available
- Minimum 10 samples required for initial model
- Updates triggered after every 20 new samples

### Confidence Calculation
- Based on sample size and confidence interval width
- Penalized for insufficient data (< 50 samples)
- Used to determine recommendation fallback

## Machine Learning

### Supported Models
- **Logistic Regression:** Fast, interpretable binary classification
- **Gradient Boosting:** More complex patterns, higher accuracy

### Feature Engineering
- Categorical encoding via hash functions
- Temporal features (time of day, since last failure)
- System health metrics integration
- Cross-validation for hyperparameter tuning

### Deployment Strategy
- **Canary Testing:** Gradual rollout with percentage-based traffic
- **A/B Testing:** Compare ML vs Bayesian recommendations
- **Rollback:** Instant rollback to previous model or Bayesian layer

## Safety Guardrails

### Hard Limits
- **Max Attempts:** Absolute maximum retry attempts
- **Max Delay:** Maximum delay between retries
- **Budget Limits:** Percentage of total processing time for retries
- **Emergency Stop:** Manual override to disable retries

### Explainability
- All recommendations include rationale
- Decision audit trail maintained
- Feature importance tracking for ML models

## Integration Examples

### Go Client
```go
package main

import (
    "context"
    smartretry "github.com/flyingrobots/go-redis-work-queue/internal/smart-retry-strategies"
    "go.uber.org/zap"
)

func main() {
    config := smartretry.DefaultConfig()
    config.RedisAddr = "localhost:6379"

    manager, err := smartretry.NewManager(config, zap.NewExample())
    if err != nil {
        panic(err)
    }
    defer manager.Close()

    // Get recommendation
    features := smartretry.RetryFeatures{
        JobType:       "email_send",
        ErrorClass:    "timeout",
        AttemptNumber: 2,
        Queue:         "default",
    }

    rec, err := manager.GetRecommendation(features)
    if err != nil {
        panic(err)
    }

    if rec.ShouldRetry {
        // Wait for recommended delay
        time.Sleep(time.Duration(rec.DelayMs) * time.Millisecond)

        // Retry the job
        // ... retry logic
    }

    // Record the attempt outcome
    attempt := smartretry.AttemptHistory{
        JobID:         "job_123",
        JobType:       features.JobType,
        AttemptNumber: features.AttemptNumber,
        ErrorClass:    features.ErrorClass,
        Success:       true, // or false
        DelayMs:       rec.DelayMs,
        Timestamp:     time.Now(),
    }

    manager.RecordAttempt(attempt)
}
```

### HTTP Client (JavaScript)
```javascript
const retryClient = {
  async getRecommendation(features) {
    const response = await fetch('/api/v1/retry/recommendation', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(features)
    });
    return response.json();
  },

  async recordAttempt(attempt) {
    await fetch('/api/v1/retry/attempt', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(attempt)
    });
  }
};

// Usage
const features = {
  job_type: 'email_send',
  error_class: 'timeout',
  attempt_number: 2
};

const recommendation = await retryClient.getRecommendation(features);

if (recommendation.should_retry) {
  await new Promise(resolve =>
    setTimeout(resolve, recommendation.delay_ms));

  // Retry the operation
  const success = await retryOperation();

  // Record the attempt
  await retryClient.recordAttempt({
    job_id: 'job_123',
    job_type: features.job_type,
    attempt_number: features.attempt_number,
    success: success,
    delay_ms: recommendation.delay_ms,
    timestamp: new Date().toISOString()
  });
}
```

## Configuration

### Environment Variables
- `RETRY_REDIS_ADDR` - Redis connection address
- `RETRY_REDIS_PASSWORD` - Redis password
- `RETRY_ENABLED` - Enable/disable retry strategies
- `RETRY_SAMPLE_RATE` - Data collection sample rate (0.0-1.0)
- `RETRY_ML_ENABLED` - Enable ML models
- `RETRY_BAYESIAN_THRESHOLD` - Minimum confidence for Bayesian recommendations

### Configuration File Example
```json
{
  "enabled": true,
  "redis_addr": "localhost:6379",
  "redis_password": "",
  "redis_db": 0,
  "strategy": {
    "name": "production",
    "enabled": true,
    "bayesian_threshold": 0.70,
    "ml_enabled": true,
    "guardrails": {
      "max_attempts": 10,
      "max_delay_ms": 300000,
      "max_budget_percent": 20.0,
      "per_tenant_limits": true,
      "emergency_stop": false,
      "explainability_required": true
    },
    "data_collection": {
      "enabled": true,
      "sample_rate": 1.0,
      "retention_days": 30,
      "aggregation_interval": "5m",
      "feature_extraction": true
    }
  },
  "cache": {
    "enabled": true,
    "ttl": "5m",
    "max_entries": 1000
  },
  "api": {
    "enabled": true,
    "port": 8080,
    "path": "/api/v1/retry"
  }
}
```

## Monitoring & Observability

### Health Check
```http
GET /api/v1/retry/health
```

### Metrics Endpoint
```http
GET /api/v1/retry/metrics
```

### Key Metrics
- Recommendation accuracy
- Model performance (accuracy, F1 score)
- Cache hit rates
- Policy match rates
- Guardrail trigger frequency
- Data collection rates

## Error Handling

All API endpoints return structured error responses:

```json
{
  "error": "Invalid request body",
  "status": 400,
  "timestamp": "2025-01-15T14:30:00Z",
  "details": "missing required field: job_type"
}
```

### Error Categories
- **Configuration Errors** (400): Invalid config or parameters
- **Model Errors** (404/500): Model not found or training failed
- **Data Errors** (500): Storage or retrieval failures
- **Guardrail Errors** (429): Safety limits exceeded

## Best Practices

1. **Start Simple:** Begin with rule-based policies before enabling ML
2. **Collect Data:** Ensure comprehensive attempt history collection
3. **Monitor Performance:** Track recommendation accuracy and system impact
4. **Gradual Rollout:** Use canary deployments for new models
5. **Safety First:** Configure appropriate guardrails for your use case
6. **Regular Updates:** Retrain models periodically with fresh data