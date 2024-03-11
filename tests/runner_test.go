package tests

import (
	"gotasker/src/runner"
	"testing"
)

func TestNewExecution(t *testing.T) {
	name := "echo"
	params := map[string]interface{}{
		"args": []interface{}{"hello", "world"},
	}
	e := runner.NewExecution(name, params)
	if e == nil {
		t.Error("NewExecution returned nil")
	}
}

func TestExecute(t *testing.T) {
	name := "echo"
	params := map[string]interface{}{
		"args": []interface{}{"hello", "world"},
	}
	e := runner.NewExecution(name, params)
	_, err := e.Execute()
	if err != nil {
		t.Errorf("Execute returned an error: %v", err)
	}
}

func TestExecuteMultipleArgs(t *testing.T) {
	name := "echo"
	params := map[string]interface{}{
		"args": []interface{}{"hello", "world", "foo", "bar"},
	}
	e := runner.NewExecution(name, params)
	_, err := e.Execute()
	if err != nil {
		t.Errorf("Execute returned an error: %v", err)
	}
}

func TestExecuteComplex(t *testing.T) {
	name := "echo"
	params := map[string]interface{}{
		"args": []interface{}{
			"hello",
			"world",
			map[string]interface{}{
				"arg1": "value1",
				"arg2": []interface{}{"value2", "value3"},
			},
		},
	}
	e := runner.NewExecution(name, params)
	_, err := e.Execute()
	if err != nil {
		t.Errorf("Execute returned an error: %v", err)
	}
}

func TestExecuteError(t *testing.T) {
	name := "nonexistent"
	params := map[string]interface{}{
		"args": []interface{}{},
	}
	e := runner.NewExecution(name, params)
	_, err := e.Execute()
	if err == nil {
		t.Error("Execute did not return an error")
	}
}

func TestExecuteErrorWithArgs(t *testing.T) {
	name := "nonexistent"
	params := map[string]interface{}{
		"args": []interface{}{"hello", "world"},
	}
	e := runner.NewExecution(name, params)
	_, err := e.Execute()
	if err == nil {
		t.Error("Execute did not return an error")
	}
}

func TestExecuteErrorWithComplexArgs(t *testing.T) {
	name := "nonexistent"
	params := map[string]interface{}{
		"args": []interface{}{
			"hello",
			"world",
			map[string]interface{}{
				"arg1": "value1",
				"arg2": []interface{}{"value2", "value3"},
			},
		},
	}
	e := runner.NewExecution(name, params)
	_, err := e.Execute()
	if err == nil {
		t.Error("Execute did not return an error")
	}
}
