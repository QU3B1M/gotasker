// Integration tests for the full gotasker pipeline.
package tests

import (
	"gotasker/src/engine"
	"gotasker/src/workflow"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func getExamplesDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "examples")
}

func TestIntegrationYAMLWorkflow(t *testing.T) {
	yamlPath := filepath.Join(getExamplesDir(), "test.yaml")
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Skipf("Example file not found: %s", yamlPath)
	}

	wf, err := workflow.NewWorkflow(yamlPath)
	if err != nil {
		t.Fatalf("NewWorkflow error: %v", err)
	}
	if len(wf.Tasks) == 0 {
		t.Fatal("Expected tasks in workflow, got 0")
	}

	foundPepe := false
	foundJuan := false
	for _, task := range wf.Tasks {
		if task.Name == "test-Pepe" {
			foundPepe = true
		}
		if task.Name == "test-Juan" {
			foundJuan = true
		}
	}
	if !foundPepe {
		t.Error("Expected task 'test-Pepe' from foreach expansion")
	}
	if !foundJuan {
		t.Error("Expected task 'test-Juan' from foreach expansion")
	}

	eng, err := engine.NewEngine(wf, 2, false)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}
	_ = eng.Run()
}

func TestIntegrationJSONWorkflow(t *testing.T) {
	jsonPath := filepath.Join(getExamplesDir(), "test.json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Skipf("Example file not found: %s", jsonPath)
	}

	wf, err := workflow.NewWorkflow(jsonPath)
	if err != nil {
		t.Fatalf("NewWorkflow error for JSON: %v", err)
	}
	if len(wf.Tasks) == 0 {
		t.Fatal("Expected tasks in JSON workflow, got 0")
	}

	foundPepe := false
	foundJuan := false
	for _, task := range wf.Tasks {
		if task.Name == "test-Pepe" {
			foundPepe = true
		}
		if task.Name == "test-Juan" {
			foundJuan = true
		}
	}
	if !foundPepe {
		t.Error("Expected task 'test-Pepe' from JSON foreach expansion")
	}
	if !foundJuan {
		t.Error("Expected task 'test-Juan' from JSON foreach expansion")
	}

	eng, err := engine.NewEngine(wf, 2, false)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}
	err = eng.Run()
	if err != nil {
		t.Errorf("Run (JSON) returned error: %v", err)
	}
	for _, task := range wf.Tasks {
		status := eng.DAG.GetStatus(task.Name)
		if status != "successful" {
			t.Errorf("Task %s status: %s, expected successful", task.Name, status)
		}
	}
}

func TestIntegrationDryRun(t *testing.T) {
	yamlPath := filepath.Join(getExamplesDir(), "test.yaml")
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Skipf("Example file not found: %s", yamlPath)
	}

	wf, err := workflow.NewWorkflow(yamlPath)
	if err != nil {
		t.Fatalf("NewWorkflow error: %v", err)
	}

	eng, err := engine.NewEngine(wf, 2, true)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}
	err = eng.Run()
	if err != nil {
		t.Errorf("Dry run returned error: %v", err)
	}
	for _, task := range wf.Tasks {
		if eng.DAG.GetStatus(task.Name) != "pending" {
			t.Errorf("Task %s should be pending in dry run, got %s", task.Name, eng.DAG.GetStatus(task.Name))
		}
	}
}

func TestIntegrationReusableWorkflow(t *testing.T) {
	importPath := filepath.Join(getExamplesDir(), "main_with_imports.yaml")
	if _, err := os.Stat(importPath); os.IsNotExist(err) {
		t.Skipf("Example file not found: %s", importPath)
	}

	wf, err := workflow.NewWorkflow(importPath)
	if err != nil {
		t.Fatalf("NewWorkflow error for imports: %v", err)
	}
	if len(wf.Tasks) == 0 {
		t.Fatal("Expected tasks in imported workflow, got 0")
	}

	foundSetupLogger := false
	foundLogComplete := false
	foundGreet := false
	foundAfterLogging := false
	for _, task := range wf.Tasks {
		switch task.Name {
		case "logging.setup-logger":
			foundSetupLogger = true
		case "logging.log-complete":
			foundLogComplete = true
		case "greet":
			foundGreet = true
		case "after-logging":
			foundAfterLogging = true
		}
	}
	if !foundSetupLogger {
		t.Error("Expected imported task 'logging.setup-logger'")
	}
	if !foundLogComplete {
		t.Error("Expected imported task 'logging.log-complete'")
	}
	if !foundGreet {
		t.Error("Expected task 'greet'")
	}
	if !foundAfterLogging {
		t.Error("Expected task 'after-logging'")
	}

	eng, err := engine.NewEngine(wf, 2, false)
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}
	err = eng.Run()
	if err != nil {
		t.Errorf("Run (with imports) returned error: %v", err)
	}
	for _, task := range wf.Tasks {
		status := eng.DAG.GetStatus(task.Name)
		if status != "successful" {
			t.Errorf("Task %s status: %s, expected successful", task.Name, status)
		}
	}
}

func TestIntegrationUnsupportedFileFormat(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "workflow.txt")
	os.WriteFile(tmpFile, []byte("not a workflow"), 0644)

	_, err := workflow.NewWorkflow(tmpFile)
	if err == nil {
		t.Error("Expected error for unsupported file format")
	}
}

func TestIntegrationNonexistentFile(t *testing.T) {
	_, err := workflow.NewWorkflow("/nonexistent/path/workflow.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}
