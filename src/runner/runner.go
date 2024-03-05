// Package to execute tasks on the system. It uses the os/exec package to execute the tasks.
package runner

import (
	"fmt"
	"os/exec"
)

type Runner interface {
	Execute() error
}

type Execution struct {
	TaskName       string
	TaskParameters map[string]interface{}
}

func NewExecution(taskName string, taskParameters map[string]interface{}) *Execution {
	return &Execution{
		TaskName:       taskName,
		TaskParameters: taskParameters,
	}
}

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
