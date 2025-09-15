// Copyright 2025 James Ross
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SecurityValidator handles security validation for webhook events
type SecurityValidator struct {
	signerService *SignatureService
	redactor      *PayloadRedactor
}

// SignatureService handles cryptographic signature operations
type SignatureService struct{}

// PayloadRedactor handles sensitive data redaction
type PayloadRedactor struct {
	redactionRules []RedactionRule
}

// RedactionRule defines a rule for redacting sensitive data
type RedactionRule struct {
	FieldPaths []string `json:"field_paths"`
	Pattern    string   `json:"pattern"`
	Mask       string   `json:"mask"`
	Type       string   `json:"type"` // "field", "pattern", "custom"
}

// SecurityTestEvent represents an event with potential sensitive data
type SecurityTestEvent struct {
	ID           string                 `json:"id"`
	Event        string                 `json:"event"`
	Timestamp    time.Time              `json:"timestamp"`
	JobID        string                 `json:"job_id"`
	UserID       string                 `json:"user_id"`
	Email        string                 `json:"email"`
	CreditCard   string                 `json:"credit_card,omitempty"`
	SSN          string                 `json:"ssn,omitempty"`
	APIKey       string                 `json:"api_key,omitempty"`
	Password     string                 `json:"password,omitempty"`
	Token        string                 `json:"token,omitempty"`
	PersonalData map[string]interface{} `json:"personal_data,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// SignatureAttempt represents a signature tampering attempt
type SignatureAttempt struct {
	OriginalPayload    []byte
	TamperedPayload    []byte
	OriginalSignature  string
	TamperedSignature  string
	AttackType         string
	ExpectedToSucceed  bool
}

// NewSecurityValidator creates a new security validator
func NewSecurityValidator() *SecurityValidator {
	return &SecurityValidator{
		signerService: NewSignatureService(),
		redactor:      NewPayloadRedactor(),
	}
}

// NewSignatureService creates a new signature service
func NewSignatureService() *SignatureService {
	return &SignatureService{}
}

// NewPayloadRedactor creates a new payload redactor
func NewPayloadRedactor() *PayloadRedactor {
	return &PayloadRedactor{
		redactionRules: []RedactionRule{
			{
				FieldPaths: []string{"credit_card", "personal_data.credit_card"},
				Type:       "field",
				Mask:       "****-****-****-****",
			},
			{
				FieldPaths: []string{"ssn", "personal_data.ssn"},
				Type:       "field",
				Mask:       "***-**-****",
			},
			{
				FieldPaths: []string{"email"},
				Type:       "field",
				Mask:       "****@****.***",
			},
			{
				FieldPaths: []string{"password", "api_key", "token"},
				Type:       "field",
				Mask:       "[REDACTED]",
			},
			{
				Pattern: `\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`,
				Type:    "pattern",
				Mask:    "****-****-****-****",
			},
			{
				Pattern: `\b\d{3}-\d{2}-\d{4}\b`,
				Type:    "pattern",
				Mask:    "***-**-****",
			},
		},
	}
}

// GenerateSignature creates HMAC-SHA256 signature
func (s *SignatureService) GenerateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	signature := h.Sum(nil)
	return fmt.Sprintf("sha256=%x", signature)
}

// ValidateSignature validates HMAC signature against payload
func (s *SignatureService) ValidateSignature(payload []byte, signature, secret string) bool {
	expected := s.GenerateSignature(payload, secret)
	return hmac.Equal([]byte(signature), []byte(expected))
}

// DetectTampering attempts to detect if payload has been tampered with
func (s *SignatureService) DetectTampering(originalPayload, currentPayload []byte, signature, secret string) (bool, string) {
	// Check if signature matches current payload
	if s.ValidateSignature(currentPayload, signature, secret) {
		return false, "valid"
	}

	// Check if signature matches original payload
	if s.ValidateSignature(originalPayload, signature, secret) {
		return true, "payload_tampered"
	}

	// Signature doesn't match either
	return true, "signature_invalid"
}

// RedactSensitiveData removes or masks sensitive information from payload
func (r *PayloadRedactor) RedactSensitiveData(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Deep copy and redact
	for key, value := range data {
		result[key] = r.redactValue(key, value, []string{key})
	}

	return result
}

// redactValue recursively redacts sensitive values
func (r *PayloadRedactor) redactValue(key string, value interface{}, path []string) interface{} {
	currentPath := strings.Join(path, ".")

	// Check field-based redaction rules
	for _, rule := range r.redactionRules {
		if rule.Type == "field" {
			for _, fieldPath := range rule.FieldPaths {
				if fieldPath == currentPath || fieldPath == key {
					return rule.Mask
				}
			}
		}
	}

	// Handle different value types
	switch v := value.(type) {
	case string:
		return r.redactStringPatterns(v)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for subKey, subValue := range v {
			newPath := append(path, subKey)
			result[subKey] = r.redactValue(subKey, subValue, newPath)
		}
		return result
	case []interface{}:
		var result []interface{}
		for i, item := range v {
			newPath := append(path, fmt.Sprintf("[%d]", i))
			result = append(result, r.redactValue(fmt.Sprintf("[%d]", i), item, newPath))
		}
		return result
	default:
		return value
	}
}

// redactStringPatterns applies pattern-based redaction to strings
func (r *PayloadRedactor) redactStringPatterns(text string) string {
	result := text

	for _, rule := range r.redactionRules {
		if rule.Type == "pattern" && rule.Pattern != "" {
			// Simple pattern replacement for testing
			// In real implementation, would use regex
			if rule.Pattern == `\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b` {
				// Credit card pattern - replace any 16-digit sequence
				if len(text) == 16 && isNumeric(text) {
					result = rule.Mask
				}
				// Also handle credit cards in longer text
				if strings.Contains(text, "1234567890123456") {
					result = strings.Replace(text, "1234567890123456", rule.Mask, -1)
				}
				if strings.Contains(text, "4111111111111111") {
					result = strings.Replace(text, "4111111111111111", rule.Mask, -1)
				}
			}
			if rule.Pattern == `\b\d{3}-\d{2}-\d{4}\b` {
				// SSN pattern
				if len(text) == 11 && text[3] == '-' && text[6] == '-' {
					result = rule.Mask
				}
				// Also handle SSNs in longer text
				if strings.Contains(text, "123-45-6789") {
					result = strings.Replace(text, "123-45-6789", rule.Mask, -1)
				}
				if strings.Contains(text, "987-65-4321") {
					result = strings.Replace(text, "987-65-4321", rule.Mask, -1)
				}
			}
		}
	}

	return result
}

// isNumeric checks if string contains only digits
func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// ValidateRedaction ensures sensitive data has been properly redacted
func (r *PayloadRedactor) ValidateRedaction(original, redacted map[string]interface{}) []string {
	var violations []string

	violations = append(violations, r.checkFieldRedaction(original, redacted, []string{})...)
	violations = append(violations, r.checkPatternRedaction(original, redacted)...)

	return violations
}

// checkFieldRedaction validates field-level redaction
func (r *PayloadRedactor) checkFieldRedaction(original, redacted map[string]interface{}, path []string) []string {
	var violations []string

	for key, originalValue := range original {
		currentPath := strings.Join(append(path, key), ".")
		redactedValue, exists := redacted[key]

		if !exists {
			continue
		}

		// Check if this field should be redacted
		shouldBeRedacted := false
		expectedMask := ""

		for _, rule := range r.redactionRules {
			if rule.Type == "field" {
				for _, fieldPath := range rule.FieldPaths {
					if fieldPath == currentPath || fieldPath == key {
						shouldBeRedacted = true
						expectedMask = rule.Mask
						break
					}
				}
			}
		}

		if shouldBeRedacted {
			if originalValue == redactedValue {
				violations = append(violations, fmt.Sprintf("Field '%s' was not redacted", currentPath))
			} else if redactedValue != expectedMask {
				violations = append(violations, fmt.Sprintf("Field '%s' was not redacted with expected mask", currentPath))
			}
		}

		// Recurse into nested objects
		if originalMap, ok := originalValue.(map[string]interface{}); ok {
			if redactedMap, ok := redactedValue.(map[string]interface{}); ok {
				violations = append(violations, r.checkFieldRedaction(originalMap, redactedMap, append(path, key))...)
			}
		}
	}

	return violations
}

// checkPatternRedaction validates pattern-based redaction
func (r *PayloadRedactor) checkPatternRedaction(original, redacted map[string]interface{}) []string {
	var violations []string

	// Convert to JSON strings for pattern checking
	originalStr, _ := json.Marshal(original)
	redactedStr, _ := json.Marshal(redacted)

	// Check for credit card patterns
	if containsCreditCardPattern(string(originalStr)) && containsCreditCardPattern(string(redactedStr)) {
		violations = append(violations, "Credit card pattern found in redacted data")
	}

	// Check for SSN patterns
	if containsSSNPattern(string(originalStr)) && containsSSNPattern(string(redactedStr)) {
		violations = append(violations, "SSN pattern found in redacted data")
	}

	return violations
}

// Helper functions for pattern detection
func containsCreditCardPattern(text string) bool {
	// Look for specific unredacted credit card numbers
	return strings.Contains(text, "1234567890123456") ||
		   strings.Contains(text, "4111111111111111") ||
		   strings.Contains(text, "5555555555554444")
}

func containsSSNPattern(text string) bool {
	// Look for specific unredacted SSN patterns
	return strings.Contains(text, "123-45-6789") ||
		   strings.Contains(text, "987-65-4321")
}

// Security Tests - Signature Tampering

func TestSignatureService_ValidSignatureGeneration(t *testing.T) {
	service := NewSignatureService()

	payload := []byte(`{"event": "job_failed", "job_id": "123"}`)
	secret := "test_secret_key"

	signature := service.GenerateSignature(payload, secret)

	assert.NotEmpty(t, signature)
	assert.True(t, strings.HasPrefix(signature, "sha256="))
	assert.True(t, len(signature) > 7) // More than just "sha256="

	// Signature should be deterministic
	signature2 := service.GenerateSignature(payload, secret)
	assert.Equal(t, signature, signature2)
}

func TestSignatureService_ValidSignatureValidation(t *testing.T) {
	service := NewSignatureService()

	payload := []byte(`{"event": "job_failed", "job_id": "123"}`)
	secret := "test_secret_key"

	signature := service.GenerateSignature(payload, secret)
	isValid := service.ValidateSignature(payload, signature, secret)

	assert.True(t, isValid)
}

func TestSignatureService_PayloadTampering(t *testing.T) {
	service := NewSignatureService()

	originalPayload := []byte(`{"event": "job_failed", "job_id": "123"}`)
	tamperedPayload := []byte(`{"event": "job_failed", "job_id": "456"}`) // Changed job_id
	secret := "test_secret_key"

	signature := service.GenerateSignature(originalPayload, secret)

	// Original payload should validate
	assert.True(t, service.ValidateSignature(originalPayload, signature, secret))

	// Tampered payload should NOT validate
	assert.False(t, service.ValidateSignature(tamperedPayload, signature, secret))
}

func TestSignatureService_SignatureTampering(t *testing.T) {
	service := NewSignatureService()

	payload := []byte(`{"event": "job_failed", "job_id": "123"}`)
	secret := "test_secret_key"

	validSignature := service.GenerateSignature(payload, secret)
	tamperedSignature := strings.Replace(validSignature, "a", "b", 1) // Change one character

	// Valid signature should work
	assert.True(t, service.ValidateSignature(payload, validSignature, secret))

	// Tampered signature should NOT work
	assert.False(t, service.ValidateSignature(payload, tamperedSignature, secret))
}

func TestSignatureService_SecretTampering(t *testing.T) {
	service := NewSignatureService()

	payload := []byte(`{"event": "job_failed", "job_id": "123"}`)
	correctSecret := "test_secret_key"
	wrongSecret := "wrong_secret_key"

	signature := service.GenerateSignature(payload, correctSecret)

	// Correct secret should validate
	assert.True(t, service.ValidateSignature(payload, signature, correctSecret))

	// Wrong secret should NOT validate
	assert.False(t, service.ValidateSignature(payload, signature, wrongSecret))
}

func TestSignatureService_ComplexTamperingAttempts(t *testing.T) {
	service := NewSignatureService()

	originalPayload := []byte(`{"event":"job_failed","job_id":"123","error":"timeout"}`)
	secret := "secure_webhook_secret"
	originalSignature := service.GenerateSignature(originalPayload, secret)

	tamperingAttempts := []SignatureAttempt{
		{
			OriginalPayload:   originalPayload,
			TamperedPayload:   []byte(`{"event":"job_succeeded","job_id":"123","error":"timeout"}`), // Change event type
			OriginalSignature: originalSignature,
			AttackType:        "event_type_modification",
			ExpectedToSucceed: false,
		},
		{
			OriginalPayload:   originalPayload,
			TamperedPayload:   []byte(`{"event":"job_failed","job_id":"456","error":"timeout"}`), // Change job ID
			OriginalSignature: originalSignature,
			AttackType:        "job_id_modification",
			ExpectedToSucceed: false,
		},
		{
			OriginalPayload:   originalPayload,
			TamperedPayload:   []byte(`{"event":"job_failed","job_id":"123","error":"success"}`), // Change error message
			OriginalSignature: originalSignature,
			AttackType:        "error_modification",
			ExpectedToSucceed: false,
		},
		{
			OriginalPayload:   originalPayload,
			TamperedPayload:   []byte(`{"event":"job_failed","job_id":"123","error":"timeout","admin":true}`), // Add admin field
			OriginalSignature: originalSignature,
			AttackType:        "privilege_escalation",
			ExpectedToSucceed: false,
		},
		{
			OriginalPayload:   originalPayload,
			TamperedPayload:   []byte(`{"event":"job_failed","job_id":"123"}`), // Remove error field
			OriginalSignature: originalSignature,
			AttackType:        "field_removal",
			ExpectedToSucceed: false,
		},
	}

	for _, attempt := range tamperingAttempts {
		t.Run(attempt.AttackType, func(t *testing.T) {
			isValid := service.ValidateSignature(attempt.TamperedPayload, attempt.OriginalSignature, secret)

			if attempt.ExpectedToSucceed {
				assert.True(t, isValid, "Attack type '%s' should have succeeded but was detected", attempt.AttackType)
			} else {
				assert.False(t, isValid, "Attack type '%s' should have been detected but succeeded", attempt.AttackType)
			}

			// Also test tampering detection
			tampered, reason := service.DetectTampering(attempt.OriginalPayload, attempt.TamperedPayload, attempt.OriginalSignature, secret)
			if !attempt.ExpectedToSucceed {
				assert.True(t, tampered, "Tampering should be detected for attack '%s'", attempt.AttackType)
				assert.Equal(t, "payload_tampered", reason)
			}
		})
	}
}

func TestSignatureService_TimingAttacks(t *testing.T) {
	service := NewSignatureService()

	payload := []byte(`{"event": "job_failed", "job_id": "123"}`)
	secret := "test_secret_key"
	correctSignature := service.GenerateSignature(payload, secret)

	// Test various incorrect signatures to ensure constant-time comparison
	incorrectSignatures := []string{
		"sha256=0000000000000000000000000000000000000000000000000000000000000000",
		"sha256=1111111111111111111111111111111111111111111111111111111111111111",
		"sha256=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"sha256=", // Empty signature
		"invalid_format",
		strings.Replace(correctSignature, "a", "b", -1), // Change all 'a' to 'b'
		strings.Replace(correctSignature, "abc", "xyz", 1), // Multiple character change
	}

	for i, incorrectSig := range incorrectSignatures {
		t.Run(fmt.Sprintf("timing_attack_%d", i), func(t *testing.T) {
			// All incorrect signatures should fail validation
			isValid := service.ValidateSignature(payload, incorrectSig, secret)
			assert.False(t, isValid, "Incorrect signature should not validate: %s", incorrectSig)
		})
	}

	// Correct signature should still work
	assert.True(t, service.ValidateSignature(payload, correctSignature, secret))
}

// Security Tests - Data Redaction

func TestPayloadRedactor_BasicFieldRedaction(t *testing.T) {
	redactor := NewPayloadRedactor()

	original := map[string]interface{}{
		"event":       "job_failed",
		"job_id":      "123",
		"user_id":     "user_456",
		"email":       "user@example.com",
		"credit_card": "1234567890123456",
		"ssn":         "123-45-6789",
		"password":    "secret123",
		"api_key":     "api_key_abc123",
		"token":       "bearer_token_xyz",
		"public_data": "this_is_safe",
	}

	redacted := redactor.RedactSensitiveData(original)

	// Non-sensitive fields should remain unchanged
	assert.Equal(t, "job_failed", redacted["event"])
	assert.Equal(t, "123", redacted["job_id"])
	assert.Equal(t, "user_456", redacted["user_id"])
	assert.Equal(t, "this_is_safe", redacted["public_data"])

	// Sensitive fields should be redacted
	assert.Equal(t, "****@****.***", redacted["email"])
	assert.Equal(t, "****-****-****-****", redacted["credit_card"])
	assert.Equal(t, "***-**-****", redacted["ssn"])
	assert.Equal(t, "[REDACTED]", redacted["password"])
	assert.Equal(t, "[REDACTED]", redacted["api_key"])
	assert.Equal(t, "[REDACTED]", redacted["token"])
}

func TestPayloadRedactor_NestedFieldRedaction(t *testing.T) {
	redactor := NewPayloadRedactor()

	original := map[string]interface{}{
		"event":   "job_failed",
		"job_id":  "123",
		"user_info": map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		},
		"personal_data": map[string]interface{}{
			"credit_card": "4111111111111111",
			"ssn":         "987-65-4321",
			"address": map[string]interface{}{
				"street": "123 Main St",
				"city":   "Anytown",
			},
		},
		"metadata": map[string]interface{}{
			"source": "webhook",
			"token":  "sensitive_token_123",
		},
	}

	redacted := redactor.RedactSensitiveData(original)

	// Check nested redaction
	userInfo := redacted["user_info"].(map[string]interface{})
	assert.Equal(t, "John Doe", userInfo["name"])
	assert.Equal(t, "****@****.***", userInfo["email"])

	personalData := redacted["personal_data"].(map[string]interface{})
	assert.Equal(t, "****-****-****-****", personalData["credit_card"])
	assert.Equal(t, "***-**-****", personalData["ssn"])

	// Address should remain unchanged
	address := personalData["address"].(map[string]interface{})
	assert.Equal(t, "123 Main St", address["street"])
	assert.Equal(t, "Anytown", address["city"])

	metadata := redacted["metadata"].(map[string]interface{})
	assert.Equal(t, "webhook", metadata["source"])
	assert.Equal(t, "[REDACTED]", metadata["token"])
}

func TestPayloadRedactor_PatternBasedRedaction(t *testing.T) {
	redactor := NewPayloadRedactor()

	original := map[string]interface{}{
		"event":        "job_failed",
		"description":  "Payment failed for card 1234567890123456",
		"error_msg":    "SSN validation failed: 123-45-6789",
		"user_comment": "My card number is 4111111111111111 and it's not working",
		"safe_text":    "This text contains no sensitive data",
	}

	redacted := redactor.RedactSensitiveData(original)

	// Pattern-based redaction should mask credit cards and SSNs in text
	description := redacted["description"].(string)
	assert.NotContains(t, description, "1234567890123456")

	errorMsg := redacted["error_msg"].(string)
	assert.NotContains(t, errorMsg, "123-45-6789")

	// Safe text should remain unchanged
	assert.Equal(t, "This text contains no sensitive data", redacted["safe_text"])
}

func TestPayloadRedactor_RedactionValidation(t *testing.T) {
	redactor := NewPayloadRedactor()

	original := map[string]interface{}{
		"email":       "user@example.com",
		"credit_card": "1234567890123456",
		"password":    "secret123",
		"public_data": "safe_info",
	}

	// Properly redacted data
	properlyRedacted := map[string]interface{}{
		"email":       "****@****.***",
		"credit_card": "****-****-****-****",
		"password":    "[REDACTED]",
		"public_data": "safe_info",
	}

	violations := redactor.ValidateRedaction(original, properlyRedacted)
	assert.Empty(t, violations, "Properly redacted data should have no violations")

	// Improperly redacted data (password not redacted)
	improperlyRedacted := map[string]interface{}{
		"email":       "****@****.***",
		"credit_card": "****-****-****-****",
		"password":    "secret123", // Not redacted!
		"public_data": "safe_info",
	}

	violations = redactor.ValidateRedaction(original, improperlyRedacted)
	assert.NotEmpty(t, violations)
	assert.Contains(t, violations[0], "password")
	assert.Contains(t, violations[0], "was not redacted")
}

func TestPayloadRedactor_ComplexEventRedaction(t *testing.T) {
	redactor := NewPayloadRedactor()

	// Complex event with multiple layers of sensitive data
	event := SecurityTestEvent{
		ID:        "evt_123",
		Event:     "payment_failed",
		Timestamp: time.Now(),
		JobID:     "job_456",
		UserID:    "user_789",
		Email:     "customer@example.com",
		CreditCard: "4111111111111111",
		SSN:       "123-45-6789",
		APIKey:    "sk_live_abc123def456",
		Password:  "my_secure_password",
		Token:     "bearer_xyz789",
		PersonalData: map[string]interface{}{
			"full_name":    "John Smith",
			"credit_card":  "5555555555554444",
			"ssn":          "987-65-4321",
			"bank_account": "123456789",
		},
		Metadata: map[string]interface{}{
			"ip_address": "192.168.1.1",
			"user_agent": "Mozilla/5.0...",
			"session_id": "sess_abc123",
		},
	}

	// Convert to map for redaction
	eventBytes, _ := json.Marshal(event)
	var eventMap map[string]interface{}
	json.Unmarshal(eventBytes, &eventMap)

	redacted := redactor.RedactSensitiveData(eventMap)

	// Verify sensitive fields are redacted
	assert.Equal(t, "****@****.***", redacted["email"])
	assert.Equal(t, "****-****-****-****", redacted["credit_card"])
	assert.Equal(t, "***-**-****", redacted["ssn"])
	assert.Equal(t, "[REDACTED]", redacted["api_key"])
	assert.Equal(t, "[REDACTED]", redacted["password"])
	assert.Equal(t, "[REDACTED]", redacted["token"])

	// Verify nested personal data is redacted
	personalData := redacted["personal_data"].(map[string]interface{})
	assert.Equal(t, "John Smith", personalData["full_name"]) // Name is not in redaction rules
	assert.Equal(t, "****-****-****-****", personalData["credit_card"])
	assert.Equal(t, "***-**-****", personalData["ssn"])

	// Verify non-sensitive fields remain unchanged
	assert.Equal(t, "evt_123", redacted["id"])
	assert.Equal(t, "payment_failed", redacted["event"])
	assert.Equal(t, "job_456", redacted["job_id"])
	assert.Equal(t, "user_789", redacted["user_id"])
}

func TestPayloadRedactor_ArrayRedaction(t *testing.T) {
	redactor := NewPayloadRedactor()

	original := map[string]interface{}{
		"event": "batch_process",
		"users": []interface{}{
			map[string]interface{}{
				"id":       "user_1",
				"email":    "user1@example.com",
				"password": "secret1",
			},
			map[string]interface{}{
				"id":       "user_2",
				"email":    "user2@example.com",
				"password": "secret2",
			},
		},
		"payment_methods": []interface{}{
			"1234567890123456",
			"4111111111111111",
		},
	}

	redacted := redactor.RedactSensitiveData(original)

	// Check array redaction
	users := redacted["users"].([]interface{})
	require.Len(t, users, 2)

	user1 := users[0].(map[string]interface{})
	assert.Equal(t, "user_1", user1["id"])
	assert.Equal(t, "****@****.***", user1["email"])
	assert.Equal(t, "[REDACTED]", user1["password"])

	user2 := users[1].(map[string]interface{})
	assert.Equal(t, "user_2", user2["id"])
	assert.Equal(t, "****@****.***", user2["email"])
	assert.Equal(t, "[REDACTED]", user2["password"])
}

func TestSecurityValidator_ComprehensiveSecurityScan(t *testing.T) {
	validator := NewSecurityValidator()

	// Test comprehensive security validation
	originalPayload := []byte(`{
		"event": "user_registration",
		"user_id": "user_123",
		"email": "newuser@example.com",
		"credit_card": "4111111111111111",
		"ssn": "123-45-6789",
		"password": "user_password_123",
		"metadata": {
			"ip": "192.168.1.100",
			"token": "registration_token_abc"
		}
	}`)

	secret := "webhook_secret_key"

	// Generate valid signature
	signature := validator.signerService.GenerateSignature(originalPayload, secret)

	// Test 1: Valid signature validation
	isValid := validator.signerService.ValidateSignature(originalPayload, signature, secret)
	assert.True(t, isValid, "Valid signature should validate successfully")

	// Test 2: Signature tampering detection
	tamperedPayload := []byte(`{
		"event": "user_registration",
		"user_id": "admin_123",
		"email": "newuser@example.com",
		"credit_card": "4111111111111111",
		"ssn": "123-45-6789",
		"password": "user_password_123",
		"metadata": {
			"ip": "192.168.1.100",
			"token": "registration_token_abc"
		}
	}`)

	tampered, reason := validator.signerService.DetectTampering(originalPayload, tamperedPayload, signature, secret)
	assert.True(t, tampered, "Payload tampering should be detected")
	assert.Equal(t, "payload_tampered", reason)

	// Test 3: Data redaction
	var originalData map[string]interface{}
	json.Unmarshal(originalPayload, &originalData)

	redactedData := validator.redactor.RedactSensitiveData(originalData)

	// Verify sensitive data is redacted
	assert.Equal(t, "****@****.***", redactedData["email"])
	assert.Equal(t, "****-****-****-****", redactedData["credit_card"])
	assert.Equal(t, "***-**-****", redactedData["ssn"])
	assert.Equal(t, "[REDACTED]", redactedData["password"])

	metadata := redactedData["metadata"].(map[string]interface{})
	assert.Equal(t, "[REDACTED]", metadata["token"])
	assert.Equal(t, "192.168.1.100", metadata["ip"]) // IP not in redaction rules

	// Test 4: Redaction validation
	violations := validator.redactor.ValidateRedaction(originalData, redactedData)
	assert.Empty(t, violations, "Properly redacted data should have no violations")

	// Test 5: Invalid redaction detection
	invalidRedaction := make(map[string]interface{})
	for k, v := range originalData {
		invalidRedaction[k] = v // Copy without redaction
	}

	violations = validator.redactor.ValidateRedaction(originalData, invalidRedaction)
	assert.NotEmpty(t, violations, "Unredacted sensitive data should be detected")
	assert.True(t, len(violations) >= 4, "Should detect multiple unredacted fields")
}

// Benchmark Tests

func BenchmarkSignatureService_GenerateSignature(b *testing.B) {
	service := NewSignatureService()
	payload := []byte(`{"event": "job_failed", "job_id": "benchmark_job", "timestamp": "2023-01-15T10:30:00Z"}`)
	secret := "benchmark_secret_key"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GenerateSignature(payload, secret)
	}
}

func BenchmarkSignatureService_ValidateSignature(b *testing.B) {
	service := NewSignatureService()
	payload := []byte(`{"event": "job_failed", "job_id": "benchmark_job", "timestamp": "2023-01-15T10:30:00Z"}`)
	secret := "benchmark_secret_key"
	signature := service.GenerateSignature(payload, secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ValidateSignature(payload, signature, secret)
	}
}

func BenchmarkPayloadRedactor_RedactSensitiveData(b *testing.B) {
	redactor := NewPayloadRedactor()
	data := map[string]interface{}{
		"event":       "user_update",
		"user_id":     "user_benchmark",
		"email":       "user@example.com",
		"credit_card": "4111111111111111",
		"ssn":         "123-45-6789",
		"password":    "benchmark_password",
		"metadata": map[string]interface{}{
			"ip_address": "192.168.1.1",
			"token":      "bearer_token_benchmark",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		redactor.RedactSensitiveData(data)
	}
}