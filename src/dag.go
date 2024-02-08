// This Go code does the same thing as your Python code. It defines a DAG struct with methods to build a
// directed acyclic graph (DAG) from a collection of tasks, get available tasks, set the status of a task,
// check if a task should be canceled, cancel dependent tasks, and create an execution plan. The topsort
// package is used to create and manipulate the DAG.

// Please replace taskCollection := []map[string]interface{}{} and reverse := false with your actual task
// collection and reverse flag.

// Remember that error handling in Go is explicit, and it’s a good practice to always check for errors
// where they can occur.



package main

import (
	"fmt"
	"github.com/stevenle/topsort"
)

type DAG struct {
	TaskCollection       []map[string]interface{}
	Reverse              bool
	Dag                  *topsort.Graph
	DependencyTree       map[string][]string
	ToBeCanceled         map[string]bool
	FinishedTasksStatus  map[string]map[string]bool
	ExecutionPlan        map[string]map[string]interface{}
}

func NewDAG(taskCollection []map[string]interface{}, reverse bool) *DAG {
	dag := topsort.NewGraph()
	dependencyTree := make(map[string][]string)
	for _, task := range taskCollection {
		taskName := task["task"].(string)
		dependencies := task["depends-on"].([]interface{})
		for _, dependency := range dependencies {
			if reverse {
				dag.AddEdge(dependency.(string), taskName)
			} else {
				dag.AddEdge(taskName, dependency.(string))
			}
			dependencyTree[taskName] = append(dependencyTree[taskName], dependency.(string))
		}
	}
	finishedTasksStatus := map[string]map[string]bool{
		"failed":     make(map[string]bool),
		"canceled":   make(map[string]bool),
		"successful": make(map[string]bool),
	}
	executionPlan := createExecutionPlan(dependencyTree)
	return &DAG{
		TaskCollection:      taskCollection,
		Reverse:             reverse,
		Dag:                 dag,
		DependencyTree:      dependencyTree,
		ToBeCanceled:        make(map[string]bool),
		FinishedTasksStatus: finishedTasksStatus,
		ExecutionPlan:       executionPlan,
	}
}

func (d *DAG) IsActive() bool {
	return len(d.Dag.Ready()) > 0
}

func (d *DAG) GetAvailableTasks() []string {
	return d.Dag.Ready()
}

func (d *DAG) GetExecutionPlan() map[string]map[string]interface{} {
	return d.ExecutionPlan
}

func (d *DAG) SetStatus(taskName string, status string) {
	d.FinishedTasksStatus[status][taskName] = true
	d.Dag.RemoveNode(taskName)
}

func (d *DAG) ShouldBeCanceled(taskName string) bool {
	return d.ToBeCanceled[taskName]
}

func (d *DAG) CancelDependantTasks(taskName string, cancelPolicy string) {
	getAllTaskSet := func(tasks map[string]map[string]interface{}) map[string]bool {
		taskSet := make(map[string]bool)
		for task, subTasks := range tasks {
			taskSet[task] = true
			for subTask := range getAllTaskSet(subTasks) {
				taskSet[subTask] = true
			}
		}
		return taskSet
	}
	if cancelPolicy == "continue" {
		return
	}
	notCancelledTasks := make(map[string]bool)
	for task := range d.FinishedTasksStatus["failed"] {
		notCancelledTasks[task] = true
	}
	for task := range d.FinishedTasksStatus["successful"] {
		notCancelledTasks[task] = true
	}
	for rootTask, subTasks := range d.ExecutionPlan {
		taskSet := getAllTaskSet(map[string]map[string]interface{}{rootTask: subTasks})
		if cancelPolicy == "abort-all" {
			for task := range taskSet {
				d.ToBeCanceled[task] = true
			}
		} else if cancelPolicy == "abort-related-flows" {
			if taskSet[taskName] {
				for task := range taskSet {
					if !notCancelledTasks[task] {
						d.ToBeCanceled[task] = true
					}
				}
			}
		} else {
			fmt.Printf("Unknown cancel policy '%s'\n", cancelPolicy)
		}
	}
}

func createExecutionPlan(dependencyDict map[string][]string) map[string]map[string]interface{} {
	executionPlan := make(map[string]map[string]interface{})
	getRootTasks := func(dependencyDict map[string][]string) []string {
		allTasks := make(map[string]bool)
		for task := range dependencyDict {
			allTasks[task] = true
		}
		dependentTasks := make(map[string]bool)
		for _, dependencies := range dependencyDict {
			for _, dependency := range dependencies {
				dependentTasks[dependency] = true
			}
		}
		var rootTasks []string
		for task := range allTasks {
			if !dependentTasks[task] {
				rootTasks = append(rootTasks, task)
			}
		}
		return rootTasks
	}
	getSubtaskPlan := func(taskName string, dependencyDict map[string][]string, level int) map[string]map[string]interface{} {
		if _, ok := dependencyDict[taskName]; !ok {
			return map[string]map[string]interface{}{taskName: make(map[string]interface{})}
		}
		dependencies := dependencyDict[taskName]
		plan := map[string]map[string]interface{}{taskName: make(map[string]interface{})}
		for _, dependency := range dependencies {
			subPlan := getSubtaskPlan(dependency, dependencyDict, level+1)
			for task, subTasks := range subPlan {
				plan[taskName][task] = subTasks
			}
		}
		return plan
	}
	rootTasks := getRootTasks(dependencyDict)
	for _, rootTask := range rootTasks {
		for task, subTasks := range getSubtaskPlan(rootTask, dependencyDict) {
			executionPlan[task] = subTasks
		}
	}
	return executionPlan
}

func main() {
	// TODO: Initialize taskCollection and reverse
	taskCollection := []map[string]interface{}{}
	reverse := false
	dag := NewDAG(taskCollection, reverse)
	fmt.Println(dag.GetExecutionPlan())
}
