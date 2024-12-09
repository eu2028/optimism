package main

import (
	"errors"
	"fmt"
	"github.com/ethereum-optimism/optimism/op-chain-ops/solc"
	"github.com/ethereum-optimism/optimism/packages/contracts-bedrock/scripts/checks/common"
	"os"
	"strings"
)

type ContractData struct {
	Imports   []solc.AstNode
	Functions []solc.AstNode
	Events    []solc.AstNode
	Errors    []solc.AstNode
	Types     []solc.AstNode
	Structs   []StructDefinition
	Enums     []EnumDefinition
	Version   string
}

type StructDefinition struct {
	Name    string
	Members []StructMember
}

type StructMember struct {
	Name string
	Type string
}

type EnumDefinition struct {
	Name    string
	Members []EnumMember
}

type EnumMember struct {
	Name string
}

func ExtractASTData(ast solc.Ast, inherited bool, version string) ContractData {
	var contractData ContractData

	for _, node := range ast.Nodes {
		switch node.NodeType {
		case "PragmaDirective":
			if !inherited {
				version = fmt.Sprint(node.Literals[1], node.Literals[2])
				contractData.Version = fmt.Sprint(node.Literals[1], node.Literals[2])
			}
		case "ImportDirective":
			contractData.Imports = append(contractData.Imports, node)
		case "ContractDefinition":
			extractedData := ExtractContractASTData(node, inherited)
			contractData.Functions = append(contractData.Functions, extractedData.Functions...)
			contractData.Events = append(contractData.Events, extractedData.Events...)
			contractData.Errors = append(contractData.Errors, extractedData.Errors...)
			contractData.Types = append(contractData.Types, extractedData.Types...)
			contractData.Structs = append(contractData.Structs, extractedData.Structs...)
			contractData.Enums = append(contractData.Enums, extractedData.Enums...)
			contractData.Imports = append(contractData.Imports, extractedData.Imports...)
			for i := 0; i < len(node.BaseContracts); i++ {
				var artifact *solc.ForgeArtifact

				versionedPath := fmt.Sprintf("packages/contracts-bedrock/forge-artifacts/%s.sol/%s.%s.json", node.BaseContracts[i].BaseName.Name,
					node.BaseContracts[i].BaseName.Name, version)

				if _, err := os.Stat(versionedPath); err == nil {
					artifact, _ = common.ReadForgeArtifact(versionedPath)

				} else if errors.Is(err, os.ErrNotExist) {
					artifact, _ = common.ReadForgeArtifact(fmt.Sprintf(
						"packages/contracts-bedrock/forge-artifacts/%s.sol/%s.json",
						node.BaseContracts[i].BaseName.Name,
						node.BaseContracts[i].BaseName.Name,
					))

				}

				data := ExtractASTData(artifact.Ast, true, version)
				contractData.Functions = append(contractData.Functions, data.Functions...)
				contractData.Events = append(contractData.Events, data.Events...)
				contractData.Errors = append(contractData.Errors, data.Errors...)
				contractData.Types = append(contractData.Types, data.Types...)
				contractData.Structs = append(contractData.Structs, data.Structs...)
				contractData.Enums = append(contractData.Enums, data.Enums...)
				//contractData.Imports = append(contractData.Imports, data.Imports...)
			}
		case "ErrorDefinition":
			contractData.Errors = append(contractData.Errors, node)
		case "UserDefinedValueTypeDefinition":
			contractData.Types = append(contractData.Types, node)
		}
	}

	return contractData
}

func ExtractContractASTData(node solc.AstNode, inherited bool) ContractData {
	var data ContractData
	for _, innerNode := range node.Nodes {
		switch innerNode.NodeType {
		case "FunctionDefinition":
			if inherited && innerNode.Kind == "constructor" {
				continue
			}
			if innerNode.Visibility == "public" || innerNode.Visibility == "external" {
				data.Functions = append(data.Functions, innerNode)
			}
		case "VariableDeclaration":
			if innerNode.Visibility == "public" {
				data.Functions = append(data.Functions, innerNode)
			}

		case "EventDefinition":
			data.Events = append(data.Events, innerNode)
		case "ErrorDefinition":
			data.Errors = append(data.Errors, innerNode)
		case "StructDefinition":
			structDef := StructDefinition{
				Name: innerNode.Name,
			}

			for _, member := range innerNode.Members {
				memberType := member.TypeDescriptions.TypeString

				idx := strings.Index(memberType, ".")
				if idx != -1 {

					memberType = memberType[idx+1:]
				}

				memberName := member.Name
				structDef.Members = append(structDef.Members, StructMember{
					Name: memberName,
					Type: memberType,
				})
			}

			data.Structs = append(data.Structs, structDef)
		case "EnumDefinition":
			enumDef := EnumDefinition{
				Name: innerNode.Name,
			}

			for _, member := range innerNode.Members {
				enumDef.Members = append(enumDef.Members, EnumMember{
					Name: member.Name,
				})
			}

			data.Enums = append(data.Enums, enumDef)
		case "UserDefinedValueTypeDefinition":
			data.Types = append(data.Types, innerNode)
		}
	}
	return data
}
