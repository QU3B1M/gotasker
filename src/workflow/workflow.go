// Package workflow is responsible for processing the workflow file and creating the tasks.
package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

// Task represents a task in the workflow with its dependencies and actions.
type Task struct {
	Name      string    `json:"name"`
	Do        Action    `json:"do"`
	Cleanup   Action    `json:"cleanup"`
	DependsOn []string  `json:"depends-on"`
	ForEach   []ForEach `json:"foreach"`
}

// Action represents an action to be performed with its parameters.
type Action struct {
	This string `json:"this"`
	With With   `json:"with"`
}

// With represents the parameters for an action.
type With struct {
	This string        `json:"this"`
	Args []interface{} `json:"args"`
	Path string        `json:"path"`
}

// ForEach represents a foreach loop in the workflow.
type ForEach struct {
	Variable string    `json:"variable"`
	As       string    `json:"as"`
	ForEach  []ForEach `json:"foreach"`
}

// Import represents a workflow import declaration.
type Import struct {
	File string `json:"file" yaml:"file"`
	As   string `json:"as" yaml:"as"`
}

// Workflow represents a workflow with tasks and variables.
type Workflow struct {
	Tasks     []Task      `json:"tasks"`
	Variables interface{} `json:"variables"`
	Imports   []Import    `json:"imports,omitempty" yaml:"imports,omitempty"`
}

// NewWorkflow loads a workflow from a file, processes it,
// and returns a pointer to a Workflow struct. It can return an error
// if there's a problem with marshalling or unmarshalling the data.
func NewWorkflow(workflowFilePath string) (*Workflow, error) {
	var wf Workflow

	workflowData, err := loadWorkflowFile(workflowFilePath)
	if err != nil {
		return nil, fmt.Errorf("error loading workflow: %w", err)
	}

	// Process imports if present
	if imports, ok := workflowData["imports"]; ok {
		err := processImports(workflowFilePath, imports, workflowData)
		if err != nil {
			return nil, fmt.Errorf("error processing imports: %w", err)
		}
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
	err = json.Unmarshal(jsonData, &wf)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	return &wf, nil
}

// processImports loads imported workflows and merges their tasks into the main workflow.
// Imported task names are prefixed with the namespace ("as" field) to avoid collisions.
func processImports(mainFilePath string, importsRaw interface{}, workflowData map[string]interface{}) error {
	importsList, ok := importsRaw.([]interface{})
	if !ok {
		return fmt.Errorf("imports must be a list")
	}

	mainDir := filepath.Dir(mainFilePath)
	tasks, ok := workflowData["tasks"].([]interface{})
	if !ok {
		tasks = []interface{}{}
	}

	for _, imp := range importsList {
		impMap, ok := imp.(map[string]interface{})
		if !ok {
			return fmt.Errorf("each import must be a map with 'file' and 'as' keys")
		}

		fileVal, ok := impMap["file"].(string)
		if !ok || fileVal == "" {
			return fmt.Errorf("import missing 'file' field")
		}

		namespace, ok := impMap["as"].(string)
		if !ok || namespace == "" {
			return fmt.Errorf("import missing 'as' field")
		}

		// Resolve import path relative to the main workflow file
		importPath := fileVal
		if !filepath.IsAbs(importPath) {
			importPath = filepath.Join(mainDir, importPath)
		}

		importedData, err := loadWorkflowFile(importPath)
		if err != nil {
			return fmt.Errorf("error loading import %q: %w", fileVal, err)
		}

		// Merge variables from imported workflow
		if importVars, ok := importedData["variables"].(map[string]interface{}); ok {
			if mainVars, ok := workflowData["variables"].(map[string]interface{}); ok {
				for k, v := range importVars {
					// Only add if not already defined in main workflow
					if _, exists := mainVars[k]; !exists {
						mainVars[k] = v
					}
				}
			}
		}

		// Merge tasks with namespace prefix
		if importedTasks, ok := importedData["tasks"].([]interface{}); ok {
			for _, task := range importedTasks {
				taskMap, ok := task.(map[string]interface{})
				if !ok {
					continue
				}

				// Prefix the task name with namespace
				if name, ok := taskMap["name"].(string); ok {
					taskMap["name"] = namespace + "." + name
				}

				// Prefix depends-on references with namespace
				if deps, ok := taskMap["depends-on"]; ok {
					taskMap["depends-on"] = prefixDependencies(deps, namespace)
				}

				tasks = append(tasks, taskMap)
			}
		}
	}

	workflowData["tasks"] = tasks
	return nil
}

// prefixDependencies adds a namespace prefix to dependency references
// unless they already contain a dot (already namespaced).
func prefixDependencies(deps interface{}, namespace string) interface{} {
	switch d := deps.(type) {
	case []interface{}:
		result := make([]interface{}, len(d))
		for i, dep := range d {
			if s, ok := dep.(string); ok {
				if !strings.Contains(s, ".") {
					result[i] = namespace + "." + s
				} else {
					result[i] = s
				}
			} else {
				result[i] = dep
			}
		}
		return result
	case []string:
		result := make([]string, len(d))
		for i, dep := range d {
			if !strings.Contains(dep, ".") {
				result[i] = namespace + "." + dep
			} else {
				result[i] = dep
			}
		}
		return result
	}
	return deps
}

// loadWorkflowFile reads a workflow from a file (YAML or JSON),
// converts all keys to strings, and returns the result as a map.
func loadWorkflowFile(filePath string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		err = json.Unmarshal(file, &data)
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON: %w", err)
		}
	case ".yaml", ".yml":
		err = yaml.Unmarshal(file, &data)
		if err != nil {
			return nil, fmt.Errorf("error parsing YAML: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file format: %s (use .yaml, .yml, or .json)", ext)
	}

	converted, ok := ConvertKeysToString(data).(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("error converting keys to string")
	}

	return converted, nil
}

// ProcessWorkflow processes the raw workflow data and returns a collection of tasks.
// It can return an error if there's a problem with parsing the variables or tasks.
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

// ExpandTask expands a task with foreach loops into multiple tasks based on the variables.
func ExpandTask(task interface{}, variables map[string]interface{}) []map[string]interface{} {
	taskMap, ok := task.(map[string]interface{})
	if !ok {
		return nil
	}
	iterator, ok := taskMap["foreach"].([]interface{})
	if !ok {
		return nil
	}
	foreachMap := make([]map[string]interface{}, len(iterator))

	for i, v := range iterator {
		mapValue, ok := v.(map[string]interface{})
		if !ok {
			fmt.Printf("Error parsing foreach field: %v\n", v)
			return nil
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
			fmt.Printf("Error: variable %q not found\n", name)
			return nil
		}
		vals, ok := value.([]interface{})
		if !ok {
			fmt.Printf("Error: variable %q is not a list\n", name)
			return nil
		}
		variableValues[i] = vals
	}

	// Clone the task map without the 'foreach' key to avoid mutating the original.
	cleanTask := make(map[string]interface{})
	for k, v := range taskMap {
		if k != "foreach" {
			cleanTask[k] = v
		}
	}

	for _, combination := range product(variableValues) {
		variablesWithItems := make(map[string]interface{})
		for k, v := range variables {
			variablesWithItems[k] = v
		}
		for i, v := range combination {
			variablesWithItems[asIdentifiers[i]] = v
		}
		// Replace the placeholders in the task with the actual values
		taskToAdd := ReplacePlaceholders(cleanTask, variablesWithItems).(map[string]interface{})
		newTasks = append(newTasks, taskToAdd)
	}
	return newTasks
}

// ReplacePlaceholders replaces placeholders in the item with actual values from the variables.
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
		temp, err := temp.Parse(x)
		if err != nil {
			fmt.Printf("Error parsing template %q: %v\n", x, err)
			return x
		}
		buf := &bytes.Buffer{}
		err = temp.Execute(buf, variables)
		if err != nil {
			fmt.Printf("Error executing template %q: %v\n", x, err)
			return x
		}
		return buf.String()
	}

	return item
}

// product generates the Cartesian product of a slice of slices.
// It returns a slice of slices, where each inner slice is a combination
// of elements from the input slices.
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
