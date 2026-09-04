// Copyright 2026 James Ross
package main

import "testing"

func TestCLIWorkerRequiresExplicitBenchMode(t *testing.T) {
	for _, role := range []string{"worker", "all"} {
		t.Run(role, func(t *testing.T) {
			if handler, err := cliWorkerHandler(role, false); err == nil || handler != nil {
				t.Fatalf("handler = %v, err = %v; want explicit opt-in error", handler, err)
			}
			if handler, err := cliWorkerHandler(role, true); err != nil || handler == nil {
				t.Fatalf("explicit bench handler = %v, err = %v", handler, err)
			}
		})
	}
}

func TestCLIProducerDoesNotRequireWorkerHandler(t *testing.T) {
	if handler, err := cliWorkerHandler("producer", false); err != nil || handler != nil {
		t.Fatalf("producer handler = %v, err = %v", handler, err)
	}
}
