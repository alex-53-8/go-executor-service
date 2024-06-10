package executor

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_GeneralUsage(t *testing.T) {
	var onDoneWait sync.WaitGroup
	onDoneWait.Add(1)

	queueSize := 10000
	poolSize := 4

	// creating an executor
	cfg := ExecutorConfig{PoolSize: poolSize, QueueSize: queueSize, OnAllWorkersStopped: func() {
		// called when all workers are exited
		log.Println("on all workers are done")
		onDoneWait.Done()
	}}
	executor, _ := CreateExecutor(cfg)

	// scheduling all jobs here
	for i := 0; i < queueSize; i++ {
		err := executor.Schedule(func() {
			time.Sleep(time.Duration(100) * time.Millisecond)
			log.Printf("I'm your lambda function %d\n", i)
		})

		assert.Nil(t, err)
	}

	log.Println("queue size: ", executor.QueueLength())
	assert.Equal(t, int(queueSize-poolSize), executor.QueueLength())

	time.Sleep(time.Duration(50) * time.Millisecond)

	// stopping an executor
	assert.Nil(t, executor.Stop())

	// already it is not possible to schedule - service is already stopped
	err := executor.Schedule(func() { /* some code here, does not matter - it will be executed*/ })
	assert.Equal(t, "Executor pool is stopped, cannot accept a job", err.Error())

	//
	onDoneWait.Wait()

	assert.Equal(t, int32(4), executor.ProcessedCount())

	// a final accord - we are here only after `onDoneWait` is passed
	assert.True(t, true)
}
