package kurtosis

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"text/template"
)

const DefaultPackageName = "github.com/ethpandaops/optimism-package"

// KurtosisEnvironment represents the output of a Kurtosis deployment
// Note: This is a placeholder - we'll define the actual structure later
type KurtosisEnvironment struct {
	// TODO: Add environment details
}

// KurtosisDeployer handles deploying packages using Kurtosis
type KurtosisDeployer struct {
	// Base directory where the deployment commands should be executed
	baseDir string
	// Template for the deployment command
	cmdTemplate *template.Template
	// Package name to deploy
	packageName string
	// Dry run mode
	dryRun bool
	// Enclave name
	enclave string
}

const cmdTemplateStr = "just _kurtosis-run {{.PackageName}} {{.ArgFile}} {{.Enclave}}"

var defaultCmdTemplate *template.Template

func init() {
	defaultCmdTemplate = template.Must(template.New("kurtosis_deploy_cmd").Parse(cmdTemplateStr))
}

type KurtosisDeployerOptions func(*KurtosisDeployer)

func WithKurtosisBaseDir(baseDir string) KurtosisDeployerOptions {
	return func(d *KurtosisDeployer) {
		d.baseDir = baseDir
	}
}

func WithKurtosisCmdTemplate(cmdTemplate *template.Template) KurtosisDeployerOptions {
	return func(d *KurtosisDeployer) {
		d.cmdTemplate = cmdTemplate
	}
}

func WithKurtosisPackageName(packageName string) KurtosisDeployerOptions {
	return func(d *KurtosisDeployer) {
		d.packageName = packageName
	}
}

func WithKurtosisDryRun(dryRun bool) KurtosisDeployerOptions {
	return func(d *KurtosisDeployer) {
		d.dryRun = dryRun
	}
}

func WithKurtosisEnclave(enclave string) KurtosisDeployerOptions {
	return func(d *KurtosisDeployer) {
		d.enclave = enclave
	}
}

// NewKurtosisDeployer creates a new KurtosisDeployer instance
func NewKurtosisDeployer(opts ...KurtosisDeployerOptions) *KurtosisDeployer {
	d := &KurtosisDeployer{
		baseDir:     ".",
		cmdTemplate: defaultCmdTemplate,
		packageName: DefaultPackageName,
		dryRun:      false,
		enclave:     "devnet",
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

// templateData holds the data for the command template
type templateData struct {
	PackageName string
	ArgFile     string
	Enclave     string
}

// Deploy executes the Kurtosis deployment command with the provided input
func (d *KurtosisDeployer) Deploy(input io.Reader) (*KurtosisEnvironment, error) {
	// Create temporary file for arguments
	argFile, err := os.CreateTemp("", "kurtosis-args-*.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary arg file: %w", err)
	}
	defer os.Remove(argFile.Name())

	var writer io.Writer = argFile
	writer = io.MultiWriter(writer, os.Stdout)

	log.Print("Running Kurtosis with the following arguments:")
	// Copy input to arg file
	if _, err := io.Copy(writer, input); err != nil {
		return nil, fmt.Errorf("failed to write arg file: %w", err)
	}

	// Prepare template data
	data := templateData{
		PackageName: d.packageName,
		ArgFile:     argFile.Name(),
		Enclave:     d.enclave,
	}
	argFile.Close()

	// Execute template to get command string
	var cmdBuf bytes.Buffer
	if err := d.cmdTemplate.Execute(&cmdBuf, data); err != nil {
		return nil, fmt.Errorf("failed to execute command template: %w", err)
	}

	// Create command
	cmd := exec.Command("sh", "-c", cmdBuf.String())
	cmd.Dir = d.baseDir

	if d.dryRun {
		fmt.Println("Dry run mode enabled, kurtosis would run the following command:")
		fmt.Println(cmdBuf.String())
		return &KurtosisEnvironment{}, nil
	}

	// Stream output to stdout and stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("kurtosis deployment failed: %w", err)
	}

	// TODO: Populate KurtosisEnvironment with the actual environment details
	env := &KurtosisEnvironment{}

	return env, nil
}
