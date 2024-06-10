# Goroutine executor

An executor service that helps to schedule an execution with `goroutine`. There are permanently 
configured number of `goroutine` that execute from a queue procedures.


## Usage

```go
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