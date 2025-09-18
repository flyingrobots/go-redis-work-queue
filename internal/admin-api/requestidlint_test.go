package adminapi

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRequestIDAnalyzer(t *testing.T) {
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	cmd := exec.Command("go", "run", "./tools/requestidlint/cmd/requestidlint", "./internal/admin-api")
	cmd.Dir = repoRoot

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		t.Fatalf("requestidlint failed:\n%s", out.String())
	}
}
