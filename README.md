# GoExecutorService

An executor service that helps to schedule an execution of a lambda function with `goroutine` executor pool.
The service creates a fixed size queue for items to be processed and runs fixed number of goroutines (workers) 
which pull items from the queue and execute lambda functions one by one.

An executor service can be stopped at any time. After stopping a service all unprocessed items in a service's queue are dismissed, queue is close, and all items are being processed will finish their work, and all workers are stopped finally.

## Import

To use the executor service in your application add following import.

```go
import (
	"github.com/alex-53-8/go-executor-service/executor"
)
```

## Usage

Specifying a goroutines pool size and size of a queue

```go
import (
	"github.com/alex-53-8/go-executor-service/executor"
)
//....

executor, _ := executor.CreateExecutor(executor.ExecutorConfig{PoolSize: 4, QueueSize: 100})

// schedule invocation of a procedure
error := executor.Schedule(func() { /* your code here */ })
// error should be `nil`
// ...
// do some other stuff
// ...
// do not need an executor any more, stopping it
// all in-progress items will continue processing and
// all unprocessed items in a queue are dismissed
executor.Stop()

// it is not possible to schedule new items after stopping
error = executor.Schedule(func() {})
// error is not nil, it indicates failed scheduling now
```

## Advanced usage

Here is another example: it is possible to specify a final callback which is called when all workers are done

```go
import (
	"github.com/alex-53-8/go-executor-service/executor"
)
//....
	var onDoneWait sync.WaitGroup
	onDoneWait.Add(1)

	queueSize := 10000
	poolSize := 4

	// creating an executor
	cfg := executor.ExecutorConfig{PoolSize: poolSize, QueueSize: queueSize, OnAllWorkersStopped: func() {
		// called when all workers are exited
		log.Println("on all workers are done")
		onDoneWait.Done()
	}}
	executor, _ := executor.CreateExecutor(cfg)

	// scheduling all jobs here
	for i := 0; i < queueSize; i++ {
		executor.Schedule(func() {
			log.Printf("I'm your lambda function %d\n", i)
		})
	}

	// stopping an executor
	executor.Stop()

	// already it is not possible to schedule - service is already stopped
	err := executor.Schedule(func() { /* some code here, does not matter - it will be executed*/ })
	// err variable is not nil here

	// wait until all worker are done
	onDoneWait.Wait()

    // continue code execution
```