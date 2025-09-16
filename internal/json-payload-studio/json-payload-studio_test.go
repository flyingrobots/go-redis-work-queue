package jsonpayloadstudio

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewJSONPayloadStudio(t *testing.T) {
	config := &StudioConfig{
		EditorTheme:     "dark",
		SyntaxHighlight: true,
		LineNumbers:     true,
		MaxPayloadSize:  1024 * 1024,
		MaxFieldCount:   1000,
		MaxNestingDepth: 10,
		HistorySize:     50,
	}

	jps := NewJSONPayloadStudio(config, nil)
	if jps == nil {
		t.Fatal("NewJSONPayloadStudio returned nil")
	}

	if jps.config != config {
		t.Error("Config not set correctly")
	}

	if jps.sessions == nil {
		t.Error("Sessions map not initialized")
	}

	if jps.templates == nil {
		t.Error("Templates map not initialized")
	}

	if jps.schemas == nil {
		t.Error("Schemas map not initialized")
	}
}

func TestValidateJSON(t *testing.T) {
	jps := NewJSONPayloadStudio(&StudioConfig{
		MaxNestingDepth: 3,
	}, nil)

	tests := []struct {
		name        string
		content     string
		expectValid bool
		errorCount  int
	}{
		{
			name:        "Valid JSON",
			content:     `{"name": "test", "value": 123}`,
			expectValid: true,
			errorCount:  0,
		},
		{
			name:        "Invalid JSON - missing closing brace",
			content:     `{"name": "test"`,
			expectValid: false,
			errorCount:  1,
		},
		{
			name:        "Invalid JSON - trailing comma",
			content:     `{"name": "test",}`,
			expectValid: false,
			errorCount:  1,
		},
		{
			name:        "Empty string",
			content:     "",
			expectValid: false,
			errorCount:  1,
		},
		{
			name:        "Deep nesting exceeds limit",
			content:     `{"a":{"b":{"c":{"d":"too deep"}}}}`,
			expectValid: false,
			errorCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jps.ValidateJSON(tt.content, nil)
			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectValid, result.Valid)
			}
			if len(result.Errors) != tt.errorCount {
				t.Errorf("Expected %d errors, got %d", tt.errorCount, len(result.Errors))
			}
		})
	}
}

func TestValidateWithSchema(t *testing.T) {
	jps := NewJSONPayloadStudio(&StudioConfig{}, nil)

	schema := &JSONSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
			"age": map[string]interface{}{
				"type":    "number",
				"minimum": 0,
				"maximum": 150,
			},
		},
		Required: []string{"name"},
	}

	tests := []struct {
		name        string
		content     string
		expectValid bool
	}{
		{
			name:        "Valid against schema",
			content:     `{"name": "John", "age": 30}`,
			expectValid: true,
		},
		{
			name:        "Missing required field",
			content:     `{"age": 30}`,
			expectValid: false,
		},
		{
			name:        "Wrong type for field",
			content:     `{"name": "John", "age": "thirty"}`,
			expectValid: false,
		},
		{
			name:        "Value out of range",
			content:     `{"name": "John", "age": 200}`,
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jps.ValidateJSON(tt.content, schema)
			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectValid, result.Valid)
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	jps := NewJSONPayloadStudio(&StudioConfig{}, nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Compact to formatted",
			input:    `{"name":"test","value":123,"nested":{"key":"value"}}`,
			expected: "{\n  \"name\": \"test\",\n  \"value\": 123,\n  \"nested\": {\n    \"key\": \"value\"\n  }\n}",
		},
		{
			name:     "Already formatted",
			input:    "{\n  \"name\": \"test\"\n}",
			expected: "{\n  \"name\": \"test\"\n}",
		},
		{
			name:     "Empty object",
			input:    "{}",
			expected: "{}",
		},
		{
			name:     "Array formatting",
			input:    `[1,2,3,{"key":"value"}]`,
			expected: "[\n  1,\n  2,\n  3,\n  {\n    \"key\": \"value\"\n  }\n]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := jps.FormatJSON(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestTemplateManagement(t *testing.T) {
	jps := NewJSONPayloadStudio(&StudioConfig{}, nil)

	template := &Template{
		ID:          "test-template",
		Name:        "Test Template",
		Description: "A test template",
		Category:    "testing",
		Tags:        []string{"test", "example"},
		Content: map[string]interface{}{
			"type": "test",
			"data": map[string]interface{}{
				"key": "value",
			},
		},
		Variables: []TemplateVariable{
			{
				Name:         "username",
				Type:         "string",
				Required:     true,
				DefaultValue: "user",
			},
		},
	}

	// Test saving template
	err := jps.SaveTemplate(template)
	if err != nil {
		t.Fatalf("Failed to save template: %v", err)
	}

	// Test getting template
	retrieved, err := jps.GetTemplate("test-template")
	if err != nil {
		t.Fatalf("Failed to get template: %v", err)
	}

	if retrieved.ID != template.ID {
		t.Errorf("Template ID mismatch: expected %s, got %s", template.ID, retrieved.ID)
	}

	// Test listing templates
	templates := jps.ListTemplates()
	if len(templates) != 1 {
		t.Errorf("Expected 1 template, got %d", len(templates))
	}

	// Test searching templates
	results := jps.SearchTemplates(&TemplateFilter{
		Query: "test",
	})
	if results.TotalCount != 1 {
		t.Errorf("Expected 1 search result, got %d", results.TotalCount)
	}

	// Test deleting template
	err = jps.DeleteTemplate("test-template")
	if err != nil {
		t.Fatalf("Failed to delete template: %v", err)
	}

	templates = jps.ListTemplates()
	if len(templates) != 0 {
		t.Errorf("Expected 0 templates after deletion, got %d", len(templates))
	}
}

func TestApplyTemplate(t *testing.T) {
	jps := NewJSONPayloadStudio(&StudioConfig{}, nil)

	template := &Template{
		ID:   "var-template",
		Name: "Variable Template",
		Content: map[string]interface{}{
			"user":      "{{username}}",
			"timestamp": "{{now}}",
			"id":        "{{uuid}}",
			"env":       "{{ENV}}",
		},
		Variables: []TemplateVariable{
			{
				Name:         "username",
				DefaultValue: "default_user",
			},
		},
	}

	jps.SaveTemplate(template)

	variables := map[string]interface{}{
		"username": "john_doe",
		"ENV":      "production",
	}

	result, err := jps.ApplyTemplate("var-template", variables)
	if err != nil {
		t.Fatalf("Failed to apply template: %v", err)
	}

	// Check that username was replaced
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	if resultMap["user"] != "john_doe" {
		t.Errorf("Expected user to be 'john_doe', got %v", resultMap["user"])
	}

	if resultMap["env"] != "production" {
		t.Errorf("Expected env to be 'production', got %v", resultMap["env"])
	}

	// Check that dynamic variables were expanded
	if resultMap["timestamp"] == "{{now}}" {
		t.Error("Timestamp variable was not expanded")
	}

	if resultMap["id"] == "{{uuid}}" {
		t.Error("UUID variable was not expanded")
	}
}

func TestSessionManagement(t *testing.T) {
	jps := NewJSONPayloadStudio(&StudioConfig{
		HistorySize: 10,
	}, nil)

	// Create session
	sessionID := jps.CreateSession()
	if sessionID == "" {
		t.Fatal("Failed to create session")
	}

	// Get session
	session, err := jps.GetSession(sessionID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if session.ID != sessionID {
		t.Errorf("Session ID mismatch: expected %s, got %s", sessionID, session.ID)
	}

	// Update editor state
	state := &EditorState{
		Content:      `{"test": "data"}`,
		CursorLine:   1,
		CursorColumn: 5,
		Modified:     true,
	}

	err = jps.UpdateEditorState(sessionID, state)
	if err != nil {
		t.Fatalf("Failed to update editor state: %v", err)
	}

	// Get updated session
	session, err = jps.GetSession(sessionID)
	if err != nil {
		t.Fatalf("Failed to get updated session: %v", err)
	}

	if session.EditorState.Content != state.Content {
		t.Error("Editor state content not updated")
	}

	// Delete session
	err = jps.DeleteSession(sessionID)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	_, err = jps.GetSession(sessionID)
	if err == nil {
		t.Error("Expected error when getting deleted session")
	}
}

func TestAutoCompletion(t *testing.T) {
	jps := NewJSONPayloadStudio(&StudioConfig{}, nil)

	schema := &JSONSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "User's name",
			},
			"age": map[string]interface{}{
				"type":        "number",
				"description": "User's age",
			},
			"email": map[string]interface{}{
				"type":        "string",
				"format":      "email",
				"description": "User's email address",
			},
		},
	}

	context := `{"na`
	position := &Position{Line: 1, Column: 5}

	completions := jps.GetCompletions(context, position, schema)
	if len(completions) == 0 {
		t.Error("Expected completions, got none")
	}

	// Check that "name" is in completions
	found := false
	for _, c := range completions {
		if strings.Contains(c.Label, "name") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'name' in completions")
	}
}

func TestSnippetExpansion(t *testing.T) {
	jps := NewJSONPayloadStudio(&StudioConfig{}, nil)

	snippet := &Snippet{
		ID:      "user-snippet",
		Name:    "User Object",
		Trigger: "user",
		Content: map[string]interface{}{
			"id":        "{{uuid}}",
			"name":      "{{$1:username}}",
			"email":     "{{$2:email}}",
			"createdAt": "{{now}}",
		},
	}

	jps.SaveSnippet(snippet)

	expanded, err := jps.ExpandSnippet("user", nil)
	if err != nil {
		t.Fatalf("Failed to expand snippet: %v", err)
	}

	expandedMap, ok := expanded.(map[string]interface{})
	if !ok {
		t.Fatal("Expanded snippet is not a map")
	}

	// Check that dynamic variables were expanded
	if expandedMap["id"] == "{{uuid}}" {
		t.Error("UUID was not expanded")
	}

	if expandedMap["createdAt"] == "{{now}}" {
		t.Error("Timestamp was not expanded")
	}

	// Check that placeholders are present
	if expandedMap["name"] != "{{$1:username}}" {
		t.Error("Name placeholder was incorrectly expanded")
	}
}

func TestDiffPayloads(t *testing.T) {
	jps := NewJSONPayloadStudio(&StudioConfig{}, nil)

	old := map[string]interface{}{
		"name":    "John",
		"age":     30,
		"city":    "New York",
		"hobbies": []interface{}{"reading", "gaming"},
	}

	new := map[string]interface{}{
		"name":    "John",
		"age":     31,
		"country": "USA",
		"hobbies": []interface{}{"reading", "gaming", "coding"},
	}

	result, err := jps.DiffPayloads(old, new)
	if err != nil {
		t.Fatalf("Failed to diff payloads: %v", err)
	}

	if !result.HasChanges {
		t.Error("Expected changes to be detected")
	}

	// Check for specific changes
	if len(result.Modified) == 0 {
		t.Error("Expected modified fields")
	}

	if len(result.Added) == 0 {
		t.Error("Expected added fields")
	}

	if len(result.Removed) == 0 {
		t.Error("Expected removed fields")
	}
}

func TestEnqueuePayload(t *testing.T) {
	// This test requires a mock Redis client
	mockRedis := &mockRedisClient{
		enqueueFunc: func(queue string, payload interface{}, options ...interface{}) (string, error) {
			return "job-123", nil
		},
	}

	jps := NewJSONPayloadStudio(&StudioConfig{
		MaxPayloadSize:  1024 * 1024,
		RequireConfirm:  false,
		StripSecrets:    true,
		SecretPatterns:  []string{"password", "secret", "token", "key"},
	}, mockRedis)

	sessionID := jps.CreateSession()
	state := &EditorState{
		Content: `{
			"action": "process",
			"data": {
				"user": "john",
				"password": "secret123",
				"value": 100
			}
		}`,
	}
	jps.UpdateEditorState(sessionID, state)

	options := &EnqueueOptions{
		Queue:      "test-queue",
		Count:      1,
		Priority:   5,
		MaxRetries: 3,
	}

	result, err := jps.EnqueuePayload(sessionID, options)
	if err != nil {
		t.Fatalf("Failed to enqueue payload: %v", err)
	}

	if result.Queue != "test-queue" {
		t.Errorf("Expected queue 'test-queue', got %s", result.Queue)
	}

	if len(result.JobIDs) != 1 {
		t.Errorf("Expected 1 job ID, got %d", len(result.JobIDs))
	}

	// Check that password was stripped
	payloadMap, ok := result.Payload.(map[string]interface{})
	if !ok {
		t.Fatal("Payload is not a map")
	}

	dataMap, ok := payloadMap["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Data is not a map")
	}

	if dataMap["password"] != "[REDACTED]" {
		t.Error("Password was not stripped from payload")
	}
}

func TestHistoryManagement(t *testing.T) {
	jps := NewJSONPayloadStudio(&StudioConfig{
		HistorySize: 5,
	}, nil)

	sessionID := jps.CreateSession()

	// Add multiple history entries
	states := []string{
		`{"step": 1}`,
		`{"step": 2}`,
		`{"step": 3}`,
		`{"step": 4}`,
		`{"step": 5}`,
		`{"step": 6}`, // This should cause the first entry to be removed
	}

	for _, content := range states {
		state := &EditorState{Content: content}
		jps.UpdateEditorState(sessionID, state)
		jps.AddToHistory(sessionID, content)
	}

	session, _ := jps.GetSession(sessionID)

	// Check that history is limited to 5 entries
	if len(session.EditorState.History) != 5 {
		t.Errorf("Expected history size of 5, got %d", len(session.EditorState.History))
	}

	// Check that the oldest entry was removed
	if session.EditorState.History[0] == states[0] {
		t.Error("Oldest entry was not removed from history")
	}

	// Test undo
	err := jps.Undo(sessionID)
	if err != nil {
		t.Fatalf("Failed to undo: %v", err)
	}

	session, _ = jps.GetSession(sessionID)
	if session.EditorState.Content == states[5] {
		t.Error("Undo did not change content")
	}

	// Test redo
	err = jps.Redo(sessionID)
	if err != nil {
		t.Fatalf("Failed to redo: %v", err)
	}
}

func TestSecurityFeatures(t *testing.T) {
	jps := NewJSONPayloadStudio(&StudioConfig{
		MaxPayloadSize:  100, // Small limit for testing
		MaxFieldCount:   5,
		MaxNestingDepth: 2,
		StripSecrets:    true,
		SecretPatterns:  []string{"password", "secret", "api_key", "token"},
	}, nil)

	tests := []struct {
		name        string
		content     string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Payload too large",
			content:     `{"data": "` + strings.Repeat("x", 200) + `"}`,
			shouldError: true,
			errorMsg:    "exceeds maximum size",
		},
		{
			name: "Too many fields",
			content: `{
				"field1": "value1",
				"field2": "value2",
				"field3": "value3",
				"field4": "value4",
				"field5": "value5",
				"field6": "value6"
			}`,
			shouldError: true,
			errorMsg:    "exceeds maximum field count",
		},
		{
			name:        "Nesting too deep",
			content:     `{"a": {"b": {"c": "too deep"}}}`,
			shouldError: true,
			errorMsg:    "exceeds maximum nesting depth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionID := jps.CreateSession()
			state := &EditorState{Content: tt.content}
			jps.UpdateEditorState(sessionID, state)

			_, err := jps.EnqueuePayload(sessionID, &EnqueueOptions{
				Queue: "test",
				Count: 1,
			})

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestStripSecrets(t *testing.T) {
	jps := NewJSONPayloadStudio(&StudioConfig{
		StripSecrets:   true,
		SecretPatterns: []string{"password", "secret", "api_key", "token", "Bearer\\s+[\\w-]+"},
	}, nil)

	input := map[string]interface{}{
		"username":       "john",
		"password":       "super_secret_123",
		"api_key":        "sk-1234567890",
		"access_token":   "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		"data": map[string]interface{}{
			"secret_value": "hidden",
			"public_value": "visible",
		},
		"tokens": []interface{}{
			"token1",
			"token2",
		},
	}

	result := jps.stripSecrets(input)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Check that secrets were stripped
	if resultMap["password"] != "[REDACTED]" {
		t.Error("Password was not redacted")
	}

	if resultMap["api_key"] != "[REDACTED]" {
		t.Error("API key was not redacted")
	}

	if resultMap["access_token"] != "[REDACTED]" {
		t.Error("Access token was not redacted")
	}

	// Check nested secrets
	dataMap, ok := resultMap["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Data is not a map")
	}

	if dataMap["secret_value"] != "[REDACTED]" {
		t.Error("Nested secret was not redacted")
	}

	if dataMap["public_value"] != "visible" {
		t.Error("Public value was incorrectly redacted")
	}

	// Check array secrets
	tokens, ok := resultMap["tokens"].([]interface{})
	if !ok {
		t.Fatal("Tokens is not an array")
	}

	for _, token := range tokens {
		if token != "[REDACTED]" {
			t.Error("Token in array was not redacted")
		}
	}
}

// Mock Redis client for testing
type mockRedisClient struct {
	enqueueFunc func(queue string, payload interface{}, options ...interface{}) (string, error)
}

func (m *mockRedisClient) Enqueue(queue string, payload interface{}, options ...interface{}) (string, error) {
	if m.enqueueFunc != nil {
		return m.enqueueFunc(queue, payload, options...)
	}
	return "mock-job-id", nil
}

func (m *mockRedisClient) EnqueueWithSchedule(queue string, payload interface{}, schedule time.Time) (string, error) {
	return "mock-scheduled-job-id", nil
}

func (m *mockRedisClient) EnqueueWithCron(queue string, payload interface{}, cronSpec string) (string, error) {
	return "mock-cron-job-id", nil
}

// Benchmark tests
func BenchmarkValidateJSON(b *testing.B) {
	jps := NewJSONPayloadStudio(&StudioConfig{}, nil)
	content := `{
		"name": "test",
		"value": 123,
		"nested": {
			"key": "value",
			"array": [1, 2, 3, 4, 5]
		}
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jps.ValidateJSON(content, nil)
	}
}

func BenchmarkFormatJSON(b *testing.B) {
	jps := NewJSONPayloadStudio(&StudioConfig{}, nil)
	content := `{"name":"test","value":123,"nested":{"key":"value","array":[1,2,3,4,5]}}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jps.FormatJSON(content)
	}
}

func BenchmarkStripSecrets(b *testing.B) {
	jps := NewJSONPayloadStudio(&StudioConfig{
		StripSecrets:   true,
		SecretPatterns: []string{"password", "secret", "token", "key"},
	}, nil)

	payload := map[string]interface{}{
		"user":     "john",
		"password": "secret123",
		"data": map[string]interface{}{
			"api_key": "sk-123456",
			"value":   100,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jps.stripSecrets(payload)
	}
}

func TestErrorCases(t *testing.T) {
	jps := NewJSONPayloadStudio(&StudioConfig{}, nil)

	// Test getting non-existent session
	_, err := jps.GetSession("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent session")
	}

	// Test getting non-existent template
	_, err = jps.GetTemplate("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent template")
	}

	// Test applying non-existent template
	_, err = jps.ApplyTemplate("non-existent", nil)
	if err == nil {
		t.Error("Expected error for non-existent template")
	}

	// Test expanding non-existent snippet
	_, err = jps.ExpandSnippet("non-existent", nil)
	if err == nil {
		t.Error("Expected error for non-existent snippet")
	}

	// Test undo with no history
	sessionID := jps.CreateSession()
	err = jps.Undo(sessionID)
	if err == nil {
		t.Error("Expected error for undo with no history")
	}

	// Test redo with no redo history
	err = jps.Redo(sessionID)
	if err == nil {
		t.Error("Expected error for redo with no redo history")
	}
}

func TestConcurrentAccess(t *testing.T) {
	jps := NewJSONPayloadStudio(&StudioConfig{}, nil)

	// Test concurrent session creation
	sessionIDs := make([]string, 10)
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			sessionIDs[index] = jps.CreateSession()
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Check that all sessions were created
	uniqueSessions := make(map[string]bool)
	for _, id := range sessionIDs {
		if id == "" {
			t.Error("Empty session ID")
		}
		uniqueSessions[id] = true
	}

	if len(uniqueSessions) != 10 {
		t.Errorf("Expected 10 unique sessions, got %d", len(uniqueSessions))
	}

	// Test concurrent template operations
	for i := 0; i < 5; i++ {
		go func(index int) {
			template := &Template{
				ID:   fmt.Sprintf("template-%d", index),
				Name: fmt.Sprintf("Template %d", index),
			}
			jps.SaveTemplate(template)
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	templates := jps.ListTemplates()
	if len(templates) != 5 {
		t.Errorf("Expected 5 templates, got %d", len(templates))
	}
}

func TestComplexJSONStructures(t *testing.T) {
	jps := NewJSONPayloadStudio(&StudioConfig{
		MaxNestingDepth: 10,
		MaxFieldCount:   100,
	}, nil)

	// Test with complex nested structure
	complex := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{
				"id":    1,
				"name":  "Alice",
				"roles": []interface{}{"admin", "user"},
				"settings": map[string]interface{}{
					"theme": "dark",
					"notifications": map[string]interface{}{
						"email": true,
						"push":  false,
					},
				},
			},
			map[string]interface{}{
				"id":    2,
				"name":  "Bob",
				"roles": []interface{}{"user"},
				"settings": map[string]interface{}{
					"theme": "light",
					"notifications": map[string]interface{}{
						"email": false,
						"push":  true,
					},
				},
			},
		},
		"metadata": map[string]interface{}{
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
			"features": map[string]interface{}{
				"experimental": []interface{}{"feature1", "feature2"},
				"stable":       []interface{}{"core", "auth", "api"},
			},
		},
	}

	jsonBytes, err := json.Marshal(complex)
	if err != nil {
		t.Fatalf("Failed to marshal complex structure: %v", err)
	}

	result := jps.ValidateJSON(string(jsonBytes), nil)
	if !result.Valid {
		t.Error("Complex structure should be valid")
	}

	// Test formatting
	formatted, err := jps.FormatJSON(string(jsonBytes))
	if err != nil {
		t.Errorf("Failed to format complex structure: %v", err)
	}

	if formatted == "" {
		t.Error("Formatted output is empty")
	}

	// Test diffing complex structures
	modified := make(map[string]interface{})
	for k, v := range complex {
		modified[k] = v
	}

	// Modify the structure
	users := modified["users"].([]interface{})
	users = append(users, map[string]interface{}{
		"id":   3,
		"name": "Charlie",
	})
	modified["users"] = users

	diff, err := jps.DiffPayloads(complex, modified)
	if err != nil {
		t.Errorf("Failed to diff complex structures: %v", err)
	}

	if !diff.HasChanges {
		t.Error("Should detect changes in complex structure")
	}
}

func TestLintStats(t *testing.T) {
	jps := NewJSONPayloadStudio(&StudioConfig{}, nil)

	content := `{
		"string": "value",
		"number": 42,
		"boolean": true,
		"null": null,
		"array": [1, 2, 3],
		"object": {
			"nested": {
				"deep": "value"
			}
		}
	}`

	result := jps.ValidateJSON(content, nil)
	if !result.Valid {
		t.Error("Content should be valid")
	}

	stats := result.Stats
	if stats.StringCount != 2 {
		t.Errorf("Expected 2 strings, got %d", stats.StringCount)
	}

	if stats.NumberCount != 4 { // 42, 1, 2, 3
		t.Errorf("Expected 4 numbers, got %d", stats.NumberCount)
	}

	if stats.BooleanCount != 1 {
		t.Errorf("Expected 1 boolean, got %d", stats.BooleanCount)
	}

	if stats.NullCount != 1 {
		t.Errorf("Expected 1 null, got %d", stats.NullCount)
	}

	if stats.ArrayCount != 1 {
		t.Errorf("Expected 1 array, got %d", stats.ArrayCount)
	}

	if stats.ObjectCount != 3 { // root, object, nested
		t.Errorf("Expected 3 objects, got %d", stats.ObjectCount)
	}

	if stats.MaxDepth != 3 {
		t.Errorf("Expected max depth of 3, got %d", stats.MaxDepth)
	}
}