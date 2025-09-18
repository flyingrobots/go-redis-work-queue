package voice

import (
	"fmt"
	"regexp"
	"strings"
)

// NewCommandProcessor creates a new command processor with predefined patterns
func NewCommandProcessor() (*CommandProcessor, error) {
	processor := &CommandProcessor{
		patterns: make([]CommandPattern, 0),
		context:  NewCommandContext(),
	}

	extractor, err := NewEntityExtractor()
	if err != nil {
		return nil, fmt.Errorf("failed to create entity extractor: %w", err)
	}
	processor.entities = extractor

	// Initialize command patterns
	processor.initializePatterns()

	return processor, nil
}

// NewCommandContext creates a new command context
func NewCommandContext() *CommandContext {
	return &CommandContext{
		confirmPending: false,
	}
}

// ParseCommand parses a voice command and extracts intent and entities
func (c *CommandProcessor) ParseCommand(cmd *Command) error {
	if cmd.RawText == "" {
		return fmt.Errorf("empty command text")
	}

	// Normalize input text
	normalized := strings.ToLower(strings.TrimSpace(cmd.RawText))

	// Try each pattern until we find a match
	for _, pattern := range c.patterns {
		if matches := pattern.Pattern.FindStringSubmatch(normalized); matches != nil {
			// Extract entities from the match
			entities, err := c.entities.Extract(cmd.RawText, matches)
			if err != nil {
				continue
			}

			// Validate required entities are present
			if !c.validateRequiredEntities(pattern, entities) {
				continue
			}

			// Update command with parsed information
			cmd.Intent = pattern.Intent
			cmd.Entities = entities
			cmd.Confidence = c.calculatePatternConfidence(matches, entities)

			return nil
		}
	}

	return fmt.Errorf("no matching command pattern for: %s", cmd.RawText)
}

// initializePatterns sets up all command recognition patterns
func (c *CommandProcessor) initializePatterns() {
	c.patterns = []CommandPattern{
		// Status queries
		{
			Pattern:     regexp.MustCompile(`(?:show|display|tell|what).*(?:queue|status)`),
			Intent:      IntentStatusQuery,
			Required:    []EntityType{},
			Optional:    []EntityType{EntityTarget},
			Description: "Show queue status",
			Examples:    []string{"show queue status", "what is the queue status", "display queues"},
		},
		{
			Pattern:     regexp.MustCompile(`(?:show|display|tell|what).*workers?`),
			Intent:      IntentStatusQuery,
			Required:    []EntityType{},
			Optional:    []EntityType{EntityTarget},
			Description: "Show worker status",
			Examples:    []string{"show workers", "what are the workers doing", "display worker status"},
		},
		{
			Pattern:     regexp.MustCompile(`(?:show|display|go to|check).*(?:dlq|dead letter|failed)`),
			Intent:      IntentStatusQuery,
			Required:    []EntityType{},
			Optional:    []EntityType{EntityTarget},
			Description: "Show dead letter queue",
			Examples:    []string{"show dlq", "go to dead letter queue", "check failed jobs"},
		},
		{
			Pattern:     regexp.MustCompile(`(?:how many|count).*jobs`),
			Intent:      IntentStatusQuery,
			Required:    []EntityType{},
			Optional:    []EntityType{EntityTarget},
			Description: "Count jobs",
			Examples:    []string{"how many jobs", "count the jobs", "how many jobs in high priority"},
		},

		// Worker control
		{
			Pattern:     regexp.MustCompile(`(?:drain|stop|shutdown).*worker\s*(\d+|all)`),
			Intent:      IntentWorkerControl,
			Required:    []EntityType{EntityWorkerID},
			Optional:    []EntityType{EntityAction},
			Description: "Drain/stop worker",
			Examples:    []string{"drain worker 3", "stop worker 1", "shutdown all workers"},
		},
		{
			Pattern:     regexp.MustCompile(`(?:(?:please|kindly|immediately|urgently)\s+)?(?:drain|stop|shutdown)\s+(?:the\s+)?(first|second|third|fourth|fifth|one|two|three|four|five|\d+|all)\s+workers?`),
			Intent:      IntentWorkerControl,
			Required:    []EntityType{EntityWorkerID},
			Optional:    []EntityType{EntityAction},
			Description: "Drain/stop worker (natural language)",
			Examples:    []string{"please drain the third worker", "kindly stop the first worker", "immediately stop all workers"},
		},
		{
			Pattern:     regexp.MustCompile(`(?:pause|halt).*worker\s*(\d+|all)`),
			Intent:      IntentWorkerControl,
			Required:    []EntityType{EntityWorkerID},
			Optional:    []EntityType{EntityAction},
			Description: "Pause worker",
			Examples:    []string{"pause worker 2", "halt worker 1", "pause all workers"},
		},
		{
			Pattern:     regexp.MustCompile(`(?:resume|start|restart).*worker\s*(\d+|all)`),
			Intent:      IntentWorkerControl,
			Required:    []EntityType{EntityWorkerID},
			Optional:    []EntityType{EntityAction},
			Description: "Resume/start worker",
			Examples:    []string{"resume worker 1", "start worker 3", "restart all workers"},
		},

		// Queue management
		{
			Pattern:     regexp.MustCompile(`(?:requeue|retry).*(?:failed|error|dlq)`),
			Intent:      IntentQueueManagement,
			Required:    []EntityType{},
			Optional:    []EntityType{EntityAction, EntityTimeRange},
			Description: "Requeue failed jobs",
			Examples:    []string{"requeue failed jobs", "retry all errors", "requeue dlq"},
		},
		{
			Pattern:     regexp.MustCompile(`(?:clear|cleanup|remove).*(?:completed|finished|done)`),
			Intent:      IntentQueueManagement,
			Required:    []EntityType{},
			Optional:    []EntityType{EntityAction, EntityTimeRange},
			Description: "Clear completed jobs",
			Examples:    []string{"clear completed jobs", "cleanup finished jobs", "remove done tasks"},
		},
		{
			Pattern:     regexp.MustCompile(`(?:pause|stop).*queue`),
			Intent:      IntentQueueManagement,
			Required:    []EntityType{},
			Optional:    []EntityType{EntityQueueName},
			Description: "Pause queue",
			Examples:    []string{"pause queue", "stop the queue", "pause high priority queue"},
		},

		// Navigation
		{
			Pattern:     regexp.MustCompile(`(?:go to|show|switch to|navigate to|take(?:\s+me)?\s+to|bring\s+me\s+to).*(?:queue|workers?|dlq|stats?|charts?|logs?|config|settings?)`),
			Intent:      IntentNavigation,
			Required:    []EntityType{EntityDestination},
			Optional:    []EntityType{},
			Description: "Navigate to tab/view",
			Examples:    []string{"go to workers", "show charts", "navigate to settings"},
		},

		// Confirmations
		{
			Pattern:     regexp.MustCompile(`^(?:yes|yep|yeah)(?:\s+(?:please|sure))?(?:\s+(?:proceed|do it|go ahead|continue))?$|^(?:confirm|ok|okay|proceed|do it)$`),
			Intent:      IntentConfirmation,
			Required:    []EntityType{},
			Optional:    []EntityType{},
			Description: "Confirm action",
			Examples:    []string{"yes", "confirm", "okay", "proceed"},
		},
		{
			Pattern:     regexp.MustCompile(`^(?:no|nope|cancel|abort|stop|nevermind)$`),
			Intent:      IntentCancel,
			Required:    []EntityType{},
			Optional:    []EntityType{},
			Description: "Cancel action",
			Examples:    []string{"no", "cancel", "abort", "nevermind"},
		},

		// Help
		{
			Pattern:     regexp.MustCompile(`(?:help|what can|commands|usage)`),
			Intent:      IntentHelp,
			Required:    []EntityType{},
			Optional:    []EntityType{},
			Description: "Show help",
			Examples:    []string{"help", "what can you do", "show commands"},
		},
	}
}

// validateRequiredEntities checks if all required entities are present
func (c *CommandProcessor) validateRequiredEntities(pattern CommandPattern, entities []Entity) bool {
	if len(pattern.Required) == 0 {
		return true
	}

	requiredMap := make(map[EntityType]bool)
	for _, reqType := range pattern.Required {
		requiredMap[reqType] = false
	}

	for _, entity := range entities {
		if _, exists := requiredMap[entity.Type]; exists {
			requiredMap[entity.Type] = true
		}
	}

	for _, found := range requiredMap {
		if !found {
			return false
		}
	}

	return true
}

// calculatePatternConfidence calculates confidence based on pattern match quality
func (c *CommandProcessor) calculatePatternConfidence(matches []string, entities []Entity) float64 {
	baseConfidence := 0.7

	// Boost confidence for exact matches
	if len(matches) > 1 {
		baseConfidence += 0.1
	}

	// Boost confidence for entity extraction
	entityBonus := float64(len(entities)) * 0.05
	if entityBonus > 0.2 {
		entityBonus = 0.2
	}

	confidence := baseConfidence + entityBonus

	// Clamp to [0.0, 1.0]
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// NewEntityExtractor creates a new entity extractor
func NewEntityExtractor() (*EntityExtractor, error) {
	extractor := &EntityExtractor{
		queueNames: []string{
			"high", "medium", "low", "critical",
			"priority", "background", "urgent",
			"processing", "pending", "failed",
		},
		workerIDs: []string{
			"1", "2", "3", "4", "5", "all",
			"one", "two", "three", "four", "five",
		},
		patterns: make(map[EntityType]*regexp.Regexp),
	}

	// Initialize extraction patterns
	extractor.initializePatterns()

	return extractor, nil
}

// initializePatterns sets up entity extraction patterns
func (e *EntityExtractor) initializePatterns() {
	e.patterns = map[EntityType]*regexp.Regexp{
		EntityWorkerID:    regexp.MustCompile(`(?:worker\s*(?:number\s*)?(\d+|all|one|two|three|four|five|first|second|third|fourth|fifth)|(?:the\s+)?(first|second|third|fourth|fifth|one|two|three|four|five|all)\s+workers?)`),
		EntityQueueName:   regexp.MustCompile(`(high|medium|low|critical|priority|background|urgent|processing|pending|failed)(?:\s+priority)?\s*queue`),
		EntityNumber:      regexp.MustCompile(`(\d+)`),
		EntityTimeRange:   regexp.MustCompile(`(?:last|past)\s*(\d+)\s*(minutes?|hours?|days?)`),
		EntityDestination: regexp.MustCompile(`(?:to|the)\s*(queue|workers?|dlq|dead\s*letter|stats?|statistics|charts?|graphs?|logs?|config|settings?)`),
		EntityAction:      regexp.MustCompile(`(drain|stop|pause|resume|start|restart|requeue|retry|clear|cleanup|remove)`),
	}
}

// Extract extracts entities from text using regex patterns and context
func (e *EntityExtractor) Extract(text string, matches []string) ([]Entity, error) {
	var entities []Entity
	lowerText := strings.ToLower(text)

	// Extract entities using patterns
	for entityType, pattern := range e.patterns {
		if entityMatches := pattern.FindAllStringSubmatch(lowerText, -1); entityMatches != nil {
			for _, match := range entityMatches {
				value := ""
				for i := 1; i < len(match); i++ {
					if strings.TrimSpace(match[i]) != "" {
						value = strings.TrimSpace(match[i])
						break
					}
				}
				if value != "" {
					entity := Entity{
						Type:       entityType,
						Value:      value,
						Confidence: 0.9,
					}

					// Set position information if available
					if pos := pattern.FindStringIndex(lowerText); pos != nil {
						entity.Start = pos[0]
						entity.End = pos[1]
					}

					// Normalize entity values
					entity.Value = e.normalizeEntityValue(entityType, entity.Value)

					entities = append(entities, entity)
				}
			}
		}
	}

	// Extract target entities for status queries
	if strings.Contains(lowerText, "queue") {
		entities = append(entities, Entity{
			Type:       EntityTarget,
			Value:      "queue",
			Confidence: 0.8,
		})
	}

	if strings.Contains(lowerText, "worker") {
		entities = append(entities, Entity{
			Type:       EntityTarget,
			Value:      "workers",
			Confidence: 0.8,
		})
	}

	if strings.Contains(lowerText, "dlq") || strings.Contains(lowerText, "dead letter") {
		entities = append(entities, Entity{
			Type:       EntityTarget,
			Value:      "dlq",
			Confidence: 0.8,
		})
	}

	// Fuzzy match queue names
	entities = append(entities, e.extractQueueNamesWithFuzzy(lowerText)...)

	return entities, nil
}

// normalizeEntityValue normalizes entity values to standard forms
func (e *EntityExtractor) normalizeEntityValue(entityType EntityType, value string) string {
	switch entityType {
	case EntityWorkerID:
		// Convert word numbers to digits
		numberMap := map[string]string{
			"one": "1", "two": "2", "three": "3", "four": "4", "five": "5",
			"first": "1", "second": "2", "third": "3", "fourth": "4", "fifth": "5",
		}
		if digit, exists := numberMap[value]; exists {
			return digit
		}
		return value

	case EntityDestination:
		// Normalize destination names
		destMap := map[string]string{
			"workers":     "workers",
			"worker":      "workers",
			"queue":       "queue",
			"queues":      "queue",
			"dlq":         "dlq",
			"dead letter": "dlq",
			"stats":       "stats",
			"statistics":  "stats",
			"charts":      "charts",
			"graphs":      "charts",
			"logs":        "logs",
			"config":      "config",
			"settings":    "config",
		}
		if normalized, exists := destMap[value]; exists {
			return normalized
		}
		return value

	case EntityAction:
		// Normalize action verbs
		actionMap := map[string]string{
			"stop":    "drain",
			"halt":    "pause",
			"restart": "resume",
			"retry":   "requeue",
			"cleanup": "clear",
			"remove":  "clear",
		}
		if normalized, exists := actionMap[value]; exists {
			return normalized
		}
		return value

	default:
		return value
	}
}

// extractQueueNamesWithFuzzy performs fuzzy matching for queue names
func (e *EntityExtractor) extractQueueNamesWithFuzzy(text string) []Entity {
	var entities []Entity

	for _, queueName := range e.queueNames {
		similarity := e.calculateSimilarity(text, queueName)
		if similarity > 0.7 {
			entities = append(entities, Entity{
				Type:       EntityQueueName,
				Value:      queueName,
				Similarity: similarity,
				Confidence: similarity,
			})
		}
	}

	return entities
}

// calculateSimilarity calculates string similarity using Levenshtein distance
func (e *EntityExtractor) calculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	// Simple contains check for basic fuzzy matching
	if strings.Contains(s1, s2) || strings.Contains(s2, s1) {
		return 0.8
	}

	// Basic character overlap calculation
	s1Chars := make(map[rune]int)
	s2Chars := make(map[rune]int)

	for _, char := range s1 {
		s1Chars[char]++
	}

	for _, char := range s2 {
		s2Chars[char]++
	}

	overlap := 0
	for char, count1 := range s1Chars {
		if count2, exists := s2Chars[char]; exists {
			if count1 < count2 {
				overlap += count1
			} else {
				overlap += count2
			}
		}
	}

	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	if maxLen == 0 {
		return 0.0
	}

	return float64(overlap) / float64(maxLen)
}

// GetCommandPatterns returns all available command patterns for help/documentation
func (c *CommandProcessor) GetCommandPatterns() []CommandPattern {
	return c.patterns
}

// ValidateCommand validates a command before execution
func (c *CommandProcessor) ValidateCommand(cmd *Command) *ValidationResult {
	result := &ValidationResult{
		Valid:      true,
		Errors:     make([]string, 0),
		Warnings:   make([]string, 0),
		Confidence: cmd.Confidence,
	}

	// Check confidence threshold
	if cmd.Confidence < 0.5 {
		result.Warnings = append(result.Warnings, "Low confidence recognition")
	}

	// Validate intent-specific requirements
	switch cmd.Intent {
	case IntentWorkerControl:
		if cmd.GetEntity(EntityWorkerID) == nil {
			result.Valid = false
			result.Errors = append(result.Errors, "Worker control requires worker ID")
		}

	case IntentNavigation:
		if cmd.GetEntity(EntityDestination) == nil {
			result.Valid = false
			result.Errors = append(result.Errors, "Navigation requires destination")
		}

	case IntentStatusQuery:
		// Status queries are generally valid without specific entities
		break

	case IntentUnknown:
		result.Valid = false
		result.Errors = append(result.Errors, "Unknown command intent")
	}

	// Check for dangerous operations
	if cmd.Intent == IntentWorkerControl {
		workerEntity := cmd.GetEntity(EntityWorkerID)
		if workerEntity != nil && workerEntity.Value == "all" {
			result.Warnings = append(result.Warnings, "Operation affects all workers")
		}
	}

	if cmd.Intent == IntentQueueManagement {
		result.Warnings = append(result.Warnings, "Queue management operation")
	}

	return result
}

// UpdateContext updates command context based on executed command
func (c *CommandProcessor) UpdateContext(cmd *Command, success bool) {
	c.context.lastCommand = cmd

	// Update current view based on navigation
	if cmd.Intent == IntentNavigation && success {
		dest := cmd.GetEntity(EntityDestination)
		if dest != nil {
			c.context.currentView = dest.Value
		}
	}

	// Update selected entities
	if workerEntity := cmd.GetEntity(EntityWorkerID); workerEntity != nil {
		c.context.selectedWorker = workerEntity.Value
	}

	if queueEntity := cmd.GetEntity(EntityQueueName); queueEntity != nil {
		c.context.selectedQueue = queueEntity.Value
	}

	// Handle confirmation state
	if cmd.Intent == IntentConfirmation || cmd.Intent == IntentCancel {
		c.context.confirmPending = false
	}
}

// GetContext returns current command context
func (c *CommandProcessor) GetContext() *CommandContext {
	return c.context
}

// SetConfirmationPending sets whether a confirmation is pending
func (c *CommandProcessor) SetConfirmationPending(pending bool) {
	c.context.confirmPending = pending
}
