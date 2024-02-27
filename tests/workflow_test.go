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
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
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
