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
