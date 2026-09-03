// Copyright 2025 James Ross
package queue

import (
	"strings"
	"testing"
)

func TestMarshalUnmarshal(t *testing.T) {
	j := NewJob("id", "/tmp/x", 42, "high", "t", "s")
	s, err := j.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	j2, err := UnmarshalJob(s)
	if err != nil {
		t.Fatal(err)
	}
	if j2.ID != j.ID || j2.FilePath != j.FilePath || j2.Priority != j.Priority {
		t.Fatalf("roundtrip mismatch: %#v vs %#v", j, j2)
	}
}

func TestUnmarshalPrePayloadJob(t *testing.T) {
	const fixture = `{"id":"legacy-id","filepath":"/tmp/legacy.dat","filesize":42,"priority":"low","retries":0,"creation_time":"2025-09-13T07:18:00Z","trace_id":"","span_id":""}`

	job, err := UnmarshalJob(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if len(job.Payload) != 0 {
		t.Fatalf("expected an empty payload, got %q", job.Payload)
	}
	if job.PayloadSchema != "" {
		t.Fatalf("expected an empty payload schema, got %q", job.PayloadSchema)
	}

	encoded, err := job.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(encoded, `"payload`) {
		t.Fatalf("expected empty payload fields to be omitted, got %s", encoded)
	}
}
