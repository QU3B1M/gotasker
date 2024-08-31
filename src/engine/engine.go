// Package engine is responsible for starting and executing the workflow.
// It uses the graph package to create a directed acyclic graph (DAG) and
//
//	execute tasks in parallel. The engine package contains the following files:
package engine

import (
	"encoding/json"
	"fmt"
	"gotasker/src/dag"
	"gotasker/src/runner"
	"gotasker/src/workflow"
	// "sync"
)

// Engine is the main struct for the engine package. It contains the task collection and the DAG.
type Engine struct {
	TaskCollection []workflow.Task
	DAG            *dag.DAG
	Threads        int
}

// NewEngine creates a new Engine with the given task collection.
func NewEngine(wf workflow.Workflow, threads int) *Engine {
	var tasks []map[string]interface{}
	wfTasks := wf.Tasks
	jsonTasks, _ := json.Marshal(&wfTasks)
	_ = json.Unmarshal(jsonTasks, &tasks)
	return &Engine{
		TaskCollection: wfTasks,
		DAG:            dag.NewDAG(tasks, false),
		Threads:        threads,
	}
}

// ExecuteTask executes the given task with the given action.
func (w *Engine) ExecuteTask(task map[string]interface{}) {
	fmt.Printf("Executing task %s\n", task["task"].(string))
	execution := runner.NewExecution(task["task"].(string), task["action"].(map[string]interface{}))
	_, err := execution.Execute()
	if err != nil {
		fmt.Printf("Error executing task %s: %v\n", task["task"].(string), err)
	}
}

// ExecuteTasksParallel uses dag.TopSortedLayers to get the tasks that can be executed in parallel (the cases where the task is a list of string)
// func (w *Engine) ExecuteTaskLayerParallel() {
// 	availableTasks := w.DAG.GetExecutionPlan()
// 	for _, task := range availableTasks {
// 		for _, t := range w.TaskCollection {
// 			if t.Name == task.(string) {
// 				var command []map[string]interface{}
// 				jsonTasks, _ := json.Marshal(&t)
// 				_ = json.Unmarshal(jsonTasks, &tasks)
// 				w.ExecuteTask(t.(map[string]interface{}))
// 			}
// 		}
// 	}
// }

// Run starts the workflow processor and executes all tasks.
// func (w *Engine) Run() {
// 	w.ExecuteTaskLayerParallel()
// }

// AbortExecution aborts the execution of the workflow processor.
func (w *Engine) AbortExecution() {

}
