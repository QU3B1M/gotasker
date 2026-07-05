// Unit tests for the "gotasker/src/engine" package.
package tests

import (
	"gotasker/src/engine"
	"gotasker/src/workflow"
	"testing"
)

func newTestWorkflow(tasks []workflow.Task) *workflow.Workflow {
	return &workflow.Workflow{
		Tasks:     tasks,
		Variables: map[string]interface{}{},
	}
}

func TestNewEngine(t *testing.T) {
	wf := newTestWorkflow([]workflow.Task{
		{
			Name: "task-a",
			Do: workflow.Action{
				This: "process",
				With: workflow.With{Path: "echo", Args: []interface{}{"hello"}},
			},
		},
	})
	eng, err := engine.NewEngine(wf, 2, false)
	if err != nil {
		t.Fatalf("NewEngine returned error: %v", err)
	}
	if eng == nil {
		t.Fatal("NewEngine returned nil")
	}
	if eng.Threads != 2 {
		t.Errorf("Expected threads=2, got %d", eng.Threads)
	}
}

func TestNewEngineDryRun(t *testing.T) {
	wf := newTestWorkflow([]workflow.Task{
		{
			Name: "task-a",
			Do: workflow.Action{
				This: "process",
				With: workflow.With{Path: "echo", Args: []interface{}{"hello"}},
			},
		},
	})
	eng, err := engine.NewEngine(wf, 1, true)
	if err != nil {
		t.Fatalf("NewEngine returned error: %v", err)
	}
	if !eng.DryRun {
		t.Error("Expected DryRun=true")
	}
}

func TestExecuteTask(t *testing.T) {
	wf := newTestWorkflow([]workflow.Task{
		{
			Name: "echo-test",
			Do: workflow.Action{
				This: "process",
				With: workflow.With{
					Path: "echo",
					Args: []interface{}{"hello world"},
				},
			},
		},
	})
	eng, err := engine.NewEngine(wf, 1, false)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}
	output, err := eng.ExecuteTask(&wf.Tasks[0])
	if err != nil {
		t.Errorf("ExecuteTask returned error: %v", err)
	}
	if output == "" {
		t.Error("ExecuteTask returned empty output")
	}
}

func TestExecuteTaskError(t *testing.T) {
	wf := newTestWorkflow([]workflow.Task{
		{
			Name: "fail-task",
			Do: workflow.Action{
				This: "process",
				With: workflow.With{
					Path: "nonexistent_command_xyz",
					Args: []interface{}{},
				},
			},
		},
	})
	eng, err := engine.NewEngine(wf, 1, false)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}
	_, err = eng.ExecuteTask(&wf.Tasks[0])
	if err == nil {
		t.Error("ExecuteTask should have returned an error for nonexistent command")
	}
}

func TestExecuteTaskLayerParallel(t *testing.T) {
	wf := newTestWorkflow([]workflow.Task{
		{
			Name: "echo-a",
			Do: workflow.Action{
				This: "process",
				With: workflow.With{Path: "echo", Args: []interface{}{"a"}},
			},
		},
		{
			Name: "echo-b",
			Do: workflow.Action{
				This: "process",
				With: workflow.With{Path: "echo", Args: []interface{}{"b"}},
			},
		},
	})
	eng, err := engine.NewEngine(wf, 2, false)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}
	results := eng.ExecuteTaskLayerParallel([]string{"echo-a", "echo-b"})
	for name, taskErr := range results {
		if taskErr != nil {
			t.Errorf("Task %s failed: %v", name, taskErr)
		}
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestRunSimpleWorkflow(t *testing.T) {
	wf := newTestWorkflow([]workflow.Task{
		{
			Name: "step-1",
			Do: workflow.Action{
				This: "process",
				With: workflow.With{Path: "echo", Args: []interface{}{"step 1"}},
			},
		},
		{
			Name:      "step-2",
			DependsOn: []string{"step-1"},
			Do: workflow.Action{
				This: "process",
				With: workflow.With{Path: "echo", Args: []interface{}{"step 2"}},
			},
		},
	})
	eng, err := engine.NewEngine(wf, 2, false)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}
	err = eng.Run()
	if err != nil {
		t.Errorf("Run returned error: %v", err)
	}
	if eng.DAG.GetStatus("step-1") != "successful" {
		t.Errorf("step-1 status: %s, expected successful", eng.DAG.GetStatus("step-1"))
	}
	if eng.DAG.GetStatus("step-2") != "successful" {
		t.Errorf("step-2 status: %s, expected successful", eng.DAG.GetStatus("step-2"))
	}
}

func TestRunDryRun(t *testing.T) {
	wf := newTestWorkflow([]workflow.Task{
		{
			Name: "dry-task",
			Do: workflow.Action{
				This: "process",
				With: workflow.With{Path: "echo", Args: []interface{}{"should not run"}},
			},
		},
	})
	eng, err := engine.NewEngine(wf, 1, true)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}
	err = eng.Run()
	if err != nil {
		t.Errorf("Run (dry) returned error: %v", err)
	}
	if eng.DAG.GetStatus("dry-task") != "pending" {
		t.Errorf("dry-task should be pending in dry run, got: %s", eng.DAG.GetStatus("dry-task"))
	}
}

func TestAbortExecution(t *testing.T) {
	wf := newTestWorkflow([]workflow.Task{
		{
			Name: "abort-task",
			Do: workflow.Action{
				This: "process",
				With: workflow.With{Path: "echo", Args: []interface{}{"test"}},
			},
		},
	})
	eng, err := engine.NewEngine(wf, 1, false)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}
	eng.AbortExecution()
	err = eng.Run()
	if err == nil {
		t.Error("Run should return error after AbortExecution")
	}
}

func TestRunWithFailedTaskCancelsDependent(t *testing.T) {
	wf := newTestWorkflow([]workflow.Task{
		{
			Name: "will-fail",
			Do: workflow.Action{
				This: "process",
				With: workflow.With{
					Path: "nonexistent_command_xyz",
					Args: []interface{}{},
				},
			},
		},
		{
			Name:      "depends-on-fail",
			DependsOn: []string{"will-fail"},
			Do: workflow.Action{
				This: "process",
				With: workflow.With{Path: "echo", Args: []interface{}{"should not run"}},
			},
		},
	})
	eng, err := engine.NewEngine(wf, 2, false)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}
	_ = eng.Run()
	if eng.DAG.GetStatus("will-fail") != "failed" {
		t.Errorf("will-fail status: %s, expected failed", eng.DAG.GetStatus("will-fail"))
	}
	status := eng.DAG.GetStatus("depends-on-fail")
	if status != "canceled" {
		t.Errorf("depends-on-fail status: %s, expected canceled", status)
	}
}

func TestRunParallelIndependentTasks(t *testing.T) {
	wf := newTestWorkflow([]workflow.Task{
		{
			Name: "parallel-a",
			Do: workflow.Action{
				This: "process",
				With: workflow.With{Path: "echo", Args: []interface{}{"a"}},
			},
		},
		{
			Name: "parallel-b",
			Do: workflow.Action{
				This: "process",
				With: workflow.With{Path: "echo", Args: []interface{}{"b"}},
			},
		},
		{
			Name: "parallel-c",
			Do: workflow.Action{
				This: "process",
				With: workflow.With{Path: "echo", Args: []interface{}{"c"}},
			},
		},
	})
	eng, err := engine.NewEngine(wf, 3, false)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}
	err = eng.Run()
	if err != nil {
		t.Errorf("Run returned error: %v", err)
	}
	for _, name := range []string{"parallel-a", "parallel-b", "parallel-c"} {
		if eng.DAG.GetStatus(name) != "successful" {
			t.Errorf("%s status: %s, expected successful", name, eng.DAG.GetStatus(name))
		}
	}
}
