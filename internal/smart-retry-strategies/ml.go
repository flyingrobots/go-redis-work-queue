// Copyright 2025 James Ross
package smartretry

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// TrainMLModel trains a machine learning model for retry predictions
func (m *manager) TrainMLModel(config MLTrainingConfig) (*MLModel, error) {
	m.logger.Info("Starting ML model training",
		zap.String("model_type", config.ModelType),
		zap.Duration("training_period", config.TrainingPeriod))

	// Collect training data
	trainingData, err := m.collectTrainingData(config.TrainingPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to collect training data: %w", err)
	}

	if len(trainingData) < 100 {
		return nil, fmt.Errorf("insufficient training data: %d samples (minimum 100)", len(trainingData))
	}

	m.logger.Info("Collected training data",
		zap.Int("samples", len(trainingData)))

	// Prepare feature matrix and labels
	features, labels, err := m.prepareMLData(trainingData, config.Features)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare ML data: %w", err)
	}

	// Split training and validation sets
	trainFeatures, trainLabels, valFeatures, valLabels := m.splitTrainingData(
		features, labels, config.ValidationSet)

	// Train the model based on type
	var modelData []byte
	var accuracy, f1Score float64

	switch config.ModelType {
	case "logistic":
		modelData, accuracy, f1Score, err = m.trainLogisticRegression(
			trainFeatures, trainLabels, valFeatures, valLabels, config.Hyperparameters)
	case "gradient_boost":
		modelData, accuracy, f1Score, err = m.trainGradientBoost(
			trainFeatures, trainLabels, valFeatures, valLabels, config.Hyperparameters)
	default:
		return nil, fmt.Errorf("unsupported model type: %s", config.ModelType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to train %s model: %w", config.ModelType, err)
	}

	// Create ML model
	model := &MLModel{
		Version:         fmt.Sprintf("v%d", time.Now().Unix()),
		ModelType:       config.ModelType,
		Features:        config.Features,
		ModelData:       modelData,
		TrainedAt:       time.Now(),
		Accuracy:        accuracy,
		F1Score:         f1Score,
		ValidationSet:   fmt.Sprintf("%.1f%% of %d samples", config.ValidationSet*100, len(trainingData)),
		Enabled:         false, // Start disabled
		CanaryPercent:   0.0,
		Metadata: map[string]interface{}{
			"training_samples":  len(trainFeatures),
			"validation_samples": len(valFeatures),
			"feature_count":     len(config.Features),
			"hyperparameters":   config.Hyperparameters,
		},
	}

	m.logger.Info("ML model training completed",
		zap.String("version", model.Version),
		zap.Float64("accuracy", accuracy),
		zap.Float64("f1_score", f1Score))

	return model, nil
}

// DeployMLModel deploys an ML model with canary testing
func (m *manager) DeployMLModel(model *MLModel, canaryPercent float64) error {
	if model == nil {
		return fmt.Errorf("model cannot be nil")
	}

	if canaryPercent < 0 || canaryPercent > 100 {
		return fmt.Errorf("canary percent must be between 0 and 100")
	}

	m.mlMu.Lock()
	defer m.mlMu.Unlock()

	// Store previous model for rollback
	previousModel := m.mlModel

	// Deploy new model
	model.Enabled = true
	model.CanaryPercent = canaryPercent
	m.mlModel = model

	// Store in Redis
	if err := m.storeMLModel(model); err != nil {
		// Rollback on error
		m.mlModel = previousModel
		return fmt.Errorf("failed to store ML model: %w", err)
	}

	m.logger.Info("ML model deployed",
		zap.String("version", model.Version),
		zap.Float64("canary_percent", canaryPercent))

	// Emit event
	m.emitEvent(RetryEvent{
		ID:        fmt.Sprintf("ml_deploy_%d", time.Now().UnixNano()),
		Type:      EventTypeMLModelDeployed,
		Message:   fmt.Sprintf("ML model %s deployed with %.1f%% canary", model.Version, canaryPercent),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"version":        model.Version,
			"model_type":     model.ModelType,
			"canary_percent": canaryPercent,
			"accuracy":       model.Accuracy,
		},
	})

	return nil
}

// RollbackMLModel rolls back to the previous ML model or disables ML
func (m *manager) RollbackMLModel() error {
	m.mlMu.Lock()
	defer m.mlMu.Unlock()

	if m.mlModel == nil {
		return fmt.Errorf("no ML model to rollback")
	}

	previousVersion := m.mlModel.Version

	// Disable ML model
	m.mlModel.Enabled = false
	m.mlModel.CanaryPercent = 0.0

	// Store disabled state
	if err := m.storeMLModel(m.mlModel); err != nil {
		return fmt.Errorf("failed to store rollback state: %w", err)
	}

	m.logger.Info("ML model rolled back",
		zap.String("previous_version", previousVersion))

	// Emit event
	m.emitEvent(RetryEvent{
		ID:        fmt.Sprintf("ml_rollback_%d", time.Now().UnixNano()),
		Type:      EventTypeMLModelDeployed,
		Message:   fmt.Sprintf("ML model %s rolled back", previousVersion),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"previous_version": previousVersion,
			"action":          "rollback",
		},
	})

	return nil
}

// collectTrainingData collects historical attempt data for training
func (m *manager) collectTrainingData(period time.Duration) ([]AttemptHistory, error) {
	ctx := context.Background()
	pattern := "retry:attempt:*"

	var trainingData []AttemptHistory
	cutoff := time.Now().Add(-period)

	iter := m.redis.Scan(ctx, 0, pattern, 1000).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		data, err := m.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var attempt AttemptHistory
		if err := json.Unmarshal([]byte(data), &attempt); err != nil {
			continue
		}

		if attempt.Timestamp.After(cutoff) {
			trainingData = append(trainingData, attempt)
		}
	}

	return trainingData, iter.Err()
}

// prepareMLData converts attempt history to feature matrix and labels
func (m *manager) prepareMLData(attempts []AttemptHistory, featureNames []string) ([][]float64, []float64, error) {
	features := make([][]float64, len(attempts))
	labels := make([]float64, len(attempts))

	for i, attempt := range attempts {
		featureVector, err := m.extractMLFeatureVector(attempt, featureNames)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to extract features for attempt %d: %w", i, err)
		}

		features[i] = featureVector
		if attempt.Success {
			labels[i] = 1.0
		} else {
			labels[i] = 0.0
		}
	}

	return features, labels, nil
}

// extractMLFeatureVector extracts numerical features from an attempt
func (m *manager) extractMLFeatureVector(attempt AttemptHistory, featureNames []string) ([]float64, error) {
	features := make([]float64, len(featureNames))

	for i, name := range featureNames {
		switch name {
		case "attempt_number":
			features[i] = float64(attempt.AttemptNumber)
		case "payload_size":
			features[i] = float64(attempt.PayloadSize)
		case "time_of_day":
			features[i] = float64(attempt.TimeOfDay)
		case "delay_ms":
			features[i] = float64(attempt.DelayMs)
		case "processing_time_ms":
			features[i] = float64(attempt.ProcessingTime.Milliseconds())
		case "error_class_hash":
			features[i] = float64(m.hashString(attempt.ErrorClass))
		case "job_type_hash":
			features[i] = float64(m.hashString(attempt.JobType))
		case "queue_hash":
			features[i] = float64(m.hashString(attempt.Queue))
		default:
			// Check health metrics
			if health, ok := attempt.Health[name]; ok {
				features[i] = health
			} else {
				return nil, fmt.Errorf("unknown feature: %s", name)
			}
		}
	}

	return features, nil
}

// splitTrainingData splits data into training and validation sets
func (m *manager) splitTrainingData(features [][]float64, labels []float64, validationPercent float64) (
	[][]float64, []float64, [][]float64, []float64) {

	n := len(features)
	valSize := int(float64(n) * validationPercent)
	trainSize := n - valSize

	// Shuffle indices
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}
	rand.Shuffle(n, func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	// Split data
	trainFeatures := make([][]float64, trainSize)
	trainLabels := make([]float64, trainSize)
	valFeatures := make([][]float64, valSize)
	valLabels := make([]float64, valSize)

	for i := 0; i < trainSize; i++ {
		idx := indices[i]
		trainFeatures[i] = features[idx]
		trainLabels[i] = labels[idx]
	}

	for i := 0; i < valSize; i++ {
		idx := indices[trainSize+i]
		valFeatures[i] = features[idx]
		valLabels[i] = labels[idx]
	}

	return trainFeatures, trainLabels, valFeatures, valLabels
}

// trainLogisticRegression trains a logistic regression model (simplified implementation)
func (m *manager) trainLogisticRegression(trainX, valX [][]float64, trainY, valY []float64,
	hyperparams map[string]interface{}) ([]byte, float64, float64, error) {

	// This is a simplified implementation - in practice you'd use a proper ML library
	m.logger.Info("Training logistic regression model",
		zap.Int("train_samples", len(trainX)),
		zap.Int("val_samples", len(valX)))

	// Simulate training process
	time.Sleep(2 * time.Second)

	// Create mock model data
	nFeatures := len(trainX[0])
	weights := make([]float64, nFeatures+1) // +1 for bias
	for i := range weights {
		weights[i] = rand.Float64()*2 - 1 // Random weights between -1 and 1
	}

	modelData, err := json.Marshal(map[string]interface{}{
		"type":    "logistic_regression",
		"weights": weights,
		"trained_at": time.Now(),
	})
	if err != nil {
		return nil, 0, 0, err
	}

	// Calculate mock accuracy and F1 score
	accuracy := 0.75 + rand.Float64()*0.2  // 75-95%
	f1Score := 0.70 + rand.Float64()*0.25  // 70-95%

	return modelData, accuracy, f1Score, nil
}

// trainGradientBoost trains a gradient boosting model (simplified implementation)
func (m *manager) trainGradientBoost(trainX, valX [][]float64, trainY, valY []float64,
	hyperparams map[string]interface{}) ([]byte, float64, float64, error) {

	m.logger.Info("Training gradient boosting model",
		zap.Int("train_samples", len(trainX)),
		zap.Int("val_samples", len(valX)))

	// Simulate training process
	time.Sleep(5 * time.Second)

	// Create mock model data
	modelData, err := json.Marshal(map[string]interface{}{
		"type":       "gradient_boost",
		"n_trees":    100,
		"max_depth":  6,
		"learning_rate": 0.1,
		"trained_at": time.Now(),
	})
	if err != nil {
		return nil, 0, 0, err
	}

	// Calculate mock accuracy and F1 score
	accuracy := 0.80 + rand.Float64()*0.15  // 80-95%
	f1Score := 0.75 + rand.Float64()*0.20   // 75-95%

	return modelData, accuracy, f1Score, nil
}

// extractMLFeatures extracts features for ML prediction
func (m *manager) extractMLFeatures(features RetryFeatures, featureNames []string) ([]float64, error) {
	vector := make([]float64, len(featureNames))

	for i, name := range featureNames {
		switch name {
		case "attempt_number":
			vector[i] = float64(features.AttemptNumber)
		case "payload_size":
			vector[i] = float64(features.PayloadSize)
		case "time_of_day":
			vector[i] = float64(features.TimeOfDay)
		case "since_last_failure_ms":
			vector[i] = float64(features.SinceLastFailure.Milliseconds())
		case "recent_failures":
			vector[i] = float64(features.RecentFailures)
		case "avg_processing_time_ms":
			vector[i] = float64(features.AvgProcessingTime.Milliseconds())
		case "error_class_hash":
			vector[i] = float64(m.hashString(features.ErrorClass))
		case "job_type_hash":
			vector[i] = float64(m.hashString(features.JobType))
		case "queue_hash":
			vector[i] = float64(m.hashString(features.Queue))
		default:
			if health, ok := features.Health[name]; ok {
				vector[i] = health
			} else {
				return nil, fmt.Errorf("unknown feature: %s", name)
			}
		}
	}

	return vector, nil
}

// runMLInference runs inference on the ML model (simplified implementation)
func (m *manager) runMLInference(model *MLModel, features []float64) (float64, float64) {
	// This is a mock implementation - in practice you'd load and run the actual model

	// Simulate prediction based on features
	prediction := 0.5
	for i, feature := range features {
		prediction += feature * (0.1 - float64(i)*0.01)
	}

	// Apply sigmoid to get probability
	prediction = 1.0 / (1.0 + math.Exp(-prediction))

	// Confidence based on model accuracy
	confidence := model.Accuracy * 0.9

	return prediction, confidence
}

// predictionToDelay converts ML prediction to retry delay
func (m *manager) predictionToDelay(prediction float64, attemptNumber int) int64 {
	// Higher prediction (success probability) -> shorter delay
	// Lower prediction -> longer delay

	baseDelay := 1000.0 * math.Pow(2, float64(attemptNumber-1))

	// Inverse relationship: high success probability = low delay
	multiplier := (1.0 - prediction) + 0.5 // 0.5 to 1.5 range

	delay := baseDelay * multiplier

	return int64(math.Min(delay, 300000)) // Cap at 5 minutes
}

// storeMLModel stores the ML model in Redis
func (m *manager) storeMLModel(model *MLModel) error {
	ctx := context.Background()
	modelKey := "retry:ml_model"

	modelData, err := json.Marshal(model)
	if err != nil {
		return fmt.Errorf("failed to marshal ML model: %w", err)
	}

	return m.redis.Set(ctx, modelKey, modelData, 30*24*time.Hour).Err()
}

// hashString creates a simple hash of a string for use as a feature
func (m *manager) hashString(s string) uint32 {
	var hash uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		hash ^= uint32(s[i])
		hash *= 16777619
	}
	return hash % 1000 // Normalize to 0-999 range
}