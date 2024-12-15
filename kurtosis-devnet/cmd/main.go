package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum-optimism/optimism/kurtosis-devnet/pkg/build"
	"github.com/ethereum-optimism/optimism/kurtosis-devnet/pkg/kurtosis"
	"github.com/ethereum-optimism/optimism/kurtosis-devnet/pkg/tmpl"
)

func main() {
	// Define CLI flags
	templateFile := flag.String("template", "", "Path to the template file (required)")
	dataFile := flag.String("data", "", "Path to JSON data file (optional)")
	kurtosisPackage := flag.String("kurtosis-package", kurtosis.DefaultPackageName, "Kurtosis package to deploy")
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
		timestamp := fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
		return fmt.Sprintf("%s-kurtosis-%s", projectName, timestamp)
	}

	artifactsPath := func() string {
		timestamp := fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
		return filepath.Join(baseDir, "contracts", "dist", fmt.Sprintf("artifacts-%s.tar.gz", timestamp))
	}

	contractBuilder := build.NewContractBuilder(
		build.WithContractBaseDir(baseDir),
		build.WithContractBundlePath(artifactsPath()),
		build.WithContractDryRun(*dryRun),
	)

	opts := []tmpl.TemplateContextOptions{
		tmpl.WithFunction("localDockerImage", func(projectName string) (string, error) {
			return dockerBuilder.Build(projectName, imageTag(projectName))
		}),
		tmpl.WithFunction("localContractArtifacts", func(_ string) (string, error) {
			return contractBuilder.Build()
		}),
	}

	// Read and parse the data file if provided
	if *dataFile != "" {
		data, err := os.ReadFile(*dataFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading data file: %v\n", err)
			os.Exit(1)
		}

		var templateData map[string]interface{}
		if err := json.Unmarshal(data, &templateData); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON data: %v\n", err)
			os.Exit(1)
		}

		opts = append(opts, tmpl.WithData(templateData))
	}

	// Open template file
	tmplFile, err := os.Open(*templateFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening template file: %v\n", err)
		os.Exit(1)
	}
	defer tmplFile.Close()

	// Create template context
	ctx := tmpl.NewTemplateContext(opts...)

	// Process template
	buf := bytes.NewBuffer(nil)
	if err := ctx.InstantiateTemplate(tmplFile, buf); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing template: %v\n", err)
		os.Exit(1)
	}

	kurtosisDeployer := kurtosis.NewKurtosisDeployer(
		kurtosis.WithKurtosisBaseDir(baseDir),
		kurtosis.WithKurtosisDryRun(*dryRun),
		kurtosis.WithKurtosisPackageName(*kurtosisPackage),
	)

	kurtosisDeployer.Deploy(buf)
}
