// Copyright 2026 James Ross
package queue

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestReadmeDocsReferencesExist(t *testing.T) {
	repoRoot := filepath.Join("..", "..")
	contents, err := os.ReadFile(filepath.Join(repoRoot, "README.md"))
	if err != nil {
		t.Fatal(err)
	}

	referencePattern := regexp.MustCompile(`docs/[A-Za-z0-9._/-]+`)
	seen := map[string]struct{}{}
	for _, reference := range referencePattern.FindAllString(string(contents), -1) {
		reference = strings.TrimRight(reference, "/")
		if _, duplicate := seen[reference]; duplicate {
			continue
		}
		seen[reference] = struct{}{}
		if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(reference))); err != nil {
			t.Errorf("README reference %q is not present: %v", reference, err)
		}
	}
}

func TestReadmePackageCountMatchesEnforcedMinimum(t *testing.T) {
	repoRoot := filepath.Join("..", "..")
	readme, err := os.ReadFile(filepath.Join(repoRoot, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	gate, err := os.ReadFile(filepath.Join(repoRoot, "scripts", "check_test_package_count.sh"))
	if err != nil {
		t.Fatal(err)
	}

	readmeMatch := regexp.MustCompile(`minimum of ([0-9]+) default test-bearing packages`).FindSubmatch(readme)
	gateMatch := regexp.MustCompile(`(?m)^expected_minimum=([0-9]+)$`).FindSubmatch(gate)
	if len(readmeMatch) != 2 || len(gateMatch) != 2 {
		t.Fatalf("could not find package-count minimum in README or gate")
	}
	if string(readmeMatch[1]) != string(gateMatch[1]) {
		t.Fatalf("README minimum = %s, enforced minimum = %s", readmeMatch[1], gateMatch[1])
	}
}
