package main

import (
	"fmt"
	"strings"

	"github.com/ethereum-optimism/optimism/packages/contracts-bedrock/scripts/checks/common"
)

func GenerateSolidityInterface(contractName string, astData ContractData) string {
	seenImports := make(map[string]bool) // Track imports already added
	seenTypes := make(map[string]bool)
	seenStructs := make(map[string]bool)
	seenEnums := make(map[string]bool)
	seenErrors := make(map[string]bool)
	seenEvents := make(map[string]bool)
	seenFunctions := make(map[string]bool)

	usedTypes := map[string]bool{}

	// Collect used types from all components
	collectUsedTypes := func(typeString string) {
		if typeString == "" {
			return
		}
		parts := strings.Fields(typeString)
		if len(parts) > 1 {
			usedTypes[parts[1]] = true
		}
	}

	// Analyze Functions
	for _, fn := range astData.Functions {
		if fn.Parameters != nil {
			for _, param := range fn.Parameters.Parameters {
				collectUsedTypes(param.TypeDescriptions.TypeString)
			}
		}
		if fn.ReturnParameters != nil {
			for _, ret := range fn.ReturnParameters.Parameters {
				collectUsedTypes(ret.TypeDescriptions.TypeString)
			}
		}
	}

	// Analyze Structs
	for _, structDef := range astData.Structs {
		for _, member := range structDef.Members {
			collectUsedTypes(member.Type)
		}
	}

	// Analyze Events
	for _, event := range astData.Events {
		if event.Parameters != nil {
			for _, param := range event.Parameters.Parameters {
				collectUsedTypes(param.TypeDescriptions.TypeString)
			}
		}
	}

	// Analyze Errors
	for _, errDef := range astData.Errors {
		if errDef.Parameters != nil {
			for _, param := range errDef.Parameters.Parameters {
				collectUsedTypes(param.TypeDescriptions.TypeString)
			}
		}
	}

	// Add SPDX license and pragma version
	var builder strings.Builder
	builder.WriteString("// SPDX-License-Identifier: MIT\n")
	builder.WriteString(fmt.Sprintf("pragma solidity %s;\n\n", astData.Version))

	// Add imports
	for _, importNode := range astData.Imports {
		if importNode.NodeType == "ImportDirective" && isImportUsed(importNode, usedTypes) {
			importDefinition := GenerateImportDefinition(importNode)
			if importDefinition != "" && !seenImports[importDefinition] {
				builder.WriteString(fmt.Sprintf("%s\n", importDefinition))
				seenImports[importDefinition] = true
			}
		}
	}

	// Add user-defined types
	for _, typeDef := range astData.Types {
		typeDefinition := GenerateTypeDefinition(typeDef)
		if !seenTypes[typeDefinition] {
			builder.WriteString(fmt.Sprintf("\n%s", typeDefinition))
			seenTypes[typeDefinition] = true
		}
	}

	// Start the interface declaration
	builder.WriteString(fmt.Sprintf("\ninterface I%s {\n", contractName))

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
		functionSignature := GenerateFunctionSignature(fn)
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
	artifact, _ := common.ReadForgeArtifact("packages/contracts-bedrock/forge-artifacts/AddressManager.sol/AddressManager.json")
	astData := ExtractASTData(artifact.Ast, false, "")

	interfaceCode := GenerateSolidityInterface("DelayedWETH", astData)
	fmt.Println(interfaceCode)
}
