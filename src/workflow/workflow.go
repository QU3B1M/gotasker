// This Go code does the same thing as your Python code. It defines a Workflow struct with methods to
// load a workflow from a file, process the workflow into a collection of tasks, and expand tasks with
// variable values. The ReplacePlaceholders function is used to replace placeholders in a task with actual
// variable values. The product function is used to generate all combinations of variable values for tasks
// with a ‘foreach’ field.

// Please replace "./workflow.yaml" with your actual workflow file path. Also, replace "./schemas/schema_v1.json"
// with your actual json file path if you have one.

package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"gopkg.in/yaml.v2"
)

type Task struct {
	Name      string    `json:"name"`
	Do        Action    `json:"do"`
	Cleanup   Action    `json:"cleanup"`
	DependsOn []string  `json:"depends-on"`
	ForEach   []ForEach `json:"foreach"`
}

type Action struct {
	This string `json:"this"`
	With With   `json:"with"`
}

type With struct {
	This string        `json:"this"`
	Args []interface{} `json:"args"`
	Path string        `json:"path"`
}

type ForEach struct {
	Variable string    `json:"variable"`
	As       string    `json:"as"`
	ForEach  []ForEach `json:"foreach"`
}

type Workflow struct {
	Tasks     []Task      `json:"tasks"`
	Variables interface{} `json:"variables"`
}

// NewWorkflow loads a workflow from a file, processes it,
// and returns a pointer to a Workflow struct. It can return an error
// if there's a problem with marshalling or unmarshalling the data.
func NewWorkflow(workflowFilePath string) (*Workflow, error) {
	var workflow Workflow

	workflowData, err := loadWorkflowFile(workflowFilePath)
	if err != nil {
		return nil, fmt.Errorf("error loading workflow: %w", err)
	}

	taskCollection, err := ProcessWorkflow(workflowData)
	if err != nil {
		return nil, fmt.Errorf("error processing workflow: %w", err)
	}

	// Create a map with the workflow data
	mapWorkflow := map[string]interface{}{
		"variables": workflowData["variables"],
		"tasks":     taskCollection,
	}

	// Convert the map to JSON
	jsonData, err := json.Marshal(mapWorkflow)
	if err != nil {
		return nil, fmt.Errorf("error converting to JSON: %w", err)
	}

	// Convert the JSON to a struct
	err = json.Unmarshal(jsonData, &workflow)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	return &workflow, nil
}

// loadWorkflowFile reads a workflow from a file, parses it from YAML to JSON,
// converts all keys to strings, and returns the result as a map.
// It can return an error if there's a problem with reading the file,
// parsing the YAML, or converting the keys.
func loadWorkflowFile(filePath string) (map[string]interface{}, error) {
	json := make(map[string]interface{})
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	err = yaml.Unmarshal(file, &json)
	if err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	converted, ok := ConvertKeysToString(json).(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("error converting keys to string")
	}

	return converted, nil
}

func ProcessWorkflow(workflowRawData map[string]interface{}) ([]map[string]interface{}, error) {
	taskCollection := []map[string]interface{}{}

	// Convert the interface keys to string and separate the variables from the tasks
	variables, ok := workflowRawData["variables"].(map[string]interface{})
	if !ok {
		fmt.Println("Error parsing variables.")
		return nil, fmt.Errorf("error parsing variables")
	}
	tasks, ok := workflowRawData["tasks"].([]interface{})
	if !ok {
		fmt.Println("Error parsing tasks.")
		return nil, fmt.Errorf("error parsing tasks")
	}

	// Analyze the workflow data and creates the corresponding tasks.
	for _, task := range tasks {
		if _, ok := task.(map[string]interface{})["foreach"]; ok {
			newTasks := ExpandTask(task, variables)
			taskCollection = append(taskCollection, newTasks...)
		} else {
			// This task does not have a 'foreach' field, so we just need to replace the placeholders.
			taskToAdd := ReplacePlaceholders(task, variables)
			taskCollection = append(taskCollection, taskToAdd.(map[string]interface{}))
		}
	}

	return taskCollection, nil
}

// ConvertKeysToString recursively converts all keys in the input
// to strings if they are not already.
func ConvertKeysToString(item interface{}) interface{} {
	switch x := item.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			ks, ok := k.(string)
			if !ok {
				fmt.Println("Key" + fmt.Sprint(k) + "is not string type. Forcing conversion.")
				ks = fmt.Sprint(k)
			}
			m2[ks] = ConvertKeysToString(v)
		}
		return m2
	case map[string]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k] = ConvertKeysToString(v)
		}
		return m2
	case []interface{}:
		i2 := make([]interface{}, len(x))
		for i, v := range x {
			i2[i] = ConvertKeysToString(v)
		}
		return i2
	case []map[interface{}]interface{}:
		i2 := make([]map[string]interface{}, len(x))
		for i, v := range x {
			if vm, ok := ConvertKeysToString(v).(map[string]interface{}); ok {
				i2[i] = vm
			} else {
				fmt.Println("Value" + fmt.Sprint(v) + "is not map[string]interface{} type. Skipping.")
			}
		}
		return i2
	}
	return item
}

func ExpandTask(task interface{}, variables map[string]interface{}) []map[string]interface{} {
	iterator := task.(map[string]interface{})["foreach"].([]interface{})
	foreachMap := make([]map[string]interface{}, len(iterator))

	for i, v := range iterator {
		// Use type assertion to convert v to map[string]string
		mapValue, ok := v.(map[string]interface{})
		if !ok {
			fmt.Println("Error parsing foreach field." + fmt.Sprint(v))
		}
		foreachMap[i] = mapValue
	}

	newTasks := []map[string]interface{}{}
	variableNames := make([]string, len(foreachMap))
	asIdentifiers := make([]string, len(foreachMap))

	for i, loop := range foreachMap {
		variableNames[i] = loop["variable"].(string)
		asIdentifiers[i] = loop["as"].(string)
	}

	variableValues := make([][]interface{}, len(variableNames))

	for i, name := range variableNames {
		value, ok := variables[name]
		if !ok {
			fmt.Println("Error parsing variable name." + fmt.Sprint(name))
		}
		variableValues[i] = value.([]interface{})
	}
	for _, combination := range product(variableValues) {
		variablesWithItems := make(map[string]interface{})
		for k, v := range variables {
			variablesWithItems[k] = v
		}
		for i, v := range combination {
			variablesWithItems[asIdentifiers[i]] = v
		}
		// Remove the 'foreach' field from the task.
		delete((task.(map[string]interface{})), "foreach")
		// Replace the placeholders in the task with the actual values
		taskToAdd := ReplacePlaceholders(task, variablesWithItems).(map[string]interface{})
		newTasks = append(newTasks, taskToAdd)
	}
	return newTasks
}

func ReplacePlaceholders(item interface{}, variables map[string]interface{}) interface{} {
	switch x := item.(type) {
	case map[string]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k] = ReplacePlaceholders(v, variables)
		}
		return m2
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = ReplacePlaceholders(v, variables)
		}
		return m2
	case []interface{}:
		i2 := make([]interface{}, len(x))
		for i, v := range x {
			i2[i] = ReplacePlaceholders(v, variables)
		}
		return i2
	case []map[interface{}]interface{}:
		i2 := make([]map[string]interface{}, len(x))
		for i, v := range x {
			i2[i] = ReplacePlaceholders(v, variables).(map[string]interface{})
		}
		return i2
	case string:
		temp := template.New("workflow")
		temp, err := temp.Parse(item.(string))
		if err != nil {
			fmt.Println("Error parsing template:", err)
			os.Exit(1)
		}
		buf := &bytes.Buffer{}
		err = temp.Execute(buf, variables)
		if err != nil {
			fmt.Println("Error executing template:", err)
			os.Exit(1)
		}
		return buf.String()
	}

	return item
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
