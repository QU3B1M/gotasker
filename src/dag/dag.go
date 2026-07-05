// Package dag is responsible for creating a directed acyclic graph (DAG)
// from a given task collection.
package dag

import (
	"fmt"
	"gotasker/src/graph"
	"sync"
)

// DAG represents a directed acyclic graph with tasks and their dependencies.
type DAG struct {
	taskCollection      []map[string]interface{}
	reverse             bool
	graph               *graph.DependencyGraph
	dependencyTree      map[string][]string
	toBeCanceled        map[string]struct{}
	finishedTasksStatus map[string]map[string]struct{}
	executionPlan       map[string]interface{}
	mu                  sync.RWMutex
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

// GetTopSortedLayers returns tasks grouped in layers for parallel execution.
// Each layer contains tasks that can be run concurrently.
func (d *DAG) GetTopSortedLayers() [][]string {
	return d.graph.TopSortedLayers()
}

// SetStatus sets the status of a given task.
func (d *DAG) SetStatus(taskName string, status string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.finishedTasksStatus[status][taskName] = struct{}{}
}

// GetStatus returns the status of a given task.
func (d *DAG) GetStatus(taskName string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()
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
	d.mu.Lock()
	defer d.mu.Unlock()
	// Check status without lock (we already hold it)
	pending := true
	if _, ok := d.finishedTasksStatus["successful"][taskName]; ok {
		pending = false
	} else if _, ok := d.finishedTasksStatus["failed"][taskName]; ok {
		pending = false
	} else if _, ok := d.finishedTasksStatus["canceled"][taskName]; ok {
		pending = false
	}
	if pending {
		d.toBeCanceled[taskName] = struct{}{}
		return true
	}
	return false
}

// CancelDependentTasks cancels tasks dependent on the given task based on the cancel policy.
func (d *DAG) CancelDependentTasks(taskName string, cancelPolicy string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cancelDependantTasksLocked(taskName, cancelPolicy)
}

// GetTasksToCancel returns the tasks that are marked to be canceled.
func (d *DAG) GetTasksToCancel() map[string]struct{} {
	d.mu.RLock()
	defer d.mu.RUnlock()
	// Return a copy to avoid races
	result := make(map[string]struct{})
	for k, v := range d.toBeCanceled {
		result[k] = v
	}
	return result
}

// buildDAG constructs the dependency graph and dependency tree from the task collection.
func (d *DAG) buildDAG() (*graph.DependencyGraph, map[string][]string) {
	dependencyDict := make(map[string][]string)
	g := graph.NewGraph()
	for _, task := range d.taskCollection {
		taskName, ok := task["task"].(string)
		if !ok {
			if name, ok := task["name"].(string); ok {
				taskName = name
			} else {
				continue
			}
		}

		// Handle depends-on: could be []string or []interface{} depending on source
		var dependencies []string
		switch deps := task["depends-on"].(type) {
		case []string:
			dependencies = deps
		case []interface{}:
			for _, dep := range deps {
				if s, ok := dep.(string); ok {
					dependencies = append(dependencies, s)
				}
			}
		default:
			dependencies = []string{}
		}

		for _, dependency := range dependencies {
			var err error
			if d.reverse {
				err = g.DependOn(dependency, taskName)
			} else {
				err = g.DependOn(taskName, dependency)
			}
			if err != nil {
				fmt.Printf("Warning: could not add dependency %s -> %s: %v\n", taskName, dependency, err)
			}
		}
		// Ensure standalone tasks (no dependencies) are still in the graph
		g.AddNode(taskName)
		dependencyDict[taskName] = dependencies
	}
	return g, dependencyDict
}

// getAllTaskSet returns a set of all tasks from the given map.
func getAllTaskSet(tasks map[string]struct{}) map[string]struct{} {
	taskSet := make(map[string]struct{})
	for task := range tasks {
		taskSet[task] = struct{}{}
	}
	return taskSet
}

// cancelDependantTasksLocked cancels dependent tasks based on the given cancel policy.
// Must be called with d.mu held.
func (d *DAG) cancelDependantTasksLocked(taskName string, cancelPolicy string) {
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

	if cancelPolicy == "abort-all" {
		// Cancel every task in the execution plan
		for rootTask, subtree := range d.executionPlan {
			d.toBeCanceled[rootTask] = struct{}{}
			collectTaskNames(subtree, d.toBeCanceled)
		}
	} else if cancelPolicy == "abort-related-flows" {
		// Cancel the root task and its subtree if the failed task is part of it
		for rootTask, subtree := range d.executionPlan {
			allTasks := map[string]struct{}{rootTask: {}}
			collectTaskNames(subtree, allTasks)
			if _, ok := allTasks[taskName]; ok {
				for k := range allTasks {
					if _, done := notCancelledTasks[k]; !done {
						d.toBeCanceled[k] = struct{}{}
					}
				}
			}
		}
	}
}

// collectTaskNames recursively collects all task names from an execution plan subtree.
func collectTaskNames(tree interface{}, out map[string]struct{}) {
	switch t := tree.(type) {
	case map[string]interface{}:
		for k, v := range t {
			out[k] = struct{}{}
			collectTaskNames(v, out)
		}
	case map[string]struct{}:
		for k := range t {
			out[k] = struct{}{}
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
