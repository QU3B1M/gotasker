package tests

import (
	"flowgo/src/workflow"
	"reflect"
	"testing"
)

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
