package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ethereum-optimism/optimism/kurtosis-devnet/pkg/build"
	"github.com/ethereum-optimism/optimism/kurtosis-devnet/pkg/kurtosis"
	"github.com/ethereum-optimism/optimism/kurtosis-devnet/pkg/serve"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// we will serve content from this tmpDir for the duration of the devnet creation
	tmpDir, err := os.MkdirTemp("", "contracts-bundle")
	if err != nil {
		log.Fatalf("Error creating temporary directory: %v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	server := serve.NewServer(tmpDir)
	if err := server.Start(ctx); err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
	defer server.Stop(ctx)

	baseDir := filepath.Dir(*templateFile)

	dockerBuilder := build.NewDockerBuilder(
		build.WithDockerBaseDir(baseDir),
		build.WithDockerDryRun(*dryRun),
	)

	imageTag := func(projectName string) string {
		return fmt.Sprintf("%s:%s", projectName, *enclave)
	}

	contractsBundle := fmt.Sprintf("contracts-bundle-%s.tar.gz", *enclave)
	contractsBundlePath := func(_ string) string {
		return filepath.Join(tmpDir, contractsBundle)
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
			err := contractBuilder.Build(layer, contractsBundlePath(layer))
			if err != nil {
				return "", err
			}
			url := fmt.Sprintf("%s/%s", server.URL(), contractsBundle)
			log.Printf("Contract artifacts available at: %s\n", url)
			return url, nil
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
	tmplCtx := tmpl.NewTemplateContext(opts...)

	// Process template
	buf := bytes.NewBuffer(nil)
	if err := tmplCtx.InstantiateTemplate(tmplFile, buf); err != nil {
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
