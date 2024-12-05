package main

import (
	"fmt"
	"github.com/ethereum-optimism/optimism/packages/contracts-bedrock/scripts/checks/common"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

func GenerateSolidityInterface(contractName string, astData ContractData, abi abi.ABI) string {
	seenTypes := make(map[string]bool)
	seenStructs := make(map[string]bool)
	seenEnums := make(map[string]bool)
	seenErrors := make(map[string]bool)
	seenEvents := make(map[string]bool)
	seenFunctions := make(map[string]bool)

	var builder strings.Builder

	builder.WriteString("// SPDX-License-Identifier: MIT\n")
	builder.WriteString("pragma solidity ^0.8.0;\n\n")

	// Add user-defined types
	for _, typeDef := range astData.Types {
		typeDefinition := GenerateTypeDefinition(typeDef)
		if !seenTypes[typeDefinition] {
			builder.WriteString(fmt.Sprintf("\n%s", typeDefinition))
			seenTypes[typeDefinition] = true
		}
	}

	builder.WriteString(fmt.Sprintf("\n\ninterface I%s {\n", contractName))

	// Add structs
	for _, structDef := range astData.Structs {
		structDefinition := GenerateStructDefinition(structDef)
		if !seenStructs[structDefinition] {
			builder.WriteString(fmt.Sprintf("\n    %s", structDefinition))
			seenStructs[structDefinition] = true
		}
	}

	// Add enums
	for _, enumDef := range astData.Enums {
		enumSignature := GenerateEnumSignature(enumDef)
		if !seenEnums[enumSignature] {
			builder.WriteString(fmt.Sprintf("\n    %s", enumSignature))
			seenEnums[enumSignature] = true
		}
	}

	// Add errors
	for _, errDef := range astData.Errors {
		errorDefinition := GenerateErrorDefinition(errDef)
		if !seenErrors[errorDefinition] {
			builder.WriteString(fmt.Sprintf("\n    %s", errorDefinition))
			seenErrors[errorDefinition] = true
		}
	}

	// Add events
	for _, event := range astData.Events {
		eventDefinition := GenerateEventDefinition(event)
		if !seenEvents[eventDefinition] {
			builder.WriteString(fmt.Sprintf("\n    %s", eventDefinition))
			seenEvents[eventDefinition] = true
		}
	}

	// Add public function signatures (including public variable getters)
	for _, fn := range astData.Functions {
		functionSignature := GenerateFunctionSignature(fn, abi)
		if !seenFunctions[functionSignature] {
			builder.WriteString(fmt.Sprintf("\n    %s", functionSignature))
			seenFunctions[functionSignature] = true
		}
	}

	// Close the interface definition
	builder.WriteString("\n}\n")

	return builder.String()
}

func main() {
	artifact, _ := common.ReadForgeArtifact("packages/contracts-bedrock/forge-artifacts/DataAvailabilityChallenge.sol/DataAvailabilityChallenge.json")
	astData := ExtractASTData(artifact.Ast, false)

	interfaceCode := GenerateSolidityInterface("ProtocolVersions", astData, artifact.Abi)
	fmt.Println(interfaceCode)
}
