package tests

import (
	"gotasker/src/workflow"
	"reflect"
	"testing"
)

// --- Tests for ReplacePlaceholders ---

func TestReplacePlaceholders(t *testing.T) {
	variables := map[string]interface{}{
		"var1": "value1",
		"var2": "value2",
		"var3": "value3",
	}
	input := map[string]interface{}{
		"key1": "{{ .var1 }}",
		"key2": "{{ .var2 }}",
		"key3": "{{ .var3 }}",
	}
	expected := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	actual := workflow.ReplacePlaceholders(input, variables)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

func TestReplacePlaceholdersWithMap(t *testing.T) {
	variables := map[string]interface{}{
		"var1": "value1",
		"var2": "value2",
		"var3": "value3",
	}
	input := map[interface{}]interface{}{
		"key1": "{{ .var1 }}",
		"key2": "{{ .var2 }}",
		"key3": "{{ .var3 }}",
	}
	expected := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	actual := workflow.ReplacePlaceholders(input, variables)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

func TestReplacePlaceholdersWithSlice(t *testing.T) {
	variables := map[string]interface{}{
		"var1": "value1",
		"var2": "value2",
		"var3": "value3",
	}
	input := []interface{}{
		"{{ .var1 }}",
		"{{ .var2 }}",
		"{{ .var3 }}",
	}
	expected := []interface{}{
		"value1",
		"value2",
		"value3",
	}
	actual := workflow.ReplacePlaceholders(input, variables)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

func TestReplacePlaceholdersWithSliceOfMap(t *testing.T) {
	variables := map[string]interface{}{
		"var1": "value1",
		"var2": "value2",
		"var3": "value3",
	}
	input := []map[interface{}]interface{}{
		{
			"key1": "{{ .var1 }}",
			"key2": "{{ .var2 }}",
			"key3": "{{ .var3 }}",
		},
	}
	expected := []map[string]interface{}{
		{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
	}
	actual := workflow.ReplacePlaceholders(input, variables)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

// --- Tests for ExpandTask ---

func TestExpandTask(t *testing.T) {
	variables := map[string]interface{}{
		"var1": []interface{}{"var1-value1", "var1-value2"},
		"var2": []interface{}{"var2-value1", "var2-value2"},
	}
	input := map[string]interface{}{
		"task": "task1",
		"key1": "{{ .v1 }}",
		"key2": "{{ .v2 }}",
		"foreach": []interface{}{
			map[string]interface{}{
				"variable": "var1",
				"as":       "v1",
			},
			map[string]interface{}{
				"variable": "var2",
				"as":       "v2",
			},
		},
	}
	expected := []map[string]interface{}{
		{
			"task": "task1",
			"key1": "var1-value1",
			"key2": "var2-value1",
		},
		{
			"task": "task1",
			"key1": "var1-value1",
			"key2": "var2-value2",
		},
		{
			"task": "task1",
			"key1": "var1-value2",
			"key2": "var2-value1",
		},
		{
			"task": "task1",
			"key1": "var1-value2",
			"key2": "var2-value2",
		},
	}
	actual := workflow.ExpandTask(input, variables)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

func TestExpandTaskWithNestedForeach(t *testing.T) {
	variables := map[string]interface{}{
		"var1": []interface{}{"var1-value1", "var1-value2"},
		"var2": []interface{}{"var2-value1", "var2-value2"},
		"var3": []interface{}{"var3-value1", "var3-value2"},
	}
	input := map[string]interface{}{
		"task": "task1",
		"key1": "{{ .v1 }}",
		"key2": "{{ .v2 }}",
		"foreach": []interface{}{
			map[string]interface{}{
				"variable": "var1",
				"as":       "v1",
			},
			map[string]interface{}{
				"variable": "var2",
				"as":       "v2",
				"foreach": []interface{}{
					map[string]interface{}{
						"variable": "var3",
						"as":       "v3",
					},
				},
			},
		},
	}
	expected := []map[string]interface{}{
		{
			"task": "task1",
			"key1": "var1-value1",
			"key2": "var2-value1",
		},
		{
			"task": "task1",
			"key1": "var1-value1",
			"key2": "var2-value2",
		},
		{
			"task": "task1",
			"key1": "var1-value2",
			"key2": "var2-value1",
		},
		{
			"task": "task1",
			"key1": "var1-value2",
			"key2": "var2-value2",
		},
	}
	actual := workflow.ExpandTask(input, variables)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

func TestExpandTaskWithNestedForeachAndVariables(t *testing.T) {
	variables := map[string]interface{}{
		"var1": []interface{}{"var1-value1", "var1-value2"},
		"var2": []interface{}{"var2-value1", "var2-value2"},
		"var3": []interface{}{"var3-value1", "var3-value2"},
	}
	input := map[string]interface{}{
		"task": "task1",
		"key1": "{{ .v1 }}",
		"key2": "{{ .v2 }}",
		"foreach": []interface{}{
			map[string]interface{}{
				"variable": "var1",
				"as":       "v1",
			},
			map[string]interface{}{
				"variable": "var2",
				"as":       "v2",
				"foreach": []interface{}{
					map[string]interface{}{
						"variable": "var3",
						"as":       "v3",
					},
				},
			},
		},
	}
	expected := []map[string]interface{}{
		{
			"task": "task1",
			"key1": "var1-value1",
			"key2": "var2-value1",
		},
		{
			"task": "task1",
			"key1": "var1-value1",
			"key2": "var2-value2",
		},
		{
			"task": "task1",
			"key1": "var1-value2",
			"key2": "var2-value1",
		},
		{
			"task": "task1",
			"key1": "var1-value2",
			"key2": "var2-value2",
		},
	}
	actual := workflow.ExpandTask(input, variables)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

// --- Tests for ConvertKeysToString ---

func TestConvertKeysToString(t *testing.T) {
	input := map[interface{}]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": []interface{}{
			"item1",
			"item2",
		},
		"key4": map[interface{}]interface{}{
			"key5": "value5",
			"key6": "value6",
		},
	}
	expected := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": []interface{}{
			"item1",
			"item2",
		},
		"key4": map[string]interface{}{
			"key5": "value5",
			"key6": "value6",
		},
	}
	actual := workflow.ConvertKeysToString(input)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

func TestConvertKeysToStringWithMap(t *testing.T) {
	input := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": []interface{}{
			"item1",
			"item2",
		},
		"key4": map[string]interface{}{
			"key5": "value5",
			"key6": "value6",
		},
	}
	expected := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": []interface{}{
			"item1",
			"item2",
		},
		"key4": map[string]interface{}{
			"key5": "value5",
			"key6": "value6",
		},
	}
	actual := workflow.ConvertKeysToString(input)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

func TestConvertKeysToStringWithSlice(t *testing.T) {
	input := []interface{}{
		"item1",
		"item2",
		map[interface{}]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}
	expected := []interface{}{
		"item1",
		"item2",
		map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}
	actual := workflow.ConvertKeysToString(input)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

func TestConvertKeysToStringWithSliceOfMap(t *testing.T) {
	input := []map[interface{}]interface{}{
		{
			"key1": "value1",
			"key2": "value2",
		},
		{
			"key3": "value3",
			"key4": "value4",
		},
	}
	expected := []map[string]interface{}{
		{
			"key1": "value1",
			"key2": "value2",
		},
		{
			"key3": "value3",
			"key4": "value4",
		},
	}
	actual := workflow.ConvertKeysToString(input)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

// --- Tests for ProcessWorkflow ---

func TestProcessWorkflow(t *testing.T) {
	input := map[string]interface{}{
		"variables": map[string]interface{}{
			"var1": "value1",
			"var2": "value2",
		},
		"tasks": []interface{}{
			map[string]interface{}{
				"task": "task1",
				"key1": "{{ .var1 }}",
				"key2": "{{ .var2 }}",
			},
		},
	}
	expected := []map[string]interface{}{
		{
			"task": "task1",
			"key1": "value1",
			"key2": "value2",
		},
	}
	actual, err := workflow.ProcessWorkflow(input)
	if !reflect.DeepEqual(nil, err) {
		t.Errorf("Error at processing input. %v", err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

func TestProcessWorkflowWithForeach(t *testing.T) {
	input := map[string]interface{}{
		"variables": map[string]interface{}{
			"var1": []interface{}{"var1-value1", "var1-value2"},
			"var2": []interface{}{"var2-value1", "var2-value2"},
		},
		"tasks": []interface{}{
			map[string]interface{}{
				"task": "task1",
				"key1": "{{ .v1 }}",
				"key2": "{{ .v2 }}",
				"foreach": []interface{}{
					map[string]interface{}{
						"variable": "var1",
						"as":       "v1",
					},
					map[string]interface{}{
						"variable": "var2",
						"as":       "v2",
					},
				},
			},
		},
	}
	expected := []map[string]interface{}{
		{
			"task": "task1",
			"key1": "var1-value1",
			"key2": "var2-value1",
		},
		{
			"task": "task1",
			"key1": "var1-value1",
			"key2": "var2-value2",
		},
		{
			"task": "task1",
			"key1": "var1-value2",
			"key2": "var2-value1",
		},
		{
			"task": "task1",
			"key1": "var1-value2",
			"key2": "var2-value2",
		},
	}
	actual, err := workflow.ProcessWorkflow(input)
	if !reflect.DeepEqual(nil, err) {
		t.Errorf("Error at processing input. %v", err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

func TestProcessWorkflowWithNestedForeach(t *testing.T) {
	input := map[string]interface{}{
		"variables": map[string]interface{}{
			"var1": []interface{}{"var1-value1", "var1-value2"},
			"var2": []interface{}{"var2-value1", "var2-value2"},
			"var3": []interface{}{"var3-value1", "var3-value2"},
		},
		"tasks": []interface{}{
			map[string]interface{}{
				"task": "task1",
				"key1": "{{ .v1 }}",
				"key2": "{{ .v2 }}",
				"foreach": []interface{}{
					map[string]interface{}{
						"variable": "var1",
						"as":       "v1",
					},
					map[string]interface{}{
						"variable": "var2",
						"as":       "v2",
					},
				},
			},
		},
	}
	expected := []map[string]interface{}{
		{
			"task": "task1",
			"key1": "var1-value1",
			"key2": "var2-value1",
		},
		{
			"task": "task1",
			"key1": "var1-value1",
			"key2": "var2-value2",
		},
		{
			"task": "task1",
			"key1": "var1-value2",
			"key2": "var2-value1",
		},
		{
			"task": "task1",
			"key1": "var1-value2",
			"key2": "var2-value2",
		},
	}
	actual, err := workflow.ProcessWorkflow(input)
	if !reflect.DeepEqual(nil, err) {
		t.Errorf("Error at processing input. %v", err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

func TestProcessWorkflowWithNestedForeachAndVariables(t *testing.T) {
	input := map[string]interface{}{
		"variables": map[string]interface{}{
			"var1": []interface{}{"var1-value1", "var1-value2"},
			"var2": []interface{}{"var2-value1", "var2-value2"},
			"var3": []interface{}{"var3-value1", "var3-value2"},
		},
		"tasks": []interface{}{
			map[string]interface{}{
				"task": "task1",
				"key1": "{{ .v1 }}",
				"key2": "{{ .v2 }}",
				"foreach": []interface{}{
					map[string]interface{}{
						"variable": "var1",
						"as":       "v1",
					},
					map[string]interface{}{
						"variable": "var2",
						"as":       "v2",
						"foreach": []interface{}{
							map[string]interface{}{
								"variable": "var3",
								"as":       "v3",
							},
						},
					},
				},
			},
		},
	}
	expected := []map[string]interface{}{
		{
			"task": "task1",
			"key1": "var1-value1",
			"key2": "var2-value1",
		},
		{
			"task": "task1",
			"key1": "var1-value1",
			"key2": "var2-value2",
		},
		{
			"task": "task1",
			"key1": "var1-value2",
			"key2": "var2-value1",
		},
		{
			"task": "task1",
			"key1": "var1-value2",
			"key2": "var2-value2",
		},
	}
	actual, err := workflow.ProcessWorkflow(input)
	if !reflect.DeepEqual(nil, err) {
		t.Errorf("Error at processing input. %v", err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

func TestProcessWorkflowWithNestedForeachAndVariablesAndMap(t *testing.T) {
	input := map[string]interface{}{
		"variables": map[string]interface{}{
			"var1": []interface{}{"var1-value1", "var1-value2"},
			"var2": []interface{}{"var2-value1", "var2-value2"},
			"var3": []interface{}{"var3-value1", "var3-value2"},
		},
		"tasks": []interface{}{
			map[string]interface{}{
				"task": "task1",
				"key1": "{{ .v1 }}",
				"key2": "{{ .v2 }}",
				"foreach": []interface{}{
					map[string]interface{}{
						"variable": "var1",
						"as":       "v1",
					},
					map[string]interface{}{
						"variable": "var2",
						"as":       "v2",
						"foreach": []interface{}{
							map[string]interface{}{
								"variable": "var3",
								"as":       "v3",
							},
						},
					},
				},
			},
		},
	}
	expected := []map[string]interface{}{
		{
			"task": "task1",
			"key1": "var1-value1",
			"key2": "var2-value1",
		},
		{
			"task": "task1",
			"key1": "var1-value1",
			"key2": "var2-value2",
		},
		{
			"task": "task1",
			"key1": "var1-value2",
			"key2": "var2-value1",
		},
		{
			"task": "task1",
			"key1": "var1-value2",
			"key2": "var2-value2",
		},
	}
	actual, err := workflow.ProcessWorkflow(input)
	if !reflect.DeepEqual(nil, err) {
		t.Errorf("Error at processing input. %v", err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}
