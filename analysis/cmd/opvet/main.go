// Package opvet provides a "go vet" plugin for Optimism.
//
// It is intended to be used as `go vet -vettool=/path/to/opvet pkg...` in the
// Optimism repo.
//
// This helps enforcing some conventions in the codebase that are stricter than
// common Go practices.
package main

import (
	"golang.org/x/tools/go/analysis/unitchecker"

	"github.com/ethereum-optimism/optimism/analysis/passes/testingt"
)

func main() {
	unitchecker.Main(
		testingt.Analyzer,
	)
}
