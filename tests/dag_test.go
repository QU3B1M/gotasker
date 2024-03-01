// Unit tests to ensure the correct function of the "gotasker/src/dag" package.
package tests

import (
	"gotasker/src/dag"
	"reflect"
	"testing"
)

func TestNewDAG(t *testing.T) {
	taskCollection := []map[string]interface{}{
		{
			"task":       "a",
			"depends-on": []string{"b"},
		},
		{
			"task":       "b",
			"depends-on": []string{},
		},
	}
	d := dag.NewDAG(taskCollection, false)
	if d == nil {
		t.Error("NewDAG returned nil")
	}
}

func TestGetAvailableTasks(t *testing.T) {
	taskCollection := []map[string]interface{}{
		{
			"task":       "a",
			"depends-on": []string{"b"},
		},
		{
			"task":       "b",
			"depends-on": []string{},
		},
	}
	d := dag.NewDAG(taskCollection, false)
	availableTasks := d.GetAvailableTasks()
	if len(availableTasks) != 2 || availableTasks[0] != "b" {
		t.Errorf("GetAvailableTasks returned: %v, expected: [b a]", availableTasks)
	}
}

func TestGetExecutionPlan(t *testing.T) {
	taskCollection := []map[string]interface{}{
		{
			"task":       "a",
			"depends-on": []string{"b"},
		},
		{
			"task":       "b",
			"depends-on": []string{},
		},
		{
			"task":       "c",
			"depends-on": []string{"a"},
		},
	}
	expected := map[string]interface{}{
		"c": map[string]interface{}{
			"a": map[string]interface{}{
				"b": map[string]interface{}{},
			},
		},
	}
	d := dag.NewDAG(taskCollection, false)
	executionPlan := d.GetExecutionPlan()
	if !reflect.DeepEqual(executionPlan, expected) {
		t.Errorf("GetExecutionPlan returned: %v, expected: %v", executionPlan, expected)
	}
}

func TestSetStatusSuccessful(t *testing.T) {
	taskCollection := []map[string]interface{}{
		{
			"task":       "a",
			"depends-on": []string{"b"},
		},
		{
			"task":       "b",
			"depends-on": []string{},
		},
	}
	d := dag.NewDAG(taskCollection, false)
	d.SetStatus("a", "successful")
	if d.GetStatus("a") != "successful" {
		t.Error("SetStatus did not set the status")
	}
}

func TestSetStatusFailed(t *testing.T) {
	taskCollection := []map[string]interface{}{
		{
			"task":       "a",
			"depends-on": []string{"b"},
		},
		{
			"task":       "b",
			"depends-on": []string{},
		},
	}
	d := dag.NewDAG(taskCollection, false)
	d.SetStatus("a", "failed")
	if d.GetStatus("a") != "failed" {
		t.Error("SetStatus did not set the status")
	}
}

func TestSetStatusCanceled(t *testing.T) {
	taskCollection := []map[string]interface{}{
		{
			"task":       "a",
			"depends-on": []string{"b"},
		},
		{
			"task":       "b",
			"depends-on": []string{},
		},
	}
	d := dag.NewDAG(taskCollection, false)
	d.SetStatus("a", "canceled")
	if d.GetStatus("a") != "canceled" {
		t.Error("SetStatus did not set the status")
	}
}

func TestCancelTask(t *testing.T) {
	taskCollection := []map[string]interface{}{
		{
			"task":       "a",
			"depends-on": []string{"b"},
		},
		{
			"task":       "b",
			"depends-on": []string{},
		},
	}
	d := dag.NewDAG(taskCollection, false)
	d.CancelTask("a")
	tasksToCancel := d.GetTasksToCancel()
	if _, ok := tasksToCancel["a"]; !ok {
		t.Error("CancelTask did not set the task to cancel")
	}
}

// func TestCancelDependentTasks(t *testing.T) {
// 	taskCollection := []map[string]interface{}{
// 		{
// 			"task":       "a",
// 			"depends-on": []string{"b"},
// 		},
// 		{
// 			"task":       "b",
// 			"depends-on": []string{},
// 		},
// 		{
// 			"task":       "c",
// 			"depends-on": []string{"a"},
// 		},
// 	}
// 	d := dag.NewDAG(taskCollection, false)
// 	d.CancelDependentTasks("a", "all")
// 	tasksToCancel := d.GetTasksToCancel()
// 	if _, ok := tasksToCancel["a"]; !ok {
// 		t.Error("CancelDependentTasks did not set the task to cancel")
// 	}
// 	if _, ok := tasksToCancel["c"]; !ok {
// 		t.Error("CancelDependentTasks did not set the dependent task to cancel")
// 	}
// }
