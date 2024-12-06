package main

import (
	"github.com/ethereum-optimism/optimism/op-chain-ops/solc"
	"strings"
)

func isTrivialType(typeString string) bool {
	// TODO: Add types or find a way to generate them
	trivialTypes := []string{"uint256", "int256", "bool", "address", "bytes32", "uint", "int"}

	for _, t := range trivialTypes {
		if typeString == t {
			return true
		}
	}

	return false
}

func isImportUsed(importNode solc.AstNode, usedTypes map[string]bool) bool {
	if importNode.AbsolutePath != "" {
		baseName := getBaseName(importNode.AbsolutePath)
		return usedTypes[baseName]
	}
	return false
}

func getBaseName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		fileName := parts[len(parts)-1]
		return strings.TrimSuffix(fileName, ".sol")
	}
	return path
}

func stripContractPrefix(typeString string) string {
	if strings.HasPrefix(typeString, "contract ") {
		parts := strings.SplitN(typeString, " ", 2)
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return typeString
}
