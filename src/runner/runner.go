// Package runner executes tasks on the system using the os/exec package.
package runner

import (
	"fmt"
	"os/exec"
)

// Runner is an interface that requires an Execute method.
type Runner interface {
	// Execute runs the task and returns the output or an error if any.
	Execute() (string, error)
}

// Execution represents a task to be executed.
type Execution struct {
	// CommandName is the name of the command to be executed.
	CommandName string
	// CommandParams are the parameters for the task.
	CommandParams map[string]interface{}
}

// NewExecution creates a new Execution with the given task name and parameters.
func NewExecution(name string, parameters map[string]interface{}) *Execution {
	return &Execution{
		CommandName:   name,
		CommandParams: parameters,
	}
}

// Execute runs the task with its parameters. It returns an error if the execution fails.
func (e *Execution) Execute() (string, error) {
	var args []string

	// Determine the command binary: prefer "path" from params, fallback to CommandName.
	cmdBinary := e.CommandName
	if path, ok := e.CommandParams["path"].(string); ok && path != "" {
		cmdBinary = path
	}

	if rawArgs, ok := e.CommandParams["args"]; ok && rawArgs != nil {
		if argList, ok := rawArgs.([]interface{}); ok {
			for _, arg := range argList {
				switch v := arg.(type) {
				case string:
					args = append(args, v)
				case map[string]interface{}:
					for key, value := range v {
						switch val := value.(type) {
						case []interface{}:
							for _, argvalue := range val {
								args = append(args, fmt.Sprintf("--%s=%v", key, argvalue))
							}
						default:
							args = append(args, fmt.Sprintf("--%s=%v", key, val))
						}
					}
				}
			}
		}
	}

	cmd := exec.Command(cmdBinary, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error executing process task: %v", err)
	}

	return string(output[:]), nil
}
