// Package engine is responsible for starting and executing the workflow.
// It uses the graph package to create a directed acyclic graph (DAG) and
// execute tasks in parallel.
package engine

import (
	"encoding/json"
	"fmt"
	"gotasker/src/dag"
	"gotasker/src/runner"
	"gotasker/src/workflow"
	"sync"
)

// Engine is the main struct for the engine package. It contains the task collection and the DAG.
type Engine struct {
	TaskCollection []workflow.Task
	DAG            *dag.DAG
	Threads        int
	aborted        bool
	mu             sync.Mutex
	DryRun         bool
}

// NewEngine creates a new Engine with the given task collection.
func NewEngine(wf *workflow.Workflow, threads int, dryRun bool) (*Engine, error) {
	var tasks []map[string]interface{}
	wfTasks := wf.Tasks
	jsonTasks, err := json.Marshal(&wfTasks)
	if err != nil {
		return nil, fmt.Errorf("error marshalling tasks: %w", err)
	}
	err = json.Unmarshal(jsonTasks, &tasks)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling tasks: %w", err)
	}
	return &Engine{
		TaskCollection: wfTasks,
		DAG:            dag.NewDAG(tasks, false),
		Threads:        threads,
		DryRun:         dryRun,
	}, nil
}

// getTaskByName finds a task in the task collection by its name.
func (w *Engine) getTaskByName(name string) *workflow.Task {
	for i := range w.TaskCollection {
		if w.TaskCollection[i].Name == name {
			return &w.TaskCollection[i]
		}
	}
	return nil
}

// ExecuteTask executes a single task and returns its output or an error.
func (w *Engine) ExecuteTask(task *workflow.Task) (string, error) {
	fmt.Printf("Executing task: %s\n", task.Name)

	// Build the action params map for the runner
	params := map[string]interface{}{
		"path": task.Do.With.Path,
		"args": func() []interface{} {
			args := make([]interface{}, len(task.Do.With.Args))
			copy(args, task.Do.With.Args)
			return args
		}(),
	}

	execution := runner.NewExecution(task.Do.With.Path, params)
	output, err := execution.Execute()
	if err != nil {
		return "", fmt.Errorf("task %s failed: %w", task.Name, err)
	}

	fmt.Printf("Task %s completed. Output: %s", task.Name, output)
	return output, nil
}

// ExecuteTaskLayerParallel executes all tasks in a layer in parallel,
// limited by the configured number of threads.
func (w *Engine) ExecuteTaskLayerParallel(layer []string) map[string]error {
	results := make(map[string]error)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Semaphore channel to limit concurrency
	sem := make(chan struct{}, w.Threads)

	for _, taskName := range layer {
		// Check if we should abort
		w.mu.Lock()
		if w.aborted {
			w.mu.Unlock()
			mu.Lock()
			results[taskName] = fmt.Errorf("execution aborted")
			mu.Unlock()
			w.DAG.SetStatus(taskName, "canceled")
			continue
		}
		w.mu.Unlock()

		// Check if the task is marked for cancellation
		if _, ok := w.DAG.GetTasksToCancel()[taskName]; ok {
			mu.Lock()
			results[taskName] = fmt.Errorf("task canceled")
			mu.Unlock()
			w.DAG.SetStatus(taskName, "canceled")
			continue
		}

		task := w.getTaskByName(taskName)
		if task == nil {
			mu.Lock()
			results[taskName] = fmt.Errorf("task %s not found in collection", taskName)
			mu.Unlock()
			w.DAG.SetStatus(taskName, "failed")
			continue
		}

		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore slot
		go func(t *workflow.Task, name string) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore slot

			_, err := w.ExecuteTask(t)
			mu.Lock()
			results[name] = err
			mu.Unlock()

			if err != nil {
				w.DAG.SetStatus(name, "failed")
			} else {
				w.DAG.SetStatus(name, "successful")
			}
		}(task, taskName)
	}

	wg.Wait()
	return results
}

// Run starts the workflow execution. It processes layers from the DAG
// topological sort and executes each layer in parallel.
func (w *Engine) Run() error {
	if w.DryRun {
		w.PrintExecutionPlan()
		return nil
	}

	// Check if already aborted before starting
	w.mu.Lock()
	if w.aborted {
		w.mu.Unlock()
		fmt.Println("Execution aborted.")
		return fmt.Errorf("execution aborted")
	}
	w.mu.Unlock()

	layers := w.DAG.GetTopSortedLayers()

	for i, layer := range layers {
		// Check abort
		w.mu.Lock()
		if w.aborted {
			w.mu.Unlock()
			fmt.Println("Execution aborted.")
			return fmt.Errorf("execution aborted")
		}
		w.mu.Unlock()

		fmt.Printf("--- Layer %d: %v ---\n", i+1, layer)

		results := w.ExecuteTaskLayerParallel(layer)

		// Check for failures and handle cancel policies
		for taskName, err := range results {
			if err != nil {
				fmt.Printf("Task %s failed: %v\n", taskName, err)
				// Cancel dependent tasks using abort-related-flows policy
				w.DAG.CancelDependentTasks(taskName, "abort-related-flows")
			}
		}
	}

	// Print summary
	fmt.Println("\n=== Execution Summary ===")
	for _, task := range w.TaskCollection {
		status := w.DAG.GetStatus(task.Name)
		fmt.Printf("  %s: %s\n", task.Name, status)
	}

	return nil
}

// PrintExecutionPlan prints the execution plan without running tasks.
func (w *Engine) PrintExecutionPlan() {
	fmt.Println("=== Execution Plan (Dry Run) ===")
	layers := w.DAG.GetTopSortedLayers()
	for i, layer := range layers {
		fmt.Printf("Layer %d:\n", i+1)
		for _, taskName := range layer {
			task := w.getTaskByName(taskName)
			if task != nil {
				fmt.Printf("  - %s: %s %v\n", task.Name, task.Do.With.Path, task.Do.With.Args)
			} else {
				fmt.Printf("  - %s: (not found)\n", taskName)
			}
		}
	}
}

// AbortExecution aborts the execution of the workflow processor.
func (w *Engine) AbortExecution() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.aborted = true
	// Cancel all pending tasks
	for _, task := range w.TaskCollection {
		w.DAG.CancelTask(task.Name)
	}
	fmt.Println("Abort signal received. Canceling pending tasks...")
}
