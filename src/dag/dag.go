// This Go code does the same thing as your Python code. It defines a DAG struct with methods to build a
// directed acyclic graph (DAG) from a collection of tasks, get available tasks, set the status of a task,
// check if a task should be canceled, cancel dependent tasks, and create an execution plan. The topsort
// package is used to create and manipulate the DAG.

package dag

import "gotasker/src/graph"

type DAG struct {
	taskCollection      []map[string]interface{}
	reverse             bool
	graph               *graph.DependencyGraph
	dependencyTree      map[string][]string
	toBeCanceled        map[string]struct{}
	finishedTasksStatus map[string]map[string]struct{}
	executionPlan       map[string]interface{}
}

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

func (d *DAG) GetAvailableTasks() []string {
	return d.graph.TopSorted()
}

func (d *DAG) GetExecutionPlan() map[string]interface{} {
	return d.executionPlan
}

func (d *DAG) SetStatus(taskName string, status string) {
	d.finishedTasksStatus[status][taskName] = struct{}{}
}

func (d *DAG) ShouldBeCanceled(taskName string) bool {
	_, ok := d.toBeCanceled[taskName]
	return ok
}

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

func getAllTaskSet(tasks map[string]struct{}) map[string]struct{} {
	taskSet := make(map[string]struct{})
	for task := range tasks {
		taskSet[task] = struct{}{}
	}
	return taskSet
}

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

func (d *DAG) createExecutionPlan(dependencyDict map[string][]string) map[string]interface{} {
	executionPlan := make(map[string]interface{})
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
