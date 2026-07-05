# GoTasker

A small CLI workflow runner. Describe your tasks in YAML or JSON, declare dependencies between them, and GoTasker builds a dependency graph (DAG) and runs independent tasks in parallel — one topological layer at a time.

## Features

- [x] **Parallel execution** — independent tasks run concurrently, bounded by a configurable thread count
- [x] **YAML and JSON input** — write workflows in either format
- [x] **Terminal commands** — each task runs a command with arbitrary args
- [x] **Reusable workflows** — import tasks from other files with a namespace prefix
- [x] **Dependency DAG** — `depends-on` builds the execution order; cycles and self-references are rejected
- [x] **`foreach` expansion** — generate one task per combination of list variables
- [x] **Templating** — `{{.variable}}` placeholders resolved from `variables` (including in task names)
- [x] **Dry run** — print the execution plan without running anything
- [x] **Graceful shutdown** — SIGINT/SIGTERM cancels pending tasks

## Build & run

```bash
go build -o gotasker ./src

# or run directly
go run ./src -f examples/test.yaml
```

### CLI flags

| Flag | Shorthand | Description | Default |
|------|-----------|-------------|---------|
| `-file` | `-f` | Path to the workflow YAML/JSON file (required) | — |
| `-threads` | `-t` | Maximum number of parallel tasks | number of CPUs |
| `-dry-run` | `-d` | Print the execution plan without running tasks | `false` |

```bash
go run ./src -f examples/test.json -t 4
go run ./src -f examples/main_with_imports.yaml -d
```

## Workflow file format

```yaml
name: my workflow
description: Optional description
variables:
  greeting: "world"
  names:
    - Pepe
    - Juan
tasks:
  - name: "greet-{{.name}}"
    do:
      this: process           # action type (currently only "process")
      with:
        path: echo            # the binary to run
        args:
          - "Hello {{.name}}!"
    depends-on:               # optional; names of tasks that must finish first
      - "setup"
    foreach:                  # optional; expands into one task per list item
      - variable: names       # a list variable defined above
        as: name              # bound name used in placeholders
```

- **`do.with.args`** entries are plain strings, or maps that render as `--key=value` flags (list values repeat the flag).
- **`foreach`** with multiple loops produces the Cartesian product of the referenced list variables.
- Task names are templated, which is how expanded `foreach` tasks stay unique.

### Reusable workflows (imports)

```yaml
imports:
  - file: shared_tasks.yaml
    as: logging
tasks:
  - name: "after-logging"
    depends-on:
      - "logging.log-complete"   # imported task names are prefixed with the namespace
    do:
      this: process
      with:
        path: echo
        args: ["done"]
```

Imported task names and their `depends-on` references are prefixed with the `as` namespace to avoid collisions. Imported variables are merged in only if not already defined in the main file.

See [`examples/`](examples/) for complete YAML and JSON workflows.

## Development

```bash
go test ./...                                    # all tests live in the ./tests package
go test ./tests -run TestNewDAG                  # a single test
go vet ./...
golint -set_exit_status ./...

# coverage (CI enforces a 75% total threshold)
go test ./... -coverprofile coverage.out -covermode count
go tool cover -func coverage.out
```

CI (`.github/workflows/code-quality-checks.yaml`) runs `golint`, `go vet`, the test suite, and a 75% coverage gate.

## Architecture

The flow is one-directional across packages under `src/`:

- **`workflow`** — parses the file, expands `foreach`, resolves `{{.var}}` templates, and merges imports.
- **`graph`** — generic dependency graph; `TopSortedLayers()` groups tasks into parallel-executable layers.
- **`dag`** — wraps the graph with task status and cancellation policies.
- **`runner`** — executes a command via `os/exec`.
- **`engine`** — orchestrates: walks the layers and runs each in parallel under a thread-count semaphore.

## Roadmap

- [ ] **`cleanup` actions** — the `cleanup` field is parsed but not yet executed on task completion/failure
- [ ] **Configurable cancellation policy** — currently hardcoded to `abort-related-flows`; expose `abort-all` / `continue` per workflow
- [ ] **Additional action types** — `do.this` only supports `process` today
- [ ] **Surface task `description`** — accepted in the file but not yet used in output
