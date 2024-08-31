// Package dag is responsible for creating a directed acyclic graph (DAG)
// from a given task collection.
package dag

import "gotasker/src/graph"

// DAG represents a directed acyclic graph with tasks and their dependencies.
type DAG struct {
	taskCollection      []map[string]interface{}
	reverse             bool
	graph               *graph.DependencyGraph
	dependencyTree      map[string][]string
	toBeCanceled        map[string]struct{}
	finishedTasksStatus map[string]map[string]struct{}
	executionPlan       map[string]interface{}
}

// NewDAG creates a new DAG instance with the given task collection.
// The reverse parameter is used to determine the direction of the graph.
func NewDAG(taskCollection []map[string]interface{}, reverse bool) *DAG {
	d := &DAG{
		taskCollection: taskCollection,
		reverse:        reverse,
		toBeCanceled:   make(map[string]struct{}),
		finishedTasksStatus: map[string]map[string]struct{}{
			"failed":     {},
			"canceled":   {},
			"successful": {},
		},
	}
	d.graph, d.dependencyTree = d.buildDAG()
	d.executionPlan = d.createExecutionPlan(d.dependencyTree)
	return d
}

// GetAvailableTasks returns the list of tasks that are available for execution.
func (d *DAG) GetAvailableTasks() []string {
	return d.graph.TopSorted()
}

// GetDependencyTree returns the dependency tree of tasks.
func (d *DAG) GetDependencyTree() map[string][]string {
	return d.dependencyTree
}

// GetExecutionPlan returns the execution plan for the tasks.
func (d *DAG) GetExecutionPlan() map[string]interface{} {
	return d.executionPlan
}

// SetStatus sets the status of a given task.
func (d *DAG) SetStatus(taskName string, status string) {
	d.finishedTasksStatus[status][taskName] = struct{}{}
}

// GetStatus returns the status of a given task.
func (d *DAG) GetStatus(taskName string) string {
	if _, ok := d.finishedTasksStatus["successful"][taskName]; ok {
		return "successful"
	} else if _, ok := d.finishedTasksStatus["failed"][taskName]; ok {
		return "failed"
	} else if _, ok := d.finishedTasksStatus["canceled"][taskName]; ok {
		return "canceled"
	}
	return "pending"
}

// CancelTask marks a task for cancellation if it is still pending.
func (d *DAG) CancelTask(taskName string) bool {
	if d.GetStatus(taskName) == "pending" {
		d.toBeCanceled[taskName] = struct{}{}
		return true
	}
	return false
}

// CancelDependentTasks cancels tasks dependent on the given task based on the cancel policy.
func (d *DAG) CancelDependentTasks(taskName string, cancelPolicy string) {
	d.cancelDependantTasks(taskName, cancelPolicy)
}

// GetTasksToCancel returns the tasks that are marked to be canceled.
func (d *DAG) GetTasksToCancel() map[string]struct{} {
	return d.toBeCanceled
}

// buildDAG constructs the dependency graph and dependency tree from the task collection.
func (d *DAG) buildDAG() (*graph.DependencyGraph, map[string][]string) {
	dependencyDict := make(map[string][]string)
	graph := graph.NewGraph()
	for _, task := range d.taskCollection {
		taskName := task["task"].(string)
		dependencies, ok := task["depends-on"].([]string)
		if !ok {
			dependencies = []string{}
		}
		for _, dependency := range dependencies {
			if d.reverse {
				graph.DependOn(dependency, taskName)
			} else {
				graph.DependOn(taskName, dependency)
			}
		}
		dependencyDict[taskName] = dependencies
	}
	return graph, dependencyDict
}

// getAllTaskSet returns a set of all tasks from the given map.
func getAllTaskSet(tasks map[string]struct{}) map[string]struct{} {
	taskSet := make(map[string]struct{})
	for task := range tasks {
		taskSet[task] = struct{}{}
	}
	return taskSet
}

// cancelDependantTasks cancels dependent tasks based on the given cancel policy.
func (d *DAG) cancelDependantTasks(taskName string, cancelPolicy string) {
	if cancelPolicy == "continue" {
		return
	}
	notCancelledTasks := make(map[string]struct{})
	for k, v := range d.finishedTasksStatus["failed"] {
		notCancelledTasks[k] = v
	}
	for k, v := range d.finishedTasksStatus["successful"] {
		notCancelledTasks[k] = v
	}
	for _, tasks := range d.executionPlan {
		taskSet := getAllTaskSet(tasks.(map[string]struct{}))
		if cancelPolicy == "abort-all" {
			for k := range taskSet {
				d.toBeCanceled[k] = struct{}{}
			}
		} else if cancelPolicy == "abort-related-flows" {
			if _, ok := taskSet[taskName]; ok {
				for k := range taskSet {
					if _, ok := notCancelledTasks[k]; !ok {
						d.toBeCanceled[k] = struct{}{}
					}
				}
			}
		}
	}
}

// getSubtaskPlan generates a subtask plan for a given task based on its dependencies.
func getSubtaskPlan(taskName string, dependencyDict map[string][]string, level int) map[string]interface{} {
	dependencies, ok := dependencyDict[taskName]
	if !ok {
		dependencies = []string{}
	}
	plan := map[string]interface{}{taskName: make(map[string]interface{})}
	for _, dependency := range dependencies {
		subPlan := getSubtaskPlan(dependency, dependencyDict, level+1)
		plan[taskName].(map[string]interface{})[dependency] = subPlan[dependency]
	}
	return plan
}

// createExecutionPlan constructs an execution plan for the tasks based on their dependencies.
func (d *DAG) createExecutionPlan(dependencyDict map[string][]string) map[string]interface{} {
	executionPlan := make(map[string]interface{})
	// getRootTasks identifies the root tasks that have no dependencies.
	getRootTasks := func(dependencyDict map[string][]string) []string {
		allTasks := make(map[string]struct{})
		for k := range dependencyDict {
			allTasks[k] = struct{}{}
		}
		dependentTasks := make(map[string]struct{})
		for _, dependents := range dependencyDict {
			for _, dep := range dependents {
				dependentTasks[dep] = struct{}{}
			}
		}
		rootTasks := make([]string, 0)
		for k := range allTasks {
			if _, ok := dependentTasks[k]; !ok {
				rootTasks = append(rootTasks, k)
			}
		}
		return rootTasks
	}
	rootTasks := getRootTasks(dependencyDict)
	for _, rootTask := range rootTasks {
		executionPlan[rootTask] = getSubtaskPlan(rootTask, dependencyDict, 0)[rootTask]
	}
	return executionPlan
}
