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

func TestCancelDependentTasksAbortAll(t *testing.T) {
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
	d := dag.NewDAG(taskCollection, false)
	d.CancelDependentTasks("a", "abort-all")
	tasksToCancel := d.GetTasksToCancel()
	// abort-all should cancel all tasks in all execution plan entries
	if len(tasksToCancel) == 0 {
		t.Error("CancelDependentTasks abort-all did not cancel any tasks")
	}
}

func TestCancelDependentTasksAbortRelated(t *testing.T) {
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
	d := dag.NewDAG(taskCollection, false)
	d.SetStatus("b", "successful")
	d.CancelDependentTasks("a", "abort-related-flows")
	tasksToCancel := d.GetTasksToCancel()
	// Should have canceled tasks related to "a" that aren't already finished
	if len(tasksToCancel) == 0 {
		t.Error("CancelDependentTasks abort-related-flows did not cancel any tasks")
	}
}

func TestCancelDependentTasksContinue(t *testing.T) {
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
	d.CancelDependentTasks("a", "continue")
	tasksToCancel := d.GetTasksToCancel()
	if len(tasksToCancel) != 0 {
		t.Error("CancelDependentTasks with continue policy should not cancel tasks")
	}
}

func TestGetDependencyTree(t *testing.T) {
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
	expected := map[string][]string{
		"a": {"b"},
		"b": {},
		"c": {"a"},
	}
	d := dag.NewDAG(taskCollection, false)
	dependencyTree := d.GetDependencyTree()
	if !reflect.DeepEqual(dependencyTree, expected) {
		t.Errorf("GetDependencyTree returned: %v, expected: %v", dependencyTree, expected)
	}
}

func TestGetTopSortedLayers(t *testing.T) {
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
	d := dag.NewDAG(taskCollection, false)
	layers := d.GetTopSortedLayers()
	if len(layers) != 3 {
		t.Errorf("Expected 3 layers, got %d: %v", len(layers), layers)
	}
	if layers[0][0] != "b" {
		t.Errorf("First layer should be [b], got %v", layers[0])
	}
}

func TestDAGWithInterfaceDependencies(t *testing.T) {
	// Simulate what happens when deps come from JSON unmarshal ([]interface{} not []string)
	taskCollection := []map[string]interface{}{
		{
			"task":       "a",
			"depends-on": []interface{}{"b"},
		},
		{
			"task":       "b",
			"depends-on": []interface{}{},
		},
	}
	d := dag.NewDAG(taskCollection, false)
	if d == nil {
		t.Error("NewDAG returned nil for []interface{} deps")
	}
	available := d.GetAvailableTasks()
	if len(available) != 2 {
		t.Errorf("Expected 2 available tasks, got %d", len(available))
	}
}

// GetExecutionPlan

func TestGetExecutionPlanLayered(t *testing.T) {
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

func TestGetExecutionPlanSingle(t *testing.T) {
	taskCollection := []map[string]interface{}{
		{
			"task":       "a",
			"depends-on": []string{},
		},
	}
	expected := map[string]interface{}{
		"a": map[string]interface{}{},
	}
	d := dag.NewDAG(taskCollection, false)
	executionPlan := d.GetExecutionPlan()
	if !reflect.DeepEqual(executionPlan, expected) {
		t.Errorf("GetExecutionPlan returned: %v, expected: %v", executionPlan, expected)
	}
}

func TestGetExecutionPlanMultiple(t *testing.T) {
	taskCollection := []map[string]interface{}{
		{
			"task":       "a",
			"depends-on": []string{},
		},
		{
			"task":       "b",
			"depends-on": []string{},
		},
	}
	expected := map[string]interface{}{
		"a": map[string]interface{}{},
		"b": map[string]interface{}{},
	}
	d := dag.NewDAG(taskCollection, false)
	executionPlan := d.GetExecutionPlan()
	if !reflect.DeepEqual(executionPlan, expected) {
		t.Errorf("GetExecutionPlan returned: %v, expected: %v", executionPlan, expected)
	}
}

func TestGetExecutionPlanComplex(t *testing.T) {
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
		{
			"task":       "d",
			"depends-on": []string{"a"},
		},
	}
	expected := map[string]interface{}{
		"c": map[string]interface{}{
			"a": map[string]interface{}{
				"b": map[string]interface{}{},
			},
		},
		"d": map[string]interface{}{
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
