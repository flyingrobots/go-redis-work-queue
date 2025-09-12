package queue

import "testing"

func TestMarshalUnmarshal(t *testing.T) {
    j := NewJob("id", "/tmp/x", 42, "high", "t", "s")
    s, err := j.Marshal()
    if err != nil { t.Fatal(err) }
    j2, err := UnmarshalJob(s)
    if err != nil { t.Fatal(err) }
    if j2.ID != j.ID || j2.FilePath != j.FilePath || j2.Priority != j.Priority {
        t.Fatalf("roundtrip mismatch: %#v vs %#v", j, j2)
    }
}

