// Package runner executes tasks on the system using the os/exec package.
package runner

import (
	"fmt"
	"os/exec"
)

// Runner is an interface that requires an Execute method.
type Runner interface {
	// Execute runs the task and returns an error if any.
	Execute() error
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
func (e *Execution) Execute() ([]byte, error) {
	var args []string
	for _, arg := range e.CommandParams["args"].([]interface{}) {
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

	cmd := exec.Command(e.CommandName, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error executing process task: %v", err)
	}

	return output, nil
}
