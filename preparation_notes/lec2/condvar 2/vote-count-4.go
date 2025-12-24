package main

import "sync"
import "time"
import "math/rand"

func main() {
	rand.Seed(time.Now().UnixNano())
	//what this one does is to have the goroutine actually wake the main thread up rather than have some random timing that might not work
	count := 0
	finished := 0
	var mu sync.Mutex
	cond := sync.NewCond(&mu)

	for i := 0; i < 10; i++ {
		go func() {
			vote := requestVote()
			mu.Lock()
			defer mu.Unlock()
			if vote {
				count++
			}
			finished++
			cond.Broadcast() //wake up all waiting goroutines on cond, signal only wakes up one
		}()
	}

	mu.Lock()
	for count < 5 && finished != 10 {
		cond.Wait() //release the lock so others can update state (it does mu.Unlock i think) then when done it does .Lock() again?
	}
	if count >= 5 {
		println("received 5+ votes!")
	} else {
		println("lost")
	}
	mu.Unlock()
}

func requestVote() bool {
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	return rand.Int() % 2 == 0
}
