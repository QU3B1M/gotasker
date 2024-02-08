// This Go code does the same thing as your Python code. It defines a WorkflowProcessor struct with methods to execute a task,
// create a task object, execute tasks in parallel, run the workflow, and abort execution. The sync.WaitGroup is used to wait
// for all goroutines to finish.

// Please replace workflowFilePath := "", dryRun := false, and threads := 0 with your actual workflow file path, dry run flag,
// and number of threads. Also, replace // TODO: Implement TASKS_HANDLERS and // TODO: Implement taskObject.Execute() with your
// actual task handlers and task execution code. Finally, replace // TODO: Implement AbortExecution with your actual abort execution code.

package main

import (
	"fmt"
	"sync"
	"time"
)

type WorkflowProcessor struct {
	TaskCollection []map[string]interface{}
	DryRun         bool
	Threads        int
}

func NewWorkflowProcessor(workflowFilePath string, dryRun bool, threads int) *WorkflowProcessor {
	workflowFile := NewWorkflowFile(workflowFilePath, "")
	return &WorkflowProcessor{
		TaskCollection: workflowFile.TaskCollection,
		DryRun:         dryRun,
		Threads:        threads,
	}
}

func (w *WorkflowProcessor) ExecuteTask(dag *DAG, task map[string]interface{}, action string) {
	taskName := task["task"].(string)
	if dag.ShouldBeCanceled(taskName) {
		fmt.Printf("[%s] Skipping task due to dependency failure.\n", taskName)
		dag.SetStatus(taskName, "canceled")
	} else {
		taskObject := w.CreateTaskObject(task, action)
		fmt.Printf("[%s] Starting task.\n", taskName)
		startTime := time.Now()
		// TODO: Implement taskObject.Execute()
		fmt.Printf("[%s] Finished task in %.2f seconds.\n", taskName, time.Since(startTime).Seconds())
		dag.SetStatus(taskName, "successful")
	}
}

func (w *WorkflowProcessor) CreateTaskObject(task map[string]interface{}, action string) Task {
	// TODO: Implement TASKS_HANDLERS
	taskType := task[action].(map[string]interface{})["this"].(string)
	taskHandler := TASKS_HANDLERS[taskType]
	if taskHandler != nil {
		return taskHandler(task["task"].(string), task[action].(map[string]interface{})["with"].(map[string]interface{}))
	}
	panic(fmt.Sprintf("Unknown task type '%s'.", taskType))
}

func (w *WorkflowProcessor) ExecuteTasksParallel() {
	if !w.DryRun {
		fmt.Println("Executing tasks in parallel.")
		dag := NewDAG(w.TaskCollection, false)
		var wg sync.WaitGroup
		for len(dag.GetAvailableTasks()) > 0 {
			taskName := dag.GetAvailableTasks()[0]
			task := func() map[string]interface{} {
				for _, t := range w.TaskCollection {
					if t["task"].(string) == taskName {
						return t
					}
				}
				return nil
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				w.ExecuteTask(dag, task, "do")
			}()
		}
		wg.Wait()
		fmt.Println("Executing cleanup tasks.")
		reverseDag := NewDAG(w.TaskCollection, true)
		for len(reverseDag.GetAvailableTasks()) > 0 {
			taskName := reverseDag.GetAvailableTasks()[0]
			task := func() map[string]interface{} {
				for _, t := range w.TaskCollection {
					if t["task"].(string) == taskName {
						return t
					}
				}
				return nil
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				if _, ok := task["cleanup"]; ok {
					w.ExecuteTask(reverseDag, task, "cleanup")
				} else {
					reverseDag.SetStatus(taskName, "successful")
				}
			}()
		}
		wg.Wait()
	} else {
		NewDAG(w.TaskCollection, false)
	}
}

func (w *WorkflowProcessor) Run() {
	w.ExecuteTasksParallel()
}

func (w *WorkflowProcessor) AbortExecution() {
	// TODO: Implement AbortExecution
}

func main() {
	// TODO: Initialize workflowFilePath, dryRun, and threads
	workflowFilePath := ""
	dryRun := false
	threads := 0
	workflowProcessor := NewWorkflowProcessor(workflowFilePath, dryRun, threads)
	workflowProcessor.Run()
}
