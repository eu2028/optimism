package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ethereum-optimism/optimism/kurtosis-devnet/pkg/build"
	"github.com/ethereum-optimism/optimism/kurtosis-devnet/pkg/kurtosis"
	"github.com/ethereum-optimism/optimism/kurtosis-devnet/pkg/tmpl"
)

func main() {
	// Define CLI flags
	templateFile := flag.String("template", "", "Path to the template file (required)")
	dataFile := flag.String("data", "", "Path to JSON data file (optional)")
	kurtosisPackage := flag.String("kurtosis-package", kurtosis.DefaultPackageName, "Kurtosis package to deploy")
	enclave := flag.String("enclave", kurtosis.DefaultEnclave, "Enclave name")
	environment := flag.String("environment", "", "Path to JSON environment file output (optional)")
	dryRun := flag.Bool("dry-run", false, "Dry run mode")
	flag.Parse()

	// Validate required flags
	if *templateFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	baseDir := filepath.Dir(*templateFile)

	dockerBuilder := build.NewDockerBuilder(
		build.WithDockerBaseDir(baseDir),
		build.WithDockerDryRun(*dryRun),
	)

	imageTag := func(projectName string) string {
		return fmt.Sprintf("%s:%s", projectName, *enclave)
	}

	contractBuilder := build.NewContractBuilder(
		build.WithContractBaseDir(baseDir),
		build.WithContractDryRun(*dryRun),
	)

	opts := []tmpl.TemplateContextOptions{
		tmpl.WithFunction("localDockerImage", func(projectName string) (string, error) {
			return dockerBuilder.Build(projectName, imageTag(projectName))
		}),
		tmpl.WithFunction("localContractArtifacts", func(layer string) (string, error) {
			return contractBuilder.Build(layer)
		}),
	}

	// Read and parse the data file if provided
	if *dataFile != "" {
		data, err := os.ReadFile(*dataFile)
		if err != nil {
			log.Fatalf("Error reading data file: %v\n", err)
		}

		var templateData map[string]interface{}
		if err := json.Unmarshal(data, &templateData); err != nil {
			log.Fatalf("Error parsing JSON data: %v\n", err)
		}

		opts = append(opts, tmpl.WithData(templateData))
	}

	// Open template file
	tmplFile, err := os.Open(*templateFile)
	if err != nil {
		log.Fatalf("Error opening template file: %v\n", err)
	}
	defer tmplFile.Close()

	// Create template context
	ctx := tmpl.NewTemplateContext(opts...)

	// Process template
	buf := bytes.NewBuffer(nil)
	if err := ctx.InstantiateTemplate(tmplFile, buf); err != nil {
		log.Fatalf("Error processing template: %v\n", err)
	}

	kurtosisDeployer := kurtosis.NewKurtosisDeployer(
		kurtosis.WithKurtosisBaseDir(baseDir),
		kurtosis.WithKurtosisDryRun(*dryRun),
		kurtosis.WithKurtosisPackageName(*kurtosisPackage),
		kurtosis.WithKurtosisEnclave(*enclave),
	)

	env, err := kurtosisDeployer.Deploy(buf)
	if err != nil {
		log.Fatalf("Error deploying kurtosis: %v\n", err)
	}

	envOutput := os.Stdout
	if *environment != "" {
		envOutput, err = os.Create(*environment)
		if err != nil {
			log.Fatalf("Error creating environment file: %v\n", err)
		}
	} else {
		log.Println("Environment description:")
	}

	enc := json.NewEncoder(envOutput)
	enc.SetIndent("", "  ")
	if err := enc.Encode(env); err != nil {
		log.Fatalf("Error encoding environment: %v\n", err)
	}
}
