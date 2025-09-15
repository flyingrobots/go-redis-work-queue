package voice

import (
	"testing"
)

func TestNewCommandProcessor(t *testing.T) {
	processor, err := NewCommandProcessor()
	if err != nil {
		t.Fatalf("Failed to create command processor: %v", err)
	}

	if processor == nil {
		t.Fatal("Processor is nil")
	}

	if len(processor.patterns) == 0 {
		t.Error("No command patterns loaded")
	}

	if processor.entities == nil {
		t.Error("Entity extractor is nil")
	}

	if processor.context == nil {
		t.Error("Command context is nil")
	}
}

func TestParseCommand(t *testing.T) {
	processor, err := NewCommandProcessor()
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	tests := []struct {
		name           string
		text           string
		expectedIntent Intent
		expectError    bool
	}{
		{
			name:           "queue status query",
			text:           "show queue status",
			expectedIntent: IntentStatusQuery,
			expectError:    false,
		},
		{
			name:           "worker status query",
			text:           "show workers",
			expectedIntent: IntentStatusQuery,
			expectError:    false,
		},
		{
			name:           "dlq status query",
			text:           "show dlq",
			expectedIntent: IntentStatusQuery,
			expectError:    false,
		},
		{
			name:           "drain worker",
			text:           "drain worker 3",
			expectedIntent: IntentWorkerControl,
			expectError:    false,
		},
		{
			name:           "pause worker",
			text:           "pause worker 1",
			expectedIntent: IntentWorkerControl,
			expectError:    false,
		},
		{
			name:           "resume worker",
			text:           "resume worker 2",
			expectedIntent: IntentWorkerControl,
			expectError:    false,
		},
		{
			name:           "requeue failed jobs",
			text:           "requeue failed jobs",
			expectedIntent: IntentQueueManagement,
			expectError:    false,
		},
		{
			name:           "clear completed jobs",
			text:           "clear completed jobs",
			expectedIntent: IntentQueueManagement,
			expectError:    false,
		},
		{
			name:           "navigate to workers",
			text:           "go to workers",
			expectedIntent: IntentNavigation,
			expectError:    false,
		},
		{
			name:           "navigate to dlq",
			text:           "show dlq",
			expectedIntent: IntentStatusQuery, // This could match status query first
			expectError:    false,
		},
		{
			name:           "confirmation",
			text:           "yes",
			expectedIntent: IntentConfirmation,
			expectError:    false,
		},
		{
			name:           "cancellation",
			text:           "cancel",
			expectedIntent: IntentCancel,
			expectError:    false,
		},
		{
			name:           "help request",
			text:           "help",
			expectedIntent: IntentHelp,
			expectError:    false,
		},
		{
			name:        "unknown command",
			text:        "unknown command xyz",
			expectError: true,
		},
		{
			name:        "empty command",
			text:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				RawText: tt.text,
			}

			err := processor.ParseCommand(cmd)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if cmd.Intent != tt.expectedIntent {
					t.Errorf("Expected intent %v, got %v", tt.expectedIntent, cmd.Intent)
				}

				if cmd.Confidence <= 0.0 || cmd.Confidence > 1.0 {
					t.Errorf("Invalid confidence: %f", cmd.Confidence)
				}
			}
		})
	}
}

func TestNewEntityExtractor(t *testing.T) {
	extractor, err := NewEntityExtractor()
	if err != nil {
		t.Fatalf("Failed to create entity extractor: %v", err)
	}

	if extractor == nil {
		t.Fatal("Extractor is nil")
	}

	if len(extractor.queueNames) == 0 {
		t.Error("No queue names configured")
	}

	if len(extractor.workerIDs) == 0 {
		t.Error("No worker IDs configured")
	}

	if len(extractor.patterns) == 0 {
		t.Error("No extraction patterns configured")
	}
}

func TestEntityExtraction(t *testing.T) {
	extractor, err := NewEntityExtractor()
	if err != nil {
		t.Fatalf("Failed to create extractor: %v", err)
	}

	tests := []struct {
		name           string
		text           string
		expectedTypes  []EntityType
		expectedValues []string
	}{
		{
			name:           "worker ID extraction",
			text:           "drain worker 3",
			expectedTypes:  []EntityType{EntityWorkerID, EntityAction},
			expectedValues: []string{"3", "drain"},
		},
		{
			name:           "worker ID with word",
			text:           "pause worker two",
			expectedTypes:  []EntityType{EntityWorkerID, EntityAction},
			expectedValues: []string{"2", "pause"}, // Should normalize "two" to "2"
		},
		{
			name:           "queue name extraction",
			text:           "show high priority queue",
			expectedTypes:  []EntityType{EntityQueueName, EntityTarget},
			expectedValues: []string{"high", "queue"},
		},
		{
			name:           "action extraction",
			text:           "requeue failed jobs",
			expectedTypes:  []EntityType{EntityAction},
			expectedValues: []string{"requeue"},
		},
		{
			name:           "destination extraction",
			text:           "go to workers",
			expectedTypes:  []EntityType{EntityDestination},
			expectedValues: []string{"workers"},
		},
		{
			name:           "number extraction",
			text:           "show last 5 jobs",
			expectedTypes:  []EntityType{EntityNumber},
			expectedValues: []string{"5"},
		},
		{
			name:           "multiple entities",
			text:           "drain worker 2 from high priority queue",
			expectedTypes:  []EntityType{EntityWorkerID, EntityAction, EntityQueueName},
			expectedValues: []string{"2", "drain", "high"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entities, err := extractor.Extract(tt.text, []string{})
			if err != nil {
				t.Fatalf("Failed to extract entities: %v", err)
			}

			if len(entities) == 0 && len(tt.expectedTypes) > 0 {
				t.Fatal("No entities extracted")
			}

			// Check that expected entity types are present
			foundTypes := make(map[EntityType]bool)
			foundValues := make(map[string]bool)

			for _, entity := range entities {
				foundTypes[entity.Type] = true
				foundValues[entity.Value] = true
			}

			for i, expectedType := range tt.expectedTypes {
				if !foundTypes[expectedType] {
					t.Errorf("Expected entity type %v not found", expectedType)
				}

				if i < len(tt.expectedValues) {
					expectedValue := tt.expectedValues[i]
					if !foundValues[expectedValue] {
						t.Errorf("Expected entity value '%s' not found", expectedValue)
					}
				}
			}
		})
	}
}

func TestNormalizeEntityValue(t *testing.T) {
	extractor, err := NewEntityExtractor()
	if err != nil {
		t.Fatalf("Failed to create extractor: %v", err)
	}

	tests := []struct {
		entityType EntityType
		input      string
		expected   string
	}{
		{EntityWorkerID, "one", "1"},
		{EntityWorkerID, "two", "2"},
		{EntityWorkerID, "three", "3"},
		{EntityWorkerID, "5", "5"},
		{EntityDestination, "workers", "workers"},
		{EntityDestination, "worker", "workers"},
		{EntityDestination, "dlq", "dlq"},
		{EntityDestination, "dead letter", "dlq"},
		{EntityAction, "stop", "drain"},
		{EntityAction, "restart", "resume"},
		{EntityAction, "retry", "requeue"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractor.normalizeEntityValue(tt.entityType, tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestCalculateSimilarity(t *testing.T) {
	extractor, err := NewEntityExtractor()
	if err != nil {
		t.Fatalf("Failed to create extractor: %v", err)
	}

	tests := []struct {
		s1       string
		s2       string
		minScore float64
	}{
		{"hello", "hello", 1.0},
		{"high priority", "high", 0.7},
		{"worker", "workers", 0.7},
		{"dlq", "dead letter queue", 0.1},
		{"completely different", "xyz", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.s1+"_"+tt.s2, func(t *testing.T) {
			similarity := extractor.calculateSimilarity(tt.s1, tt.s2)

			if similarity < tt.minScore {
				t.Errorf("Expected similarity >= %f, got %f", tt.minScore, similarity)
			}

			if similarity < 0.0 || similarity > 1.0 {
				t.Errorf("Similarity out of range [0,1]: %f", similarity)
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	processor, err := NewCommandProcessor()
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	tests := []struct {
		name        string
		command     *Command
		expectValid bool
		expectError bool
	}{
		{
			name: "valid status query",
			command: &Command{
				Intent:     IntentStatusQuery,
				Confidence: 0.8,
				Entities: []Entity{
					{Type: EntityTarget, Value: "queue"},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "valid worker control",
			command: &Command{
				Intent:     IntentWorkerControl,
				Confidence: 0.9,
				Entities: []Entity{
					{Type: EntityWorkerID, Value: "1"},
					{Type: EntityAction, Value: "drain"},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "worker control without worker ID",
			command: &Command{
				Intent:     IntentWorkerControl,
				Confidence: 0.8,
			},
			expectValid: false,
			expectError: true,
		},
		{
			name: "navigation without destination",
			command: &Command{
				Intent:     IntentNavigation,
				Confidence: 0.8,
			},
			expectValid: false,
			expectError: true,
		},
		{
			name: "low confidence command",
			command: &Command{
				Intent:     IntentStatusQuery,
				Confidence: 0.2,
			},
			expectValid: true, // Valid but with warning
			expectError: false,
		},
		{
			name: "unknown intent",
			command: &Command{
				Intent:     IntentUnknown,
				Confidence: 0.8,
			},
			expectValid: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.ValidateCommand(tt.command)

			if result == nil {
				t.Fatal("Validation result is nil")
			}

			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got valid=%v", tt.expectValid, result.Valid)
			}

			hasErrors := len(result.Errors) > 0
			if hasErrors != tt.expectError {
				t.Errorf("Expected hasErrors=%v, got hasErrors=%v (errors: %v)",
					tt.expectError, hasErrors, result.Errors)
			}
		})
	}
}

func TestUpdateContext(t *testing.T) {
	processor, err := NewCommandProcessor()
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	// Test navigation context update
	navCmd := &Command{
		Intent: IntentNavigation,
		Entities: []Entity{
			{Type: EntityDestination, Value: "workers"},
		},
	}

	processor.UpdateContext(navCmd, true)

	if processor.context.currentView != "workers" {
		t.Errorf("Expected current view 'workers', got '%s'", processor.context.currentView)
	}

	if processor.context.lastCommand != navCmd {
		t.Error("Last command not updated")
	}

	// Test worker selection context update
	workerCmd := &Command{
		Intent: IntentWorkerControl,
		Entities: []Entity{
			{Type: EntityWorkerID, Value: "3"},
			{Type: EntityAction, Value: "drain"},
		},
	}

	processor.UpdateContext(workerCmd, true)

	if processor.context.selectedWorker != "3" {
		t.Errorf("Expected selected worker '3', got '%s'", processor.context.selectedWorker)
	}

	// Test confirmation context
	processor.SetConfirmationPending(true)
	if !processor.context.confirmPending {
		t.Error("Expected confirmation to be pending")
	}

	confirmCmd := &Command{
		Intent: IntentConfirmation,
	}

	processor.UpdateContext(confirmCmd, true)

	if processor.context.confirmPending {
		t.Error("Expected confirmation to no longer be pending")
	}
}

func TestGetCommandPatterns(t *testing.T) {
	processor, err := NewCommandProcessor()
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	patterns := processor.GetCommandPatterns()

	if len(patterns) == 0 {
		t.Error("No command patterns returned")
	}

	// Check that all patterns have required fields
	for i, pattern := range patterns {
		if pattern.Pattern == nil {
			t.Errorf("Pattern %d has nil regex", i)
		}

		if pattern.Intent == IntentUnknown {
			t.Errorf("Pattern %d has unknown intent", i)
		}

		if pattern.Description == "" {
			t.Errorf("Pattern %d has empty description", i)
		}

		if len(pattern.Examples) == 0 {
			t.Errorf("Pattern %d has no examples", i)
		}
	}
}

func TestContextGetters(t *testing.T) {
	processor, err := NewCommandProcessor()
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	// Test initial context
	context := processor.GetContext()
	if context == nil {
		t.Fatal("Context is nil")
	}

	if context.currentView != "" {
		t.Errorf("Expected empty current view, got '%s'", context.currentView)
	}

	if context.confirmPending {
		t.Error("Expected confirmation not to be pending initially")
	}

	// Test confirmation pending setter
	processor.SetConfirmationPending(true)
	if !processor.context.confirmPending {
		t.Error("Failed to set confirmation pending")
	}

	processor.SetConfirmationPending(false)
	if processor.context.confirmPending {
		t.Error("Failed to clear confirmation pending")
	}
}

func TestComplexCommandParsing(t *testing.T) {
	processor, err := NewCommandProcessor()
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	complexCommands := []struct {
		name     string
		text     string
		expectOK bool
	}{
		{
			name:     "natural language worker control",
			text:     "please drain the third worker",
			expectOK: true,
		},
		{
			name:     "natural language status query",
			text:     "can you show me what's in the high priority queue",
			expectOK: true,
		},
		{
			name:     "casual navigation",
			text:     "take me to the workers tab",
			expectOK: true,
		},
		{
			name:     "polite confirmation",
			text:     "yes please proceed",
			expectOK: true,
		},
		{
			name:     "urgent worker control",
			text:     "immediately stop all workers",
			expectOK: true,
		},
		{
			name:     "queue management with time",
			text:     "requeue failed jobs from the last hour",
			expectOK: true,
		},
	}

	for _, tt := range complexCommands {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				RawText: tt.text,
			}

			err := processor.ParseCommand(cmd)

			if tt.expectOK && err != nil {
				t.Errorf("Expected successful parsing, got error: %v", err)
			}

			if !tt.expectOK && err == nil {
				t.Error("Expected parsing to fail")
			}

			if tt.expectOK {
				if cmd.Intent == IntentUnknown {
					t.Error("Failed to determine intent for complex command")
				}

				t.Logf("Command: '%s' -> Intent: %v, Confidence: %.2f, Entities: %d",
					tt.text, cmd.Intent, cmd.Confidence, len(cmd.Entities))
			}
		})
	}
}