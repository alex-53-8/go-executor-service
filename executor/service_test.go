package executor

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const queueSize = 10000
const poolSize = 4

func Test_GeneralUsage(t *testing.T) {
	var onDoneWait sync.WaitGroup
	onDoneWait.Add(1)

	// creating an executor
	cfg := ExecutorConfig{PoolSize: poolSize, QueueSize: queueSize, OnAllWorkersStopped: func() {
		// called when all workers are exited
		log.Println("on all workers are done")
		onDoneWait.Done()
	}}
	executor, _ := CreateExecutor(cfg)

	// scheduling all jobs here
	for i := 0; i < queueSize; i++ {
		index := i
		err := executor.Schedule(func() {
			time.Sleep(time.Duration(100) * time.Millisecond)
			log.Printf("I'm your lambda function %d\n", index)
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

func Test_CreateAndStopEmptyService(t *testing.T) {
	var onDoneWait sync.WaitGroup
	onDoneWait.Add(1)

	// creating an executor
	cfg := ExecutorConfig{PoolSize: poolSize, QueueSize: queueSize, OnAllWorkersStopped: func() {
		// called when all workers are exited
		log.Println("on all workers are done")
		onDoneWait.Done()
	}}

	executor, _ := CreateExecutor(cfg)

	// service is not stopped yet, so no error is expected here
	assert.Nil(t, executor.Stop())

	// stopping again a service returns an error
	assert.Equal(t, "service already stopped", executor.Stop().Error())

	onDoneWait.Wait()

	// ensure that nothing was processed
	assert.Equal(t, int32(0), executor.ProcessedCount())
}

func Test_ScheduleNilValuesReturnsError(t *testing.T) {
	var onDoneWait sync.WaitGroup
	onDoneWait.Add(1)

	// creating an executor
	cfg := ExecutorConfig{PoolSize: poolSize, QueueSize: queueSize, OnAllWorkersStopped: func() {
		// called when all workers are exited
		log.Println("on all workers are done")
		onDoneWait.Done()
	}}

	executor, _ := CreateExecutor(cfg)

	// schedule not nil work
	assert.Nil(t, executor.Schedule(func() {
		log.Println("I'm a lambda function")
	}))
	// in case of nil work an error is returned
	assert.NotNil(t, executor.Schedule(nil))

	// sleep for a short time - let a goroutine to be picked up for execution
	time.Sleep(time.Duration(50) * time.Millisecond)

	// stopping
	assert.Nil(t, executor.Stop())

	onDoneWait.Wait()

	// ensure one scheduled piece was processed
	assert.Equal(t, int32(1), executor.ProcessedCount())
}
