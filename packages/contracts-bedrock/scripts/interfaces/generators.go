package main

import (
	"fmt"
	"strings"

	"github.com/ethereum-optimism/optimism/op-chain-ops/solc"
)

func GenerateInterfaceDeclaration(contractName string, inherited []solc.AstBaseContract) string {
	if contractName == "" {
		return ""
	}

	// Start the interface declaration
	interfaceHeader := fmt.Sprintf("interface I%s", contractName)

	// Add inherited base contracts if any
	if len(inherited) > 0 {
		baseContracts := []string{}
		for _, baseContract := range inherited {
			baseContracts = append(baseContracts, baseContract.BaseName.Name)
		}
		interfaceHeader += " is " + strings.Join(baseContracts, ", ")
	}

	// Close the interface declaration
	interfaceHeader += " {\n"
	return interfaceHeader
}

func GenerateImportDefinition(importNode solc.AstNode) string {
	filePath := importNode.AbsolutePath
	if filePath == "" {
		return ""
	}

	// Handle simple and aliased imports
	if len(importNode.SymbolAliases) > 0 {
		var aliasDefinitions []string
		for _, alias := range importNode.SymbolAliases {
			if alias.Local != "" {
				// Aliased symbol: `x as y`
				aliasDefinitions = append(aliasDefinitions, fmt.Sprintf("%s as %s", alias.Foreign.Name, alias.Local))
			} else {
				// Simple symbol: `x`
				aliasDefinitions = append(aliasDefinitions, alias.Foreign.Name)
			}
		}
		return fmt.Sprintf("import { %s } from \"%s\";", strings.Join(aliasDefinitions, ", "), filePath)
	}

	// Handle unit alias (e.g., `import * as x from "..."`)
	if importNode.UnitAlias != "" {
		return fmt.Sprintf("import * as %s from \"%s\";", importNode.UnitAlias, filePath)
	}

	// Default case: entire file import
	return fmt.Sprintf("import \"%s\";", filePath)
}

func GenerateTypeDefinition(udtype solc.AstNode) string {
	return fmt.Sprintf("type %s is %s;", udtype.Name, udtype.UnderlyingType.Name)
}

func GenerateFunctionSignature(fn solc.AstNode) string {
	signature := "function "

	// Handle receive function
	if fn.Kind == "receive" {
		return "receive() external payable;"
	}

	// Handle fallback function
	if fn.Kind == "fallback" {
		return "fallback() external payable;"
	}

	// Handle public variables
	if fn.NodeType == "VariableDeclaration" {
		// Start the signature
		signature += fn.Name + "("

		if fn.TypeDescriptions != nil {
			typeString := fn.TypeDescriptions.TypeString

			// Handle mappings
			if strings.HasPrefix(typeString, "mapping(") {
				params, returnType := extractMappingDetails(typeString)
				signature += strings.Join(params, ", ") + ") external view returns (" + returnType + ");"
				return signature
			}

			// Handle non-mapping types
			signature += ") external view"
			returnType := stripContractPrefix(typeString)
			if !isTrivialType(returnType) {
				returnType += " memory"
			}
			signature += " returns (" + returnType + ");"
		}

		return signature
	}

	// Handle constructor
	if fn.Kind == "constructor" {
		fn.Name = "__constructor__"
	}

	// Start regular function signature
	signature += fn.Name + "("

	// Add function parameters
	if fn.Parameters != nil {
		params := []string{}
		for _, param := range fn.Parameters.Parameters {
			paramType := stripContractPrefix(param.TypeDescriptions.TypeString)
			paramName := param.Name
			if paramName == "" {
				paramName = "_"
			}
			params = append(params, fmt.Sprintf("%s %s", paramType, paramName))
		}
		signature += strings.Join(params, ", ")
	}

	signature += ") external"

	// Add state mutability (view/pure/payable)
	if fn.StateMutability == "view" || fn.StateMutability == "pure" || fn.StateMutability == "payable" {
		signature += " " + fn.StateMutability
	}

	// Add return parameters
	if fn.ReturnParameters != nil && len(fn.ReturnParameters.Parameters) > 0 {
		var returns []string
		for _, ret := range fn.ReturnParameters.Parameters {
			returnType := stripContractPrefix(ret.TypeDescriptions.TypeString)
			returns = append(returns, returnType)
		}
		signature += " returns (" + strings.Join(returns, ", ") + ")"
	}

	signature += ";"
	return signature
}

func GenerateEventDefinition(event solc.AstNode) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("event %s(", event.Name))

	if event.Parameters != nil {
		params := []string{}
		for _, param := range event.Parameters.Parameters {
			paramType := param.TypeDescriptions.TypeString
			paramName := param.Name
			if paramName == "" {
				paramName = "_"
			}

			if strings.HasPrefix(paramType, "enum ") {
				paramType = paramType[strings.LastIndex(paramType, ".")+1:]
			}
			if strings.HasPrefix(paramType, "contract ") {
				paramType = paramType[strings.LastIndex(paramType, ".")+1:]
			}
			if strings.HasPrefix(paramType, "struct ") {
				paramType = paramType[strings.LastIndex(paramType, ".")+1:]
			}

			if param.Indexed {
				params = append(params, fmt.Sprintf("%s indexed %s", paramType, paramName))
			} else {
				params = append(params, fmt.Sprintf("%s %s", paramType, paramName))
			}
		}
		builder.WriteString(strings.Join(params, ", "))
	}

	// Close the event definition
	builder.WriteString(");")

	return builder.String()
}

func GenerateErrorDefinition(errorDef solc.AstNode) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("error %s(", errorDef.Name))

	if errorDef.Parameters != nil && len(errorDef.Parameters.Parameters) > 0 {
		for i, param := range errorDef.Parameters.Parameters {
			if i > 0 {
				builder.WriteString(", ")
			}
			paramType := param.TypeDescriptions.TypeString
			paramName := param.Name
			if paramName == "" {
				builder.WriteString(fmt.Sprintf("%s", paramType))
			} else {
				builder.WriteString(fmt.Sprintf("%s %s", paramType, paramName))
			}
		}
	}

	builder.WriteString(");")

	return builder.String()
}

func GenerateStructDefinition(structDef StructDefinition) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("struct %s {\n", structDef.Name))

	for _, member := range structDef.Members {
		builder.WriteString(fmt.Sprintf("\t\t%s %s;\n", member.Type, member.Name))
	}

	builder.WriteString("\t}\n")

	return builder.String()
}

func GenerateEnumSignature(enumDef EnumDefinition) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("enum %s {\n", enumDef.Name))

	for i, member := range enumDef.Members {
		if i > 0 {
			builder.WriteString(",\n")
		}
		builder.WriteString(fmt.Sprintf("\t\t%s", member.Name))
	}

	builder.WriteString("\n\t}\n")

	return builder.String()
}
