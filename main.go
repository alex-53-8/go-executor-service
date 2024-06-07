package main

import (
	"log"
	"time"

	executor "github.com/alex-53-8/go-executor-service/executor"
)

func main() {
	executor := executor.CreateExecutor()

	for i := 0; i < 10; i++ {
		err := executor.Schedule(func() {
			log.Printf("I'm your lambda function %d\n", i)
		})

		if err != nil {
			log.Println("error: ", err)
		}
	}

	time.Sleep(time.Duration(100) * time.Microsecond)

	executor.Stop()

	err := executor.Schedule(func() {
		log.Println("I'm your lambda function after an executor was stopped")
	})
	log.Println("error: ", err)

	time.Sleep(time.Duration(2) * time.Second)

	log.Println("exiting")
}
