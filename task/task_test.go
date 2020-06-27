package task_test

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/nickhstr/goweb/task"
	"github.com/stretchr/testify/assert"
)

var names = []string{
	"Iron Man",
	"Captain America",
	"Thor",
	"Hulk",
	"Black Widow",
	"Hawkeye",
	"Black Panther",
	"Ant Man",
	"Spider-man",
	"Thanos",
	"Dr. Strange",
	"Scarlet Witch",
}

type namePrinter struct {
	name string
}

func (np namePrinter) Work() {
	log.Println(np.name)
	// Set low to keep tests fast
	time.Sleep(1 * time.Microsecond)
}

func TestTask(t *testing.T) {
	var (
		assert   = assert.New(t)
		poolSize = 3
		wg       sync.WaitGroup
	)

	tsk := task.New(poolSize)

	for i := 0; i < poolSize; i++ {
		for _, name := range names {
			wg.Add(1)

			np := namePrinter{name}

			go func() {
				err := tsk.Do(context.Background(), np)
				assert.Nil(err)

				wg.Done()
			}()

			// call these to trigger any possible data races
			tsk.Completed()
			tsk.Submitted()
		}
	}

	wg.Wait()
	tsk.Shutdown()

	assert.Equal(poolSize*len(names), int(tsk.Completed()))
	assert.Equal(tsk.Submitted(), tsk.Completed())
}

func TestTaskPoolBusy(t *testing.T) {
	var (
		assert   = assert.New(t)
		poolSize = 5
		wg       sync.WaitGroup
	)

	tsk := task.New(poolSize)

	for i := 0; i < poolSize; i++ {
		for _, name := range names {
			wg.Add(1)

			np := namePrinter{name}

			go func() {
				// timeout very quickly, to prevent some work from submitting
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				defer cancel()

				_ = tsk.Do(ctx, np)

				wg.Done()
			}()

			// call these to trigger any possible data races
			tsk.Completed()
			tsk.Submitted()
		}
	}

	wg.Wait()
	tsk.Shutdown()

	assert.Less(tsk.Completed(), tsk.Submitted())
}

func TestWorkerFunc(t *testing.T) {
	var (
		assert   = assert.New(t)
		poolSize = 3
		wg       sync.WaitGroup
	)

	tsk := task.New(poolSize)

	for _, name := range names {
		wg.Add(1)

		go func(n string) {
			printName := task.WorkerFunc(func() {
				fmt.Println(n)
			})

			_ = tsk.Do(context.Background(), printName)

			wg.Done()
		}(name)

		// call these to trigger any possible data races
		tsk.Completed()
		tsk.Submitted()
	}

	wg.Wait()
	tsk.Shutdown()

	assert.Equal(len(names), int(tsk.Completed()))
	assert.Equal(tsk.Submitted(), tsk.Completed())
}
