// Copyright 2025 James Ross
package chaosharness

import "fmt"

// InjectedError represents an error injected by chaos harness
type InjectedError struct {
	Message    string
	InjectorID string
	Scope      InjectorScope
	ScopeValue string
}

func (e *InjectedError) Error() string {
	return fmt.Sprintf("[CHAOS] Injected error: %s (injector=%s, scope=%s:%s)",
		e.Message, e.InjectorID, e.Scope, e.ScopeValue)
}

// ScenarioError represents an error during scenario execution
type ScenarioError struct {
	ScenarioID string
	Stage      string
	Cause      error
}

func (e *ScenarioError) Error() string {
	return fmt.Sprintf("scenario %s failed at stage %s: %v", e.ScenarioID, e.Stage, e.Cause)
}

func (e *ScenarioError) Unwrap() error {
	return e.Cause
}

// GuardrailViolation represents a guardrail violation
type GuardrailViolation struct {
	ScenarioID string
	Guardrail  string
	Current    interface{}
	Limit      interface{}
}

func (e *GuardrailViolation) Error() string {
	return fmt.Sprintf("guardrail violation in scenario %s: %s (current=%v, limit=%v)",
		e.ScenarioID, e.Guardrail, e.Current, e.Limit)
}
