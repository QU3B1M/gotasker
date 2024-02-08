
// This Go code does the same thing as your Python code. It defines a WorkflowFile struct with methods to
// load a workflow from a file, process the workflow into a collection of tasks, and expand tasks with
// variable values. The replacePlaceholders function is used to replace placeholders in a task with actual
// variable values. The product function is used to generate all combinations of variable values for tasks
// with a ‘foreach’ field.

// Please replace "./workflow.yaml" with your actual workflow file path. Also, replace "./schemas/schema_v1.json"
// with your actual schema file path if you have one.

// Remember that error handling in Go is explicit, and it’s a good practice to always check for errors where
// they can occur.


package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type WorkflowFile struct {
	SchemaPath       string
	WorkflowRawData  map[string]interface{}
	TaskCollection   []map[string]interface{}
}

func NewWorkflowFile(workflowFilePath string, schemaPath string) *WorkflowFile {
	if schemaPath == "" {
		schemaPath = "./schemas/schema_v1.json"
	}
	workflowRawData := loadWorkflow(workflowFilePath)
	taskCollection := processWorkflow(workflowRawData)
	return &WorkflowFile{
		SchemaPath:      schemaPath,
		WorkflowRawData: workflowRawData,
		TaskCollection:  taskCollection,
	}
}

func loadWorkflow(filePath string) map[string]interface{} {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}
	data := make(map[string]interface{})
	err = yaml.Unmarshal(file, &data)
	if err != nil {
		fmt.Println("Error parsing YAML:", err)
		os.Exit(1)
	}
	return data
}

func processWorkflow(workflowRawData map[string]interface{}) []map[string]interface{} {
	taskCollection := []map[string]interface{}{}
	variables := workflowRawData["variables"].(map[string]interface{})
	tasks := workflowRawData["tasks"].([]interface{})
	for _, task := range tasks {
		taskMap := task.(map[string]interface{})
		expandedTasks := expandTask(taskMap, variables)
		taskCollection = append(taskCollection, expandedTasks...)
	}
	return taskCollection
}

func replacePlaceholders(element interface{}, values map[string]interface{}) interface{} {
	switch v := element.(type) {
	case map[string]interface{}:
		newMap := make(map[string]interface{})
		for key, value := range v {
			newMap[key] = replacePlaceholders(value, values)
		}
		return newMap
	case []interface{}:
		newSlice := make([]interface{}, len(v))
		for i, subElement := range v {
			newSlice[i] = replacePlaceholders(subElement, values)
		}
		return newSlice
	case string:
		for key, value := range values {
			v = strings.ReplaceAll(v, "{"+key+"}", fmt.Sprint(value))
		}
		return v
	default:
		return v
	}
}

func expandTask(task map[string]interface{}, variables map[string]interface{}) []map[string]interface{} {
	expandedTasks := []map[string]interface{}{}

	if _, ok := task["foreach"]; ok {
		loopVariables := task["foreach"].([]interface{})

		variableNames := make([]string, len(loopVariables))
		asIdentifiers := make([]string, len(loopVariables))
		variableValues := make([][]interface{}, len(loopVariables))

		for i, loopVariableData := range loopVariables {
			loopVariableDataMap := loopVariableData.(map[string]interface{})
			variableNames[i] = loopVariableDataMap["variable"].(string)
			asIdentifiers[i] = loopVariableDataMap["as"].(string)
			variableValues[i] = variables[variableNames[i]].([]interface{})
		}

		for _, combination := range product(variableValues) {
			variablesWithItems := make(map[string]interface{})
			for key, value := range variables {
				variablesWithItems[key] = value
			}
			for i, value := range combination {
				variablesWithItems[asIdentifiers[i]] = value
			}
			expandedTasks = append(expandedTasks, replacePlaceholders(task, variablesWithItems).(map[string]interface{}))
		}
	} else {
		expandedTasks = append(expandedTasks, replacePlaceholders(task, variables).(map[string]interface{}))
	}

	return expandedTasks
}

func product(arrays [][]interface{}) [][]interface{} {
	length := len(arrays)
	if length == 0 {
		return [][]interface{}{}
	} else if length == 1 {
		result := make([][]interface{}, len(arrays[0]))
		for i, value := range arrays[0] {
			result[i] = []interface{}{value}
		}
		return result
	} else {
		results := [][]interface{}{}
		for _, value := range arrays[0] {
			for _, subProduct := range product(arrays[1:]) {
				result := append([]interface{}{value}, subProduct...)
				results = append(results, result)
			}
		}
		return results
	}
}


// This Go code does the same thing as your Python code. It defines StaticWorkflowValidation, CheckDuplicatedTasks,
// and CheckNotExistingTasks methods on the WorkflowFile struct. The StaticWorkflowValidation method calls the other
// two methods to perform static validation of the workflow. The CheckDuplicatedTasks method checks for duplicated
// task names, and the CheckNotExistingTasks method checks for tasks that do not exist.

// Please replace "./workflow.yaml" with your actual workflow file path. Also, replace "./schemas/schema_v1.json"
// with your actual schema file path if you have one.

// Remember that error handling in Go is explicit, and it’s a good practice to always check for errors where they
// can occur.

func (w *WorkflowFile) StaticWorkflowValidation() {
	w.CheckDuplicatedTasks()
	w.CheckNotExistingTasks()
}

func (w *WorkflowFile) CheckDuplicatedTasks() {
	taskNameCounts := make(map[string]int)
	for _, task := range w.TaskCollection {
		taskName := task["task"].(string)
		taskNameCounts[taskName]++
	}

	var duplicates []string
	for name, count := range taskNameCounts {
		if count > 1 {
			duplicates = append(duplicates, name)
		}
	}

	if len(duplicates) > 0 {
		fmt.Printf("Duplicated task names: %s\n", strings.Join(duplicates, ", "))
	}
}

func (w *WorkflowFile) CheckNotExistingTasks() {
	taskNames := make(map[string]bool)
	for _, task := range w.TaskCollection {
		taskName := task["task"].(string)
		taskNames[taskName] = true
	}

	for _, task := range w.TaskCollection {
		if dependencies, ok := task["depends-on"].([]interface{}); ok {
			for _, dependency := range dependencies {
				if _, ok := taskNames[dependency.(string)]; !ok {
					fmt.Printf("Tasks do not exist: %s\n", dependency)
				}
			}
		}
	}
}


func main() {
	workflowFile := NewWorkflowFile("./workflow.yaml", "")
	fmt.Println(workflowFile.TaskCollection)
}


package main

import (
	"fmt"
)

func main() {
	workflowFile := NewWorkflowFile("./workflow.yaml", "")
	workflowFile.StaticWorkflowValidation()
}
