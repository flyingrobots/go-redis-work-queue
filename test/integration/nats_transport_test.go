// Copyright 2025 James Ross
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockNATSServer simulates a NATS server for testing
type MockNATSServer struct {
	subjects     map[string][]MockMessage
	subscribers  map[string][]MockSubscription
	isConnected  bool
	mu           sync.RWMutex
}

// MockMessage represents a NATS message
type MockMessage struct {
	Subject   string
	Data      []byte
	Headers   map[string]string
	Timestamp time.Time
	ReplyTo   string
}

// MockSubscription represents a NATS subscription
type MockSubscription struct {
	Subject string
	Handler func(*MockMessage)
	Active  bool
}

// NATSTransport handles NATS-based event transport
type NATSTransport struct {
	server    *MockNATSServer
	publisher *NATSPublisher
	config    NATSConfig
}

// NATSPublisher publishes events to NATS subjects
type NATSPublisher struct {
	server *MockNATSServer
	config NATSConfig
}

// NATSConfig defines NATS transport configuration
type NATSConfig struct {
	URL             string
	SubjectPrefix   string
	MaxReconnect    int
	ReconnectWait   time.Duration
	Timeout         time.Duration
	MaxPubAcksInflight int
	EnableJetStream bool
}

// EventMessage represents an event published to NATS
type EventMessage struct {
	ID        string                 `json:"id"`
	Subject   string                 `json:"subject"`
	Event     string                 `json:"event"`
	Timestamp time.Time              `json:"timestamp"`
	JobID     string                 `json:"job_id"`
	Queue     string                 `json:"queue"`
	Priority  int                    `json:"priority"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

// SubjectPattern defines NATS subject naming patterns
type SubjectPattern struct {
	Prefix    string
	Event     string
	Queue     string
	Priority  string
}

// NewMockNATSServer creates a new mock NATS server
func NewMockNATSServer() *MockNATSServer {
	return &MockNATSServer{
		subjects:    make(map[string][]MockMessage),
		subscribers: make(map[string][]MockSubscription),
		isConnected: true,
	}
}

// Connect simulates connecting to NATS
func (s *MockNATSServer) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isConnected = true
	return nil
}

// Disconnect simulates disconnecting from NATS
func (s *MockNATSServer) Disconnect() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isConnected = false
}

// IsConnected returns connection status
func (s *MockNATSServer) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isConnected
}

// Publish publishes a message to a subject
func (s *MockNATSServer) Publish(subject string, data []byte) error {
	return s.PublishWithHeaders(subject, data, nil)
}

// PublishWithHeaders publishes a message with headers
func (s *MockNATSServer) PublishWithHeaders(subject string, data []byte, headers map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isConnected {
		return fmt.Errorf("not connected to NATS")
	}

	message := MockMessage{
		Subject:   subject,
		Data:      data,
		Headers:   headers,
		Timestamp: time.Now(),
	}

	s.subjects[subject] = append(s.subjects[subject], message)

	// Notify subscribers
	if subs, exists := s.subscribers[subject]; exists {
		for _, sub := range subs {
			if sub.Active {
				go sub.Handler(&message)
			}
		}
	}

	return nil
}

// Subscribe subscribes to a subject
func (s *MockNATSServer) Subscribe(subject string, handler func(*MockMessage)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	subscription := MockSubscription{
		Subject: subject,
		Handler: handler,
		Active:  true,
	}

	s.subscribers[subject] = append(s.subscribers[subject], subscription)
	return nil
}

// GetMessages returns all messages for a subject
func (s *MockNATSServer) GetMessages(subject string) []MockMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	messages := make([]MockMessage, len(s.subjects[subject]))
	copy(messages, s.subjects[subject])
	return messages
}

// GetSubjectCount returns the number of subjects with messages
func (s *MockNATSServer) GetSubjectCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.subjects)
}

// ClearMessages clears all messages
func (s *MockNATSServer) ClearMessages() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subjects = make(map[string][]MockMessage)
}

// NewNATSTransport creates a new NATS transport
func NewNATSTransport(server *MockNATSServer, config NATSConfig) *NATSTransport {
	return &NATSTransport{
		server: server,
		publisher: &NATSPublisher{
			server: server,
			config: config,
		},
		config: config,
	}
}

// PublishEvent publishes an event to the appropriate NATS subject
func (t *NATSTransport) PublishEvent(ctx context.Context, event EventMessage) error {
	if !t.server.IsConnected() {
		return fmt.Errorf("NATS server not connected")
	}

	// Generate subject based on event
	subject := t.GenerateSubject(event)
	event.Subject = subject

	// Serialize event
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Prepare headers
	headers := map[string]string{
		"Event-ID":    event.ID,
		"Event-Type":  event.Event,
		"Job-ID":      event.JobID,
		"Queue":       event.Queue,
		"Content-Type": "application/json",
	}

	if event.TraceID != "" {
		headers["Trace-ID"] = event.TraceID
	}

	if event.RequestID != "" {
		headers["Request-ID"] = event.RequestID
	}

	// Publish to NATS
	return t.server.PublishWithHeaders(subject, data, headers)
}

// GenerateSubject creates NATS subject based on event details
func (t *NATSTransport) GenerateSubject(event EventMessage) string {
	// Pattern: events.{queue}.{event_type}.{priority}
	if t.config.SubjectPrefix == "" {
		t.config.SubjectPrefix = "events"
	}

	priority := "normal"
	if event.Priority >= 8 {
		priority = "high"
	} else if event.Priority <= 3 {
		priority = "low"
	}

	return fmt.Sprintf("%s.%s.%s.%s",
		t.config.SubjectPrefix,
		event.Queue,
		event.Event,
		priority)
}

// SubscribeToPattern subscribes to events matching a pattern
func (t *NATSTransport) SubscribeToPattern(ctx context.Context, pattern string, handler func(EventMessage) error) error {
	return t.server.Subscribe(pattern, func(msg *MockMessage) {
		var event EventMessage
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			return
		}
		handler(event)
	})
}

// GetSubjectPatterns returns common subject patterns for subscription
func (t *NATSTransport) GetSubjectPatterns() []string {
	prefix := t.config.SubjectPrefix
	if prefix == "" {
		prefix = "events"
	}

	return []string{
		prefix + ".*",                    // All events
		prefix + ".*.job_failed.*",       // All job failures
		prefix + ".*.job_dlq.*",         // All DLQ events
		prefix + ".*.*",                 // All events (explicit wildcard)
		prefix + ".priority_queue.*.*",   // All priority queue events
		prefix + ".*.*.high",            // All high priority events
	}
}

// Integration Tests

func TestNATSTransport_BasicEventPublishing(t *testing.T) {
	server := NewMockNATSServer()
	config := NATSConfig{
		SubjectPrefix:   "test.events",
		MaxReconnect:    5,
		ReconnectWait:   time.Second,
		Timeout:         10 * time.Second,
	}

	transport := NewNATSTransport(server, config)

	ctx := context.Background()
	event := EventMessage{
		ID:        "evt_001",
		Event:     "job_failed",
		Timestamp: time.Now(),
		JobID:     "job_123",
		Queue:     "user_queue",
		Priority:  5,
		Data: map[string]interface{}{
			"error":    "Connection timeout",
			"duration": "30s",
		},
		TraceID:   "trace_456",
		RequestID: "req_789",
	}

	err := transport.PublishEvent(ctx, event)
	assert.NoError(t, err)

	// Verify message was published
	expectedSubject := "test.events.user_queue.job_failed.normal"
	messages := server.GetMessages(expectedSubject)
	require.Len(t, messages, 1)

	message := messages[0]
	assert.Equal(t, expectedSubject, message.Subject)
	assert.Equal(t, "evt_001", message.Headers["Event-ID"])
	assert.Equal(t, "job_failed", message.Headers["Event-Type"])
	assert.Equal(t, "job_123", message.Headers["Job-ID"])
	assert.Equal(t, "user_queue", message.Headers["Queue"])
	assert.Equal(t, "trace_456", message.Headers["Trace-ID"])
	assert.Equal(t, "req_789", message.Headers["Request-ID"])

	// Verify event in message body
	var receivedEvent EventMessage
	err = json.Unmarshal(message.Data, &receivedEvent)
	assert.NoError(t, err)
	assert.Equal(t, event.ID, receivedEvent.ID)
	assert.Equal(t, event.Event, receivedEvent.Event)
	assert.Equal(t, event.JobID, receivedEvent.JobID)
	assert.Equal(t, expectedSubject, receivedEvent.Subject)
}

func TestNATSTransport_SubjectGeneration(t *testing.T) {
	server := NewMockNATSServer()
	config := NATSConfig{SubjectPrefix: "events"}
	transport := NewNATSTransport(server, config)

	testCases := []struct {
		name     string
		event    EventMessage
		expected string
	}{
		{
			name: "normal priority job_failed",
			event: EventMessage{
				Event:    "job_failed",
				Queue:    "test_queue",
				Priority: 5,
			},
			expected: "events.test_queue.job_failed.normal",
		},
		{
			name: "high priority job_dlq",
			event: EventMessage{
				Event:    "job_dlq",
				Queue:    "priority_queue",
				Priority: 9,
			},
			expected: "events.priority_queue.job_dlq.high",
		},
		{
			name: "low priority job_enqueued",
			event: EventMessage{
				Event:    "job_enqueued",
				Queue:    "batch_queue",
				Priority: 2,
			},
			expected: "events.batch_queue.job_enqueued.low",
		},
		{
			name: "edge case - priority 8 (high)",
			event: EventMessage{
				Event:    "job_succeeded",
				Queue:    "edge_queue",
				Priority: 8,
			},
			expected: "events.edge_queue.job_succeeded.high",
		},
		{
			name: "edge case - priority 3 (low)",
			event: EventMessage{
				Event:    "job_retried",
				Queue:    "low_queue",
				Priority: 3,
			},
			expected: "events.low_queue.job_retried.low",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			subject := transport.GenerateSubject(tc.event)
			assert.Equal(t, tc.expected, subject)
		})
	}
}

func TestNATSTransport_CustomSubjectPrefix(t *testing.T) {
	server := NewMockNATSServer()
	config := NATSConfig{SubjectPrefix: "myapp.jobs"}
	transport := NewNATSTransport(server, config)

	event := EventMessage{
		Event:    "job_failed",
		Queue:    "notifications",
		Priority: 7,
	}

	subject := transport.GenerateSubject(event)
	expected := "myapp.jobs.notifications.job_failed.normal"
	assert.Equal(t, expected, subject)
}

func TestNATSTransport_EventSubscription(t *testing.T) {
	server := NewMockNATSServer()
	config := NATSConfig{SubjectPrefix: "events"}
	transport := NewNATSTransport(server, config)

	// Set up subscription
	var receivedEvents []EventMessage
	var mu sync.Mutex

	ctx := context.Background()
	err := transport.SubscribeToPattern(ctx, "events.test_queue.job_failed.normal", func(event EventMessage) error {
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()
		return nil
	})
	assert.NoError(t, err)

	// Publish matching event
	event1 := EventMessage{
		ID:       "evt_001",
		Event:    "job_failed",
		Queue:    "test_queue",
		Priority: 5,
		JobID:    "job_123",
	}

	err = transport.PublishEvent(ctx, event1)
	assert.NoError(t, err)

	// Publish non-matching event
	event2 := EventMessage{
		ID:       "evt_002",
		Event:    "job_succeeded",
		Queue:    "test_queue",
		Priority: 5,
		JobID:    "job_456",
	}

	err = transport.PublishEvent(ctx, event2)
	assert.NoError(t, err)

	// Wait a bit for async processing
	time.Sleep(10 * time.Millisecond)

	// Verify only matching event was received
	mu.Lock()
	assert.Len(t, receivedEvents, 1)
	assert.Equal(t, "evt_001", receivedEvents[0].ID)
	mu.Unlock()
}

func TestNATSTransport_MultipleSubscribers(t *testing.T) {
	server := NewMockNATSServer()
	config := NATSConfig{SubjectPrefix: "events"}
	transport := NewNATSTransport(server, config)

	// Set up multiple subscribers
	var subscriber1Events []EventMessage
	var subscriber2Events []EventMessage
	var mu1, mu2 sync.Mutex

	ctx := context.Background()

	// Subscriber 1 - all job failures
	err := transport.SubscribeToPattern(ctx, "events.*.job_failed.*", func(event EventMessage) error {
		mu1.Lock()
		subscriber1Events = append(subscriber1Events, event)
		mu1.Unlock()
		return nil
	})
	assert.NoError(t, err)

	// Subscriber 2 - all high priority events
	err = transport.SubscribeToPattern(ctx, "events.*.*.high", func(event EventMessage) error {
		mu2.Lock()
		subscriber2Events = append(subscriber2Events, event)
		mu2.Unlock()
		return nil
	})
	assert.NoError(t, err)

	// Publish high priority job failure (should match both)
	event1 := EventMessage{
		ID:       "evt_001",
		Event:    "job_failed",
		Queue:    "test_queue",
		Priority: 9,
		JobID:    "job_123",
	}

	err = transport.PublishEvent(ctx, event1)
	assert.NoError(t, err)

	// Publish normal priority job failure (should match subscriber 1 only)
	event2 := EventMessage{
		ID:       "evt_002",
		Event:    "job_failed",
		Queue:    "test_queue",
		Priority: 5,
		JobID:    "job_456",
	}

	err = transport.PublishEvent(ctx, event2)
	assert.NoError(t, err)

	// Publish high priority job success (should match subscriber 2 only)
	event3 := EventMessage{
		ID:       "evt_003",
		Event:    "job_succeeded",
		Queue:    "test_queue",
		Priority: 8,
		JobID:    "job_789",
	}

	err = transport.PublishEvent(ctx, event3)
	assert.NoError(t, err)

	// Wait for async processing
	time.Sleep(10 * time.Millisecond)

	// Verify subscriber 1 received job failures
	mu1.Lock()
	assert.Len(t, subscriber1Events, 2)
	eventIDs1 := []string{subscriber1Events[0].ID, subscriber1Events[1].ID}
	assert.Contains(t, eventIDs1, "evt_001")
	assert.Contains(t, eventIDs1, "evt_002")
	mu1.Unlock()

	// Verify subscriber 2 received high priority events
	mu2.Lock()
	assert.Len(t, subscriber2Events, 2)
	eventIDs2 := []string{subscriber2Events[0].ID, subscriber2Events[1].ID}
	assert.Contains(t, eventIDs2, "evt_001")
	assert.Contains(t, eventIDs2, "evt_003")
	mu2.Unlock()
}

func TestNATSTransport_ConnectionFailure(t *testing.T) {
	server := NewMockNATSServer()
	config := NATSConfig{SubjectPrefix: "events"}
	transport := NewNATSTransport(server, config)

	// Disconnect server
	server.Disconnect()

	event := EventMessage{
		ID:    "evt_disconnect",
		Event: "job_failed",
		Queue: "test_queue",
		Priority: 5,
	}

	ctx := context.Background()
	err := transport.PublishEvent(ctx, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNATSTransport_InvalidJSON(t *testing.T) {
	server := NewMockNATSServer()
	config := NATSConfig{SubjectPrefix: "events"}
	transport := NewNATSTransport(server, config)

	// Create event with invalid data that can't be marshaled
	event := EventMessage{
		ID:    "evt_invalid",
		Event: "job_failed",
		Queue: "test_queue",
		Data:  map[string]interface{}{"invalid": make(chan int)}, // channels can't be marshaled
	}

	ctx := context.Background()
	err := transport.PublishEvent(ctx, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal")
}

func TestNATSTransport_SubjectPatterns(t *testing.T) {
	server := NewMockNATSServer()
	config := NATSConfig{SubjectPrefix: "myapp"}
	transport := NewNATSTransport(server, config)

	patterns := transport.GetSubjectPatterns()

	expected := []string{
		"myapp.*",
		"myapp.*.job_failed.*",
		"myapp.*.job_dlq.*",
		"myapp.*.*",
		"myapp.priority_queue.*.*",
		"myapp.*.*.high",
	}

	assert.Equal(t, expected, patterns)
}

func TestNATSTransport_DefaultSubjectPrefix(t *testing.T) {
	server := NewMockNATSServer()
	config := NATSConfig{} // No prefix specified
	transport := NewNATSTransport(server, config)

	event := EventMessage{
		Event:    "job_failed",
		Queue:    "test_queue",
		Priority: 5,
	}

	subject := transport.GenerateSubject(event)
	expected := "events.test_queue.job_failed.normal"
	assert.Equal(t, expected, subject)

	patterns := transport.GetSubjectPatterns()
	assert.Contains(t, patterns, "events.*")
}

func TestNATSTransport_ConcurrentPublishing(t *testing.T) {
	server := NewMockNATSServer()
	config := NATSConfig{SubjectPrefix: "events"}
	transport := NewNATSTransport(server, config)

	const numEvents = 50
	var wg sync.WaitGroup
	errors := make([]error, numEvents)

	ctx := context.Background()

	for i := 0; i < numEvents; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			event := EventMessage{
				ID:       fmt.Sprintf("evt_%d", index),
				Event:    "job_completed",
				Queue:    fmt.Sprintf("queue_%d", index%5),
				Priority: index%10 + 1,
				JobID:    fmt.Sprintf("job_%d", index),
			}

			errors[index] = transport.PublishEvent(ctx, event)
		}(i)
	}

	wg.Wait()

	// All publishes should succeed
	for i, err := range errors {
		assert.NoError(t, err, "Event %d should publish successfully", i)
	}

	// Verify all events were published
	totalMessages := 0
	for i := 0; i < 5; i++ {
		for _, eventType := range []string{"job_completed"} {
			for _, priority := range []string{"low", "normal", "high"} {
				subject := fmt.Sprintf("events.queue_%d.%s.%s", i, eventType, priority)
				messages := server.GetMessages(subject)
				totalMessages += len(messages)
			}
		}
	}

	assert.Equal(t, numEvents, totalMessages)
}

// Benchmark Tests

func BenchmarkNATSTransport_PublishEvent(b *testing.B) {
	server := NewMockNATSServer()
	config := NATSConfig{SubjectPrefix: "events"}
	transport := NewNATSTransport(server, config)

	event := EventMessage{
		ID:       "benchmark_event",
		Event:    "job_completed",
		Queue:    "benchmark_queue",
		Priority: 5,
		JobID:    "benchmark_job",
		Data: map[string]interface{}{
			"benchmark": true,
			"iteration": 0,
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event.ID = fmt.Sprintf("benchmark_event_%d", i)
		event.Data["iteration"] = i
		transport.PublishEvent(ctx, event)
	}
}

func BenchmarkNATSTransport_SubjectGeneration(b *testing.B) {
	server := NewMockNATSServer()
	config := NATSConfig{SubjectPrefix: "events"}
	transport := NewNATSTransport(server, config)

	event := EventMessage{
		Event:    "job_failed",
		Queue:    "benchmark_queue",
		Priority: 7,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transport.GenerateSubject(event)
	}
}