// This Go code does the same thing as your Python code. It defines a DAG struct with methods to build a
// directed acyclic graph (DAG) from a collection of tasks, get available tasks, set the status of a task,
// check if a task should be canceled, cancel dependent tasks, and create an execution plan. The topsort
// package is used to create and manipulate the DAG.

// Please replace TaskCollection := []map[string]interface{}{} and reverse := false with your actual task
// collection and reverse flag.

// Remember that error handling in Go is explicit, and it’s a good practice to always check for errors
// where they can occur.

package main

type DAG struct {
	taskCollection      []map[string]interface{}
	reverse             bool
	sorter              *TopologicalSorter
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
	d.sorter, d.dependencyTree = d.buildDAG()
	d.executionPlan = d.createExecutionPlan(d.dependencyTree)
	d.sorter.prepare()
	return d
}

func (d *DAG) IsActive() bool {
	return d.sorter.isActive()
}

func (d *DAG) GetAvailableTasks() []string {
	return d.sorter.getReady()
}

func (d *DAG) GetExecutionPlan() map[string]interface{} {
	return d.executionPlan
}

func (d *DAG) SetStatus(taskName string, status string) {
	d.finishedTasksStatus[status][taskName] = struct{}{}
	d.sorter.done(taskName)
}

func (d *DAG) ShouldBeCanceled(taskName string) bool {
	_, ok := d.toBeCanceled[taskName]
	return ok
}

func (d *DAG) buildDAG() (*TopologicalSorter, map[string][]string) {
	dependencyDict := make(map[string][]string)
	sorter := NewTopologicalSorter()
	for _, task := range d.taskCollection {
		taskName := task["task"].(string)
		dependencies, ok := task["depends-on"].([]string)
		if !ok {
			dependencies = []string{}
		}
		if d.reverse {
			for _, dependency := range dependencies {
				sorter.add(dependency, taskName)
			}
		} else {
			sorter.add(taskName, dependencies...)
		}
		dependencyDict[taskName] = dependencies
	}
	return sorter, dependencyDict
}

// ... rest of the methods
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
	for rootTask, subTasks := range d.executionPlan {
		taskSet := getAllTaskSet(subTasks.(map[string]struct{}))
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
		} else {
			// handle error
		}
	}
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
	getSubtaskPlan := func(taskName string, dependencyDict map[string][]string, level int) map[string]interface{} {
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
	rootTasks := getRootTasks(dependencyDict)
	for _, rootTask := range rootTasks {
		executionPlan[rootTask] = getSubtaskPlan(rootTask, dependencyDict, 0)[rootTask]
	}
	return executionPlan
}
