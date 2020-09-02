/*
Package runner provides a task runner that allows to execute a list of functions with custom hooks and intervals between each call.
It features two main methods: `*Program.Run()` and `*Program.RunConc(n int)`. The latter dispatches the work among n runners (goroutines), resulting in a huge boost of speed.
*/
package runner

import (
	"sort"
	"time"
)

/*
Program represents the task list to be run with its options.

	- `Tasks` (**Required**): a slice of Task, the functions to be executed

	- `Interval`: time between two Task.

	- `PreHook`: callback executed before each Task (`i`: current Task index).

	-`PostHook`: callback executed after each Task (`i`: current Task index,
		`v`: current Task returned value).

	-`OnStart`: callback executed before the program starts.

	-`OnFinish`: callback executed after the program ends (v = slice of all
	task results)
*/
type Program struct {
	Tasks    []Task
	Interval time.Duration
	PreHook  func(i int)
	PostHook func(i int, v interface{})
	OnStart  func()
	OnFinish func(v []interface{})
	isConc   bool
}

// Task is a function run by *FuncList.Run().
type Task func() interface{}

type trackedResult struct {
	id  int
	res []interface{}
}

// Run runs a *Program sequentially, executing all specified callbacks.
// It returns a slice of all Tasks return value.
func (p *Program) Run() (results []interface{}) {
	// OnStart/OnFinish hooks
	// TODO: refacto to avoid repetition with p.RunConc()
	if p.OnStart != nil {
		p.OnStart()
	}
	defer func() {
		if p.OnFinish != nil {
			p.OnFinish(results)
		}
	}()

	results = p.run(p.Tasks, 0)
	return
}

/*
RunConc runs a *Program concurrently, dispatching its tasks among `n` runners
(goroutines), and return a slice of each Task return value.

Details:

Each runner is allocated a range (a subslice of the original
Task slice) which length is calculated upon the total length and the number
of runners. For instance, for 10 Task and 3 runners:
run0[0:3] run1[3:6] run2[6:10].

Then each subslice is run concurrently and their result stored in a channel
that also contain the runner ID. This is necessary when gathering the results,
because they don't necessarily return in the correct order. Associating a
runner ID to a set of results allows to re-order them properly.

The process of retrieving results can largely be optimized as there are
many consecutive loops and sorting operations.
*/
func (p *Program) RunConc(n int) (results []interface{}) {
	// In case run() method needs to know whether
	// p.concMode(true)
	// defer func() { p.concMode(false) }()

	// OnStart/OnFinish hooks
	// TODO: refacto to avoid repetition with p.Run()
	if p.OnStart != nil {
		p.OnStart()
	}
	defer func() {
		if p.OnFinish != nil {
			p.OnFinish(results)
		}
	}()

	length := len(p.Tasks)
	runners := safeRunnerQuantity(n, length)
	span := length / runners
	rawPartial := make(chan trackedResult, runners)

	// Dispatch tasks into zones for each runner (n)
	// and run a goroutine for each zone
	for i := 0; i < runners; i++ {
		isLastRunner := i == runners-1
		start := i * span
		end := start + span
		// Last runner goes to the end
		if isLastRunner {
			end = length
		}
		part := p.Tasks[start:end]

		// Run concurrently. Variable i is used to keep track of the runner
		// in order to sort the final slice in the correct order.
		go func(i int) {
			rawPartial <- trackedResult{i, p.run(part, i)}
		}(i)
	}

	// Gather results
	var rawResults []trackedResult
	for i := 0; i < runners; i++ {
		rawResults = append(rawResults, <-rawPartial)
	}

	// Sort results in the correct order using runner id
	sort.SliceStable(rawResults, func(i, j int) bool {
		return rawResults[i].id < rawResults[j].id
	})

	// Remove ids from results
	for _, v := range rawResults {
		results = append(results, v.res...)
	}

	return
}

func (p *Program) run(tasks []Task, offset int) []interface{} {
	l := len(tasks)
	results := make([]interface{}, l)

	for i, f := range tasks {
		isLast := i == l-1

		if p.PreHook != nil {
			p.PreHook(i + offset)
		}

		v := f()
		results[i] = v

		if p.PostHook != nil {
			p.PostHook(i+offset, v)
		}

		if p.Interval > 0 && !isLast {
			time.Sleep(p.Interval)
		}
	}

	return results
}

// concMode sets p.isConc accordingly to activate bool.
// Unused for now.
func (p *Program) concMode(activate bool) {
	if activate {
		p.isConc = true
		return
	}
	p.isConc = false
}

// safeRunnerQuantity returns a safe amount of runners depending on
// []Task length (there cannot be more runners than tasks). It also
// ensures runners don't go below 1.
func safeRunnerQuantity(runners, length int) int {
	if runners > length {
		return length
	}
	if runners < 1 {
		return 1
	}
	return runners
}
