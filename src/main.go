// Main file to execute the workflow processor.
package main

import (
	"flag"
	"fmt"
	"gotasker/src/engine"
	"gotasker/src/workflow"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {
	filePath := flag.String("file", "", "Path to the workflow YAML or JSON file (required)")
	flag.StringVar(filePath, "f", "", "Path to the workflow YAML or JSON file (shorthand)")

	dryRun := flag.Bool("dry-run", false, "Print execution plan without running tasks")
	flag.BoolVar(dryRun, "d", false, "Print execution plan without running tasks (shorthand)")

	threads := flag.Int("threads", runtime.NumCPU(), "Maximum number of parallel tasks")
	flag.IntVar(threads, "t", runtime.NumCPU(), "Maximum number of parallel tasks (shorthand)")

	flag.Parse()

	if *filePath == "" {
		fmt.Fprintln(os.Stderr, "Error: workflow file path is required. Use -file or -f flag.")
		flag.Usage()
		os.Exit(1)
	}

	if *threads < 1 {
		*threads = 1
	}

	// Load workflow
	wf, err := workflow.NewWorkflow(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading workflow: %v\n", err)
		os.Exit(1)
	}

	// Create engine
	eng, err := engine.NewEngine(wf, *threads, *dryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating engine: %v\n", err)
		os.Exit(1)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		fmt.Fprintf(os.Stderr, "\nReceived signal: %v\n", sig)
		eng.AbortExecution()
	}()

	// Run the engine
	if err := eng.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
		os.Exit(1)
	}
}
