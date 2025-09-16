// Copyright 2025 James Ross
package forecasting

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MetricsStorage handles persistence of time series data
type MetricsStorage struct {
	config  *StorageConfig
	logger  *zap.Logger
	data    map[string]*TimeSeries
	mu      sync.RWMutex

	// Background persistence
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewMetricsStorage creates a new metrics storage
func NewMetricsStorage(config *StorageConfig, logger *zap.Logger) *MetricsStorage {
	if config == nil {
		config = &StorageConfig{
			RetentionDuration: 7 * 24 * time.Hour,
			SamplingInterval:  1 * time.Minute,
			MaxDataPoints:     10080, // 7 days of minute data
			PersistToDisk:     false,
			StoragePath:       "/tmp/forecasting",
		}
	}

	ms := &MetricsStorage{
		config:   config,
		logger:   logger,
		data:     make(map[string]*TimeSeries),
		stopChan: make(chan struct{}),
	}

	// Create storage directory if persistence is enabled
	if config.PersistToDisk {
		if err := os.MkdirAll(config.StoragePath, 0755); err != nil {
			logger.Warn("Failed to create storage directory",
				zap.String("path", config.StoragePath),
				zap.Error(err))
		}

		// Load existing data
		ms.loadFromDisk()

		// Start background persistence
		ms.wg.Add(1)
		go ms.persistenceLoop()
	}

	// Start cleanup routine
	ms.wg.Add(1)
	go ms.cleanupLoop()

	return ms
}

// Store stores a metric value
func (ms *MetricsStorage) Store(metricType MetricType, queueName string, value float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := ms.makeKey(metricType, queueName)

	// Get or create time series
	ts, exists := ms.data[key]
	if !exists {
		ts = &TimeSeries{
			Name:       key,
			MetricType: metricType,
			Points:     make([]DataPoint, 0, ms.config.MaxDataPoints),
		}
		ms.data[key] = ts
	}

	// Add data point
	ts.mu.Lock()
	ts.Points = append(ts.Points, DataPoint{
		Timestamp: time.Now(),
		Value:     value,
	})

	// Limit data points
	if len(ts.Points) > ms.config.MaxDataPoints {
		ts.Points = ts.Points[len(ts.Points)-ms.config.MaxDataPoints:]
	}
	ts.mu.Unlock()
}

// GetTimeSeries retrieves time series data
func (ms *MetricsStorage) GetTimeSeries(metricType MetricType, queueName string, duration time.Duration) []DataPoint {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	key := ms.makeKey(metricType, queueName)
	ts, exists := ms.data[key]
	if !exists {
		return nil
	}

	ts.mu.RLock()
	defer ts.mu.RUnlock()

	cutoff := time.Now().Add(-duration)
	var result []DataPoint

	for _, point := range ts.Points {
		if point.Timestamp.After(cutoff) {
			result = append(result, point)
		}
	}

	return result
}

// GetLatest gets the most recent value for a metric
func (ms *MetricsStorage) GetLatest(metricType MetricType, queueName string) (float64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	key := ms.makeKey(metricType, queueName)
	ts, exists := ms.data[key]
	if !exists {
		return 0, false
	}

	ts.mu.RLock()
	defer ts.mu.RUnlock()

	if len(ts.Points) == 0 {
		return 0, false
	}

	return ts.Points[len(ts.Points)-1].Value, true
}

// GetAggregated returns aggregated data points
func (ms *MetricsStorage) GetAggregated(metricType MetricType, queueName string, duration time.Duration, aggregation string) []DataPoint {
	rawData := ms.GetTimeSeries(metricType, queueName, duration)
	if len(rawData) == 0 {
		return nil
	}

	switch aggregation {
	case "1m":
		return ms.aggregateByMinute(rawData)
	case "5m":
		return ms.aggregateByInterval(rawData, 5*time.Minute)
	case "1h":
		return ms.aggregateByInterval(rawData, 1*time.Hour)
	case "1d":
		return ms.aggregateByInterval(rawData, 24*time.Hour)
	default:
		return rawData
	}
}

// Stop stops the storage system
func (ms *MetricsStorage) Stop() {
	close(ms.stopChan)
	ms.wg.Wait()

	// Final save
	if ms.config.PersistToDisk {
		ms.saveToDisk()
	}
}

// Helper methods

func (ms *MetricsStorage) makeKey(metricType MetricType, queueName string) string {
	if queueName == "" {
		return string(metricType)
	}
	return fmt.Sprintf("%s:%s", metricType, queueName)
}

func (ms *MetricsStorage) cleanupLoop() {
	defer ms.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ms.stopChan:
			return
		case <-ticker.C:
			ms.cleanup()
		}
	}
}

func (ms *MetricsStorage) cleanup() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	cutoff := time.Now().Add(-ms.config.RetentionDuration)

	for key, ts := range ms.data {
		ts.mu.Lock()

		// Remove old data points
		var kept []DataPoint
		for _, point := range ts.Points {
			if point.Timestamp.After(cutoff) {
				kept = append(kept, point)
			}
		}

		if len(kept) == 0 {
			// Remove empty time series
			delete(ms.data, key)
		} else {
			ts.Points = kept
		}

		ts.mu.Unlock()
	}
}

func (ms *MetricsStorage) persistenceLoop() {
	defer ms.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ms.stopChan:
			return
		case <-ticker.C:
			ms.saveToDisk()
		}
	}
}

func (ms *MetricsStorage) saveToDisk() {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	for key, ts := range ms.data {
		filename := filepath.Join(ms.config.StoragePath, fmt.Sprintf("%s.json", key))

		ts.mu.RLock()
		data, err := json.Marshal(ts)
		ts.mu.RUnlock()

		if err != nil {
			ms.logger.Warn("Failed to marshal time series",
				zap.String("key", key),
				zap.Error(err))
			continue
		}

		if err := ioutil.WriteFile(filename, data, 0644); err != nil {
			ms.logger.Warn("Failed to save time series",
				zap.String("file", filename),
				zap.Error(err))
		}
	}
}

func (ms *MetricsStorage) loadFromDisk() {
	files, err := ioutil.ReadDir(ms.config.StoragePath)
	if err != nil {
		ms.logger.Warn("Failed to read storage directory",
			zap.String("path", ms.config.StoragePath),
			zap.Error(err))
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filepath := filepath.Join(ms.config.StoragePath, file.Name())
		data, err := ioutil.ReadFile(filepath)
		if err != nil {
			ms.logger.Warn("Failed to read file",
				zap.String("file", filepath),
				zap.Error(err))
			continue
		}

		var ts TimeSeries
		if err := json.Unmarshal(data, &ts); err != nil {
			ms.logger.Warn("Failed to unmarshal time series",
				zap.String("file", filepath),
				zap.Error(err))
			continue
		}

		ms.data[ts.Name] = &ts
	}

	ms.logger.Info("Loaded time series from disk",
		zap.Int("count", len(ms.data)))
}

func (ms *MetricsStorage) aggregateByMinute(data []DataPoint) []DataPoint {
	if len(data) == 0 {
		return nil
	}

	aggregated := make([]DataPoint, 0)
	currentMinute := data[0].Timestamp.Truncate(time.Minute)
	sum := 0.0
	count := 0

	for _, point := range data {
		pointMinute := point.Timestamp.Truncate(time.Minute)

		if pointMinute.Equal(currentMinute) {
			sum += point.Value
			count++
		} else {
			// Save aggregated point
			if count > 0 {
				aggregated = append(aggregated, DataPoint{
					Timestamp: currentMinute,
					Value:     sum / float64(count),
				})
			}

			// Start new aggregation
			currentMinute = pointMinute
			sum = point.Value
			count = 1
		}
	}

	// Save last aggregated point
	if count > 0 {
		aggregated = append(aggregated, DataPoint{
			Timestamp: currentMinute,
			Value:     sum / float64(count),
		})
	}

	return aggregated
}

func (ms *MetricsStorage) aggregateByInterval(data []DataPoint, interval time.Duration) []DataPoint {
	if len(data) == 0 {
		return nil
	}

	aggregated := make([]DataPoint, 0)
	currentInterval := data[0].Timestamp.Truncate(interval)
	sum := 0.0
	count := 0

	for _, point := range data {
		pointInterval := point.Timestamp.Truncate(interval)

		if pointInterval.Equal(currentInterval) {
			sum += point.Value
			count++
		} else {
			// Save aggregated point
			if count > 0 {
				aggregated = append(aggregated, DataPoint{
					Timestamp: currentInterval,
					Value:     sum / float64(count),
				})
			}

			// Start new aggregation
			currentInterval = pointInterval
			sum = point.Value
			count = 1
		}
	}

	// Save last aggregated point
	if count > 0 {
		aggregated = append(aggregated, DataPoint{
			Timestamp: currentInterval,
			Value:     sum / float64(count),
		})
	}

	return aggregated
}
