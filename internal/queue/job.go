// Copyright 2025 James Ross
package queue

import (
	"encoding/json"
	"time"
)

// Job is the durable envelope stored in Redis.
//
// Payload contains opaque application bytes. The standard JSON encoding for a
// byte slice is base64, which lets Marshal and UnmarshalJob preserve arbitrary
// bytes exactly. PayloadSchema is an optional, caller-owned type or version
// discriminator; an empty schema is valid.
//
// A non-empty OrderingKey enables crash-safe FIFO execution for jobs sharing
// the exact same key. Empty ordering keys retain the ordinary priority queues.
//
// FilePath and FileSize are legacy benchmark fields. They remain for backward
// compatibility and for the built-in benchmark handler, but they are metadata,
// not the application payload.
type Job struct {
	ID            string `json:"id"`
	FilePath      string `json:"filepath"`
	FileSize      int64  `json:"filesize"`
	Payload       []byte `json:"payload,omitempty"`
	PayloadSchema string `json:"payload_schema,omitempty"`
	OrderingKey   string `json:"ordering_key,omitempty"`
	Priority      string `json:"priority"`
	Retries       int    `json:"retries"`
	CreationTime  string `json:"creation_time"`
	TraceID       string `json:"trace_id"`
	SpanID        string `json:"span_id"`
}

func NewJob(id, path string, size int64, priority string, traceID, spanID string) Job {
	return Job{
		ID:           id,
		FilePath:     path,
		FileSize:     size,
		Priority:     priority,
		Retries:      0,
		CreationTime: time.Now().UTC().Format(time.RFC3339Nano),
		TraceID:      traceID,
		SpanID:       spanID,
	}
}

func (j Job) Marshal() (string, error) {
	b, err := json.Marshal(j)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func UnmarshalJob(s string) (Job, error) {
	var j Job
	err := json.Unmarshal([]byte(s), &j)
	return j, err
}
