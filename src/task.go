// This Go code does the same thing as your Python code. It defines a Task interface with an Execute method,
// and a ProcessTask struct that implements this interface. The Execute method of ProcessTask builds the
// command arguments and runs the command using the os/exec package. If the command fails, it returns an error.

// Please replace "/path/to/your/command" and "arg1" with your actual command and arguments. Also, replace
// "key" and []interface{}{"value1", "value2"} with your actual key and values.

package main

import (
	"fmt"
	"os/exec"
)

type Task interface {
	Execute() error
}

type ProcessTask struct {
	TaskName       string
	TaskParameters map[string]interface{}
}

func NewProcessTask(taskName string, taskParameters map[string]interface{}) *ProcessTask {
	return &ProcessTask{
		TaskName:       taskName,
		TaskParameters: taskParameters,
	}
}

func (p *ProcessTask) Execute() error {
	var taskArgs []string
	for _, arg := range p.TaskParameters["args"].([]interface{}) {
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

	cmd := exec.Command(p.TaskParameters["path"].(string), taskArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing process task: %v", err)
	}

	fmt.Printf("Finish task: %s\n", p.TaskName)
	fmt.Println(string(output))

	return nil
}

// func main() {
// 	taskParameters := map[string]interface{}{
// 		"path": "/path/to/your/command",
// 		"args": []interface{}{
// 			"arg1",
// 			map[string]interface{}{
// 				"key": []interface{}{"value1", "value2"},
// 			},
// 		},
// 	}

// 	task := NewProcessTask("MyTask", taskParameters)
// 	if err := task.Execute(); err != nil {
// 		fmt.Printf("Task execution failed: %v\n", err)
// 	}
// }
