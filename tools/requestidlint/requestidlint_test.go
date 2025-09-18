package requestidlint_test

import (
	"testing"

	"github.com/flyingrobots/go-redis-work-queue/tools/requestidlint"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), requestidlint.Analyzer, "internal/admin-api/good", "internal/admin-api/bad")
}
