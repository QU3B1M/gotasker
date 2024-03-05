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
	// TaskName is the name of the task.
	TaskName string
	// TaskParameters are the parameters for the task.
	TaskParameters map[string]interface{}
}

// NewExecution creates a new Execution with the given task name and parameters.
func NewExecution(taskName string, taskParameters map[string]interface{}) *Execution {
	return &Execution{
		TaskName:       taskName,
		TaskParameters: taskParameters,
	}
}

// Execute runs the task with its parameters. It returns an error if the execution fails.
func (e *Execution) Execute() error {
	var taskArgs []string
	for _, arg := range e.TaskParameters["args"].([]interface{}) {
		switch v := arg.(type) {
		case string:
			taskArgs = append(taskArgs, v)
		case map[string]interface{}:
			for key, value := range v {
				switch val := value.(type) {
				case []interface{}:
					for _, argvalue := range val {
						taskArgs = append(taskArgs, fmt.Sprintf("--%s=%v", key, argvalue))
					}
				default:
					taskArgs = append(taskArgs, fmt.Sprintf("--%s=%v", key, val))
				}
			}
		}
	}

	cmd := exec.Command(e.TaskParameters["path"].(string), taskArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing process task: %v", err)
	}

	fmt.Printf("Finish task: %s\n", e.TaskName)
	fmt.Println(string(output))

	return nil
}
