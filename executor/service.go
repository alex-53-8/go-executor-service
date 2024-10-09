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
	ProcessedCount() int32
	QueueLength() int
}

// Configuration for an Executor
// there are required field to be specified
type ExecutorConfig struct {
	QueueSize           int `required:"true"`
	PoolSize            int `required:"true"`
	OnAllWorkersStopped func()
}

type executorServiceStatus int32

const stopped executorServiceStatus = 0
const running executorServiceStatus = 1

type executorService struct {
	status         executorServiceStatus
	queue          chan func()
	reentrantLock  sync.Mutex
	cfg            ExecutorConfig
	totalProcessed int32
}

// Creates "Executor" instance with a specified configuration
// There are mandatory fields in `ExecutorConfig`
func CreateExecutor(cfg ExecutorConfig) (Executor, error) {
	if cfg.PoolSize < 1 {
		return nil, errors.New("pool size cannot be less than 1")
	}
	if cfg.QueueSize < 1 {
		return nil, errors.New("queue size cannot be less that 1")
	}
	if cfg.OnAllWorkersStopped == nil {
		cfg.OnAllWorkersStopped = func() {}
	}

	var queue = make(chan func(), cfg.QueueSize)
	es := &executorService{status: running, queue: queue, cfg: cfg, totalProcessed: 0}

	go initializeWorkers(es)

	return es, nil
}

func initializeWorkers(es *executorService) {
	var workersWg sync.WaitGroup
	workersWg.Add(int(es.cfg.PoolSize))

	for i := 0; i < int(es.cfg.PoolSize); i++ {
		go worker(i, es, &workersWg)
	}

	// wait when all workers are done
	workersWg.Wait()

	// then close a queue's channel
	if es.queue != nil {
		log.Println("closing channel")
		close(es.queue)
		es.queue = nil
		es.cfg.OnAllWorkersStopped()
	}
}

func worker(index int, es *executorService, workersWg *sync.WaitGroup) {
	defer workersWg.Done()
	log.Println("gcr[", index, "] ready")

	for {
		status := executorServiceStatus(atomic.LoadInt32((*int32)(&es.status)))
		if status == stopped {
			log.Println("gcr[", index, "] service has been stopped, terminating a worker")
			break
		}

		procedure := <-es.queue

		if procedure != nil {
			// invoke a scheduled procedure obtained from a queue
			procedure()
			atomic.AddInt32((*int32)(&es.totalProcessed), 1)
		} else {
			log.Println("gcr[", index, "] Received nil value, terminating a worker")
		}
	}
	log.Println("gcr[", index, "] worker exited")
}

// Add a new function, to be executed, into a queue
// Return an error when it is not possible to schedule:
// - service is stopped
func (es *executorService) Schedule(fn func()) error {
	if fn == nil {
		return errors.New("function cannot be nil")
	}

	es.reentrantLock.Lock()
	defer es.reentrantLock.Unlock()

	if es.status == stopped {
		return errors.New("Executor pool is stopped, cannot accept a job")
	}

	es.queue <- fn

	return nil
}

// Stops executor service
// by sending a stop signal to each worker and changing an executor's status
func (es *executorService) Stop() error {
	es.reentrantLock.Lock()
	defer es.reentrantLock.Unlock()

	log.Println("stopping an executor")

	if es.status == running {
		atomic.StoreInt32((*int32)(&es.status), int32(stopped))

		// send to all workers nil reference to exit
		for i := 0; i < int(es.cfg.PoolSize); i++ {
			es.queue <- nil
		}
	} else {
		return errors.New("service already stopped")
	}

	return nil
}

func (es *executorService) ProcessedCount() int32 {
	return es.totalProcessed
}

func (es *executorService) QueueLength() int {
	return len(es.queue)
}
