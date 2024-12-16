package kurtosis

import (
	"bufio"
	"fmt"
	"strings"
)

// InspectOutput represents the parsed output of "kurtosis enclave inspect"
type InspectOutput struct {
	FileArtifacts []string
	UserServices  map[string]map[string]int
}

// ParseInspectOutput parses the output of "kurtosis enclave inspect" command
func ParseInspectOutput(output string) (*InspectOutput, error) {
	result := &InspectOutput{
		FileArtifacts: make([]string, 0),
		UserServices:  make(map[string]map[string]int),
	}

	scanner := bufio.NewScanner(strings.NewReader(output))

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
				fmt.Printf("New service: %s\n", currentService)

				// Check if there's a port mapping on the same line
				if strings.Contains(line, "->") {
					// Find the position after the service name
					serviceNameEnd := strings.Index(line, currentService) + len(currentService)
					// Process the rest of the line for port mapping
					portLine := line[serviceNameEnd:]
					if strings.Contains(portLine, "->") {
						parts := strings.Split(portLine, "->")
						if len(parts) >= 2 {
							leftPart := strings.TrimRight(parts[0], " \t")
							fmt.Printf("Processing port line from service line. Left part: '%s'\n", leftPart)

							var portName string
							if strings.Contains(leftPart, ":") {
								lastColonIndex := strings.LastIndex(leftPart, ":")
								portName = strings.TrimSpace(leftPart[:lastColonIndex])
							} else {
								fields := strings.Fields(leftPart)
								if len(fields) > 0 {
									portName = fields[0]
								}
							}

							if portName != "" {
								fmt.Printf("Found port name: '%s'\n", portName)
								rightPart := strings.TrimSpace(parts[1])
								rightPart = strings.TrimPrefix(rightPart, "http://")
								if strings.HasPrefix(rightPart, "127.0.0.1:") {
									portStr := strings.TrimPrefix(rightPart, "127.0.0.1:")
									var port int
									_, err := fmt.Sscanf(portStr, "%d", &port)
									if err == nil {
										result.UserServices[currentService][portName] = port
										fmt.Printf("Added port %s:%d for service %s\n", portName, port, currentService)
									}
								}
							}
						}
					}
				}
			} else if strings.Contains(line, "->") {
				// Use original line for port parsing
				parts := strings.Split(line, "->")
				if len(parts) < 2 {
					continue
				}

				leftPart := strings.TrimRight(parts[0], " \t")
				fmt.Printf("Processing port line. Left part: '%s'\n", leftPart)

				var portName string
				if strings.Contains(leftPart, ":") {
					lastColonIndex := strings.LastIndex(leftPart, ":")
					portName = strings.TrimSpace(leftPart[:lastColonIndex])
				} else {
					fields := strings.Fields(leftPart)
					if len(fields) > 0 {
						portName = fields[0]
					}
				}

				if portName == "" {
					continue
				}

				fmt.Printf("Found port name: '%s'\n", portName)

				rightPart := strings.TrimSpace(parts[1])
				rightPart = strings.TrimPrefix(rightPart, "http://")
				if strings.HasPrefix(rightPart, "127.0.0.1:") {
					portStr := strings.TrimPrefix(rightPart, "127.0.0.1:")
					var port int
					_, err := fmt.Sscanf(portStr, "%d", &port)
					if err == nil && currentService != "" {
						result.UserServices[currentService][portName] = port
						fmt.Printf("Added port %s:%d for service %s\n", portName, port, currentService)
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning output: %w", err)
	}

	return result, nil
}
