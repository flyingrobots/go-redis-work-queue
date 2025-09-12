package queue

import (
    "encoding/json"
    "time"
)

type Job struct {
    ID           string `json:"id"`
    FilePath     string `json:"filepath"`
    FileSize     int64  `json:"filesize"`
    Priority     string `json:"priority"`
    Retries      int    `json:"retries"`
    CreationTime string `json:"creation_time"`
    TraceID      string `json:"trace_id"`
    SpanID       string `json:"span_id"`
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

