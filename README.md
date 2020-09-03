# go-task-runner

A Go task runner that allows to execute a list of functions with custom hooks and intervals between each call.  
It features two main methods: `*Program.Run()` and `*Program.RunConc(n int)`. The latter dispatches the work among n runners (goroutines), resulting in a huge boost of speed (details on the docs, comparison below)

Full documentation: :point_right: [GoDoc](https://pkg.go.dev/github.com/gregoryalbouy/go-task-runner?tab=doc)

## Installation

`go get -u github.com/gregoryalbouy/go-task-runner`

## Usage

### Basic example

```Go
package main

import (
    "github.com/gregoryalbouy/go-task-runner"
)

func main() {
    tasks := []runner.Task{
        func() interface{} { return "go" }
        func() interface{} { return 42 }
    }
    p := Program{Tasks: tasks}
    res := p.Run()

    fmt.Println(res) // ["go", 42]
}
```

### Full example

```Go
package main

import (
    "github.com/gregoryalbouy/go-task-runner"
)

var example = runner.Program{
	Tasks:    tasks(3), // generate dummy tasks
	Interval: 1 * time.Second,
	PreHook:  func(i int) { fmt.Printf("Starting task %d...\n", i) },
	PostHook: func(i int, v interface{}) { fmt.Printf("Task %d done. Output: %v\n", i, v) },
	OnStart:  func() { fmt.Println("Starting program.") },
	OnFinish: func(v []interface{}) { fmt.Printf("Program over. Final output: %v\n", v) },
}

func main(t *testing.T) {
	fmt.Println("results:", example.Run())
}

// tasks returns a slice of n Tasks that return "task"+index, e.g. "task2".
func tasks(n int) (ts []Task) {
	for i := 0; i < n; i++ {
		func(i int) {
			ts = append(ts, func() interface{} { return fmt.Sprintf("task%d", i) })
		}(i)
	}
	return
}

/*
Output:
00:00:00 Starting program.
00:00:00 Starting task 0...
00:00:00 Task 0 done. Output: task0
00:00:01 Starting task 1...
00:00:01 Task 1 done. Output: task1
00:00:02 Starting task 2...
00:00:02 Task 2 done. Output: task2
00:00:02 Program over. Final output: [task0 task1 task2]
results: [task0 task1 task2]
*/

```
## Run() / RunConc() benchmarks

Based on `/runner_test.go/Benchmark `and `/runner_test.go/BenchmarkConc`:
* Run(): 35.649s
* RunConc(2): 18.152s
* RunConc(4): 9.580s
* RunConc(10): 6.999s
* RunConc(100): 7.083s
* RunConc(1000): 6.679s

## Todo

- package description
- error handling
- better test coverage for RunConc()
- Chain() / Pipe() *Program method that allows each Task to communicate its return value to the next one