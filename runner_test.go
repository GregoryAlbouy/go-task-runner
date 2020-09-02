package runner

import (
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

var example = Program{
	Tasks:    tasks(3),
	Interval: 1 * time.Second,
	PreHook:  func(i int) { log.Printf("Starting task %d...\n", i) },
	PostHook: func(i int, v interface{}) { log.Printf("Task %d done. Output: %v\n", i, v) },
	OnStart:  func() { log.Println("Starting program.") },
	OnFinish: func(v []interface{}) { log.Printf("Program over. Final output: %v\n", v) },
}

func TestExample(t *testing.T) {
	fmt.Println("results:", example.Run())
}

func TestAll(t *testing.T) {
	tests := map[string]func(t *testing.T){
		"Hooks":    TestHooks,
		"Results":  TestResults,
		"Interval": TestInterval,
	}

	for test, testFunc := range tests {
		if ok := t.Run(test, testFunc); !ok {
			t.Errorf("%s did not pass", test)
		}
	}
}

func TestHooks(t *testing.T) {
	res := map[string]bool{
		"PreHook":  false,
		"PostHook": false,
		"OnStart":  false,
		"OnFinish": false,
	}

	preHook := func(i int) { res["PreHook"] = true }
	postHook := func(i int, v interface{}) { res["PostHook"] = true }
	onStart := func() { res["OnStart"] = true }
	onFinish := func(v []interface{}) { res["OnFinish"] = true }

	p := Program{
		Tasks:    tasks(5),
		PreHook:  preHook,
		PostHook: postHook,
		OnStart:  onStart,
		OnFinish: onFinish,
	}

	p.Run()

	for hook, ok := range res {
		if !ok {
			t.Errorf("hook %s not called", hook)
		}
	}
}

func TestResults(t *testing.T) {
	res := (&Program{Tasks: tasks(5)}).Run()
	expect := []interface{}{"task0", "task1", "task2", "task3", "task4"}

	if !reflect.DeepEqual(expect, res) {
		t.Errorf("expected %v\ngot %v\n", expect, res)
	}
}

func TestInterval(t *testing.T) {
	const (
		expectedMin = 3500 * time.Millisecond
		expectedMax = 4500 * time.Millisecond
	)

	p := Program{
		Tasks:    tasks(5),
		Interval: 1 * time.Second,
	}
	start := time.Now()
	p.Run()
	interval := time.Since(start)

	if interval < expectedMin {
		t.Errorf("expected duration < %v, got %v", expectedMin, interval)
	}

	if interval > expectedMax {
		t.Errorf("expected duration > %v, got %v", expectedMax, interval)
	}
}

func BenchmarkRun(b *testing.B) {
	p := &Program{Tasks: heavyTasks(4000, 1000)}
	p.Run()
}

func BenchmarkRunConc(b *testing.B) {
	p := &Program{Tasks: heavyTasks(4000, 1000)}
	p.RunConc(8)
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

// heavyTasks returns a slice of n func() interface{}, each returning
// a slice of length l
func heavyTasks(n, l int) (ts []Task) {
	const (
		randMax = 1000000
		// length  = 1000
	)

	for i := 0; i < n; i++ {
		ts = append(ts, func() interface{} {
			longSlice := make([]int, l)

			for i := range longSlice {
				s := rand.NewSource(time.Now().UnixNano())
				r := rand.New(s)
				longSlice[i] = r.Intn(randMax) + 1
			}

			// sort.Ints(longSlice)

			return longSlice
		})
	}
	return
}
