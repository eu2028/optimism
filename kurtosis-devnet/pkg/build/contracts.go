package build

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"text/template"
)

// ContractBuilder handles building smart contracts using just commands
type ContractBuilder struct {
	// Base directory where the build commands should be executed
	baseDir string
	// Template for the build command
	cmdTemplate *template.Template
	// Path where the contract bundle will be created
	bundlePath string
	// Dry run mode
	dryRun bool
}

const (
	contractsCmdTemplateStr = "just CONTRACTS_BUNDLE={{.BundlePath}} contracts/build"
	defaultBundlePath       = "contracts-bundle.tar.gz"
)

var defaultContractTemplate *template.Template

func init() {
	defaultContractTemplate = template.Must(template.New("contract_build_cmd").Parse(contractsCmdTemplateStr))
}

type ContractBuilderOptions func(*ContractBuilder)

func WithContractBaseDir(baseDir string) ContractBuilderOptions {
	return func(b *ContractBuilder) {
		b.baseDir = baseDir
	}
}

func WithContractTemplate(cmdTemplate *template.Template) ContractBuilderOptions {
	return func(b *ContractBuilder) {
		b.cmdTemplate = cmdTemplate
	}
}

func WithContractBundlePath(bundlePath string) ContractBuilderOptions {
	return func(b *ContractBuilder) {
		b.bundlePath = bundlePath
	}
}

func WithContractDryRun(dryRun bool) ContractBuilderOptions {
	return func(b *ContractBuilder) {
		b.dryRun = dryRun
	}
}

// NewContractBuilder creates a new ContractBuilder instance
func NewContractBuilder(opts ...ContractBuilderOptions) *ContractBuilder {
	b := &ContractBuilder{
		baseDir:     ".",
		cmdTemplate: defaultContractTemplate,
		bundlePath:  defaultBundlePath,
		dryRun:      false,
	}

	for _, opt := range opts {
		opt(b)
	}

	// Ensure bundlePath is absolute
	if !filepath.IsAbs(b.bundlePath) {
		b.bundlePath = filepath.Join(b.baseDir, b.bundlePath)
	}

	return b
}

// templateData holds the data for the command template
type contractTemplateData struct {
	BundlePath string
}

// Build executes the contract build command
func (b *ContractBuilder) Build() (string, error) {
	// Prepare template data
	data := contractTemplateData{
		BundlePath: b.bundlePath,
	}

	// Execute template to get command string
	var cmdBuf bytes.Buffer
	if err := b.cmdTemplate.Execute(&cmdBuf, data); err != nil {
		return "", fmt.Errorf("failed to execute command template: %w", err)
	}

	// Create command
	cmd := exec.Command("sh", "-c", cmdBuf.String())
	cmd.Dir = b.baseDir

	if b.dryRun {
		return b.bundlePath, nil
	}

	// Capture output and error
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("contract build command failed: %w\nOutput: %s", err, string(output))
	}

	// Return the bundle path as confirmation of successful build
	return b.bundlePath, nil
}
