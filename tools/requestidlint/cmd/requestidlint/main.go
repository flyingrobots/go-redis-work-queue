package main

import (
	"github.com/flyingrobots/go-redis-work-queue/tools/requestidlint"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(requestidlint.Analyzer)
}
