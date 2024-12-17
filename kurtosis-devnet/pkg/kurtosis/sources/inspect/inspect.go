package inspect

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type PortMap map[string]int

type ServiceMap map[string]PortMap

// InspectOutput represents the parsed output of "kurtosis enclave inspect"
type InspectOutput struct {
	FileArtifacts []string
	UserServices  ServiceMap
}

// extractPortName extracts the port name from the left part of a port mapping
func extractPortName(leftPart string) string {
	if strings.Contains(leftPart, ":") {
		lastColonIndex := strings.LastIndex(leftPart, ":")
		return strings.TrimSpace(leftPart[:lastColonIndex])
	}

	fields := strings.Fields(leftPart)
	if len(fields) > 0 {
		return fields[0]
	}
	return ""
}

// extractPort extracts the port number from the right part of a port mapping
func extractPort(rightPart string) (int, error) {
	rightPart = strings.TrimSpace(rightPart)
	rightPart = strings.TrimPrefix(rightPart, "http://")
	if !strings.HasPrefix(rightPart, "127.0.0.1:") {
		return 0, fmt.Errorf("invalid port mapping format")
	}

	portStr := strings.TrimPrefix(rightPart, "127.0.0.1:")
	var port int
	_, err := fmt.Sscanf(portStr, "%d", &port)
	return port, err
}

// parsePortMapping parses a port mapping string and adds it to the result
func parsePortMapping(line string, currentService string, result *InspectOutput) {
	parts := strings.Split(line, "->")
	if len(parts) < 2 {
		return
	}

	leftPart := strings.TrimRight(parts[0], " \t")

	portName := extractPortName(leftPart)
	if portName == "" {
		return
	}

	port, err := extractPort(parts[1])
	if err == nil && currentService != "" {
		result.UserServices[currentService][portName] = port
	}
}

// ParseInspectOutput parses the output of "kurtosis enclave inspect" command
func ParseInspectOutput(r io.Reader) (*InspectOutput, error) {
	result := &InspectOutput{
		FileArtifacts: make([]string, 0),
		UserServices:  make(ServiceMap),
	}

	scanner := bufio.NewScanner(r)

	// States for parsing different sections
	const (
		None = iota
		Files
		Services
	)

	state := None
	var currentService string

	for scanner.Scan() {
		line := scanner.Text()
		// Only trim for section detection
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			continue
		}

		// Check section headers using trimmed line
		if strings.Contains(trimmedLine, "Files Artifacts") {
			state = Files
			continue
		}
		if strings.Contains(trimmedLine, "User Services") {
			state = Services
			continue
		}

		// Skip header lines
		if strings.HasPrefix(trimmedLine, "UUID") || strings.HasPrefix(trimmedLine, "====") {
			continue
		}

		switch state {
		case Files:
			fields := strings.Fields(trimmedLine)
			if len(fields) >= 2 {
				result.FileArtifacts = append(result.FileArtifacts, fields[1])
			}

		case Services:
			fields := strings.Fields(trimmedLine)
			if len(fields) == 0 {
				continue
			}

			// If line starts with UUID, it's a new service
			if len(fields) >= 2 && len(fields[0]) == 12 {
				currentService = fields[1]
				result.UserServices[currentService] = make(map[string]int)

				// Check if there's a port mapping on the same line
				if strings.Contains(line, "->") {
					// Find the position after the service name
					serviceNameEnd := strings.Index(line, currentService) + len(currentService)
					// Process the rest of the line for port mapping
					portLine := line[serviceNameEnd:]
					if strings.Contains(portLine, "->") {
						parsePortMapping(portLine, currentService, result)
					}
				}
			} else if strings.Contains(line, "->") {
				parsePortMapping(line, currentService, result)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning output: %w", err)
	}

	return result, nil
}
