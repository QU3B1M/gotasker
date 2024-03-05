package tests

import (
	"gotasker/src/runner"
	"testing"
)

func TestNewExecution(t *testing.T) {
	taskParameters := map[string]interface{}{
		"path": "echo",
		"args": []interface{}{"hello", "world"},
	}
	e := runner.NewExecution("echo", taskParameters)
	if e == nil {
		t.Error("NewExecution returned nil")
	}
}

func TestExecute(t *testing.T) {
	taskParameters := map[string]interface{}{
		"path": "echo",
		"args": []interface{}{"hello", "world"},
	}
	e := runner.NewExecution("echo", taskParameters)
	err := e.Execute()
	if err != nil {
		t.Errorf("Execute returned an error: %v", err)
	}
}

func TestExecuteMultipleArgs(t *testing.T) {
	taskParameters := map[string]interface{}{
		"path": "echo",
		"args": []interface{}{"hello", "world", "foo", "bar"},
	}
	e := runner.NewExecution("echo", taskParameters)
	err := e.Execute()
	if err != nil {
		t.Errorf("Execute returned an error: %v", err)
	}
}

func TestExecuteComplex(t *testing.T) {
	taskParameters := map[string]interface{}{
		"path": "echo",
		"args": []interface{}{
			"hello",
			"world",
			map[string]interface{}{
				"arg1": "value1",
				"arg2": []interface{}{"value2", "value3"},
			},
		},
	}
	e := runner.NewExecution("echo", taskParameters)
	err := e.Execute()
	if err != nil {
		t.Errorf("Execute returned an error: %v", err)
	}
}

func TestExecuteError(t *testing.T) {
	taskParameters := map[string]interface{}{
		"path": "nonexistent",
		"args": []interface{}{},
	}
	e := runner.NewExecution("nonexistent", taskParameters)
	err := e.Execute()
	if err == nil {
		t.Error("Execute did not return an error")
	}
}

func TestExecuteErrorWithArgs(t *testing.T) {
	taskParameters := map[string]interface{}{
		"path": "nonexistent",
		"args": []interface{}{"hello", "world"},
	}
	e := runner.NewExecution("nonexistent", taskParameters)
	err := e.Execute()
	if err == nil {
		t.Error("Execute did not return an error")
	}
}

func TestExecuteErrorWithComplexArgs(t *testing.T) {
	taskParameters := map[string]interface{}{
		"path": "nonexistent",
		"args": []interface{}{
			"hello",
			"world",
			map[string]interface{}{
				"arg1": "value1",
				"arg2": []interface{}{"value2", "value3"},
			},
		},
	}
	e := runner.NewExecution("nonexistent", taskParameters)
	err := e.Execute()
	if err == nil {
		t.Error("Execute did not return an error")
	}
}

func TestExecuteErrorWithInvalidArgs(t *testing.T) {
	taskParameters := map[string]interface{}{
		"path": "echo",
		"args": []interface{}{1},
	}
	e := runner.NewExecution("echo", taskParameters)
	err := e.Execute()
	if err == nil {
		t.Error("Execute did not return an error")
	}
}
