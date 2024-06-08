package executor

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
)

type Executor interface {
	Schedule(fn func()) error
	Stop() error
}

type ExecutorConfig struct {
	QueueSize uint32
	PoolSize  uint32
}

type executorServiceStatus int32

const stopped executorServiceStatus = 0
const running executorServiceStatus = 1

type executorService struct {
	status executorServiceStatus
	queue  chan func()
	mu     sync.Mutex
}

// Creates "Executor" instance with a specified configuration
func CreateExecutor(cfg ExecutorConfig) (Executor, error) {
	if cfg.PoolSize < 1 {
		return nil, errors.New("pool size cannot be less than 1")
	}
	if cfg.QueueSize < 1 {
		return nil, errors.New("queue size cannot be less that 1")
	}

	var queue = make(chan func(), cfg.QueueSize)
	es := &executorService{status: running, queue: queue}

	go initializeWorkers(es, cfg)

	return es, nil
}

func worker(index int, es *executorService, wait *sync.WaitGroup) {
	defer wait.Done()
	for {
		if es.status == stopped {
			log.Println("gcr[", index, "] service has been stopped, terminating a listener")
			break
		}

		nfn, ok := <-es.queue

		if nfn != nil && ok {
			go nfn()
		} else {
			log.Println("gcr[", index, "] Received nil value, terminating a listener")
		}
	}
}

func initializeWorkers(es *executorService, cfg ExecutorConfig) {
	var wait sync.WaitGroup
	for i := 0; i < int(cfg.PoolSize); i++ {
		wait.Add(1)
		go worker(i, es, &wait)
	}

	// wait when all workers are done
	wait.Wait()

	// then close a queue's channel
	if es.queue != nil {
		log.Println("closing channel")
		close(es.queue)
	}
}

// Add a new function, to be executed, into a queue
// Return an error if not possible to schedule:
// - service is stopped
func (es *executorService) Schedule(fn func()) error {
	if fn == nil {
		return errors.New("function cannot be nil")
	}

	es.mu.Lock()
	defer es.mu.Unlock()

	if es.status == stopped {
		return errors.New("Executor pool is stopped, cannot accept a job")
	}

	es.queue <- fn

	return nil
}

// Stops executor service
// sends a stop signal to each worker
func (es *executorService) Stop() error {
	es.mu.Lock()
	defer es.mu.Unlock()
	log.Println("stopping an executor")

	if es.status == running {
		atomic.StoreInt32((*int32)(&es.status), int32(stopped))
		es.queue <- nil
	} else {
		return errors.New("service already stopped")
	}

	return nil
}
