package executor

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
)

type Executor interface {
	Schedule(fn func()) error
	Stop()
}

type executorServiceStatus int32

const stopped executorServiceStatus = 0
const running executorServiceStatus = 1

type executorService struct {
	status executorServiceStatus
	queue  chan func()
	mu     sync.Mutex
}

func CreateExecutor() Executor {
	var queue = make(chan func(), 100)
	es := &executorService{status: running, queue: queue}

	go infinitChannelReader(es)
	return es
}

func infinitChannelReader(es *executorService) {
	for {
		if es.status == stopped {
			log.Println("service has been stopped, terminating a listener")
			break
		}

		nfn := <-es.queue

		if nfn != nil {
			go nfn()
		} else {
			log.Println("Received nil value, terminating a listener")
		}
	}

	log.Println("closing channel")

	if es.queue != nil {
		close(es.queue)
	}
}

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

func (es *executorService) Stop() {
	es.mu.Lock()
	defer es.mu.Unlock()

	if es.status == running {
		atomic.StoreInt32((*int32)(&es.status), int32(stopped))
		es.queue <- nil
	}
}
