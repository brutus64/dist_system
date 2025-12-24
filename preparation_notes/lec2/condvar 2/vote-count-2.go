package main

import "sync"
import "time"
import "math/rand"

func main() {
	rand.Seed(time.Now().UnixNano())
	//have to protect every access count and finish
	count := 0
	finished := 0
	var mu sync.Mutex
	//lock when writing to it
	for i := 0; i < 10; i++ {
		go func() {
			vote := requestVote()
			mu.Lock()
			defer mu.Unlock()
			if vote {
				count++
			}
			finished++
		}()
	}
	//lock when reading, takes turn between this and goroutines
	//lowkey annoying its just waiting until count reaches 5, its like spinning on processor doing nothing, can be optimized so processor does something more interesting, like using condition variables
	for {
		mu.Lock()

		if count >= 5 || finished == 10 {
			break
		}
		mu.Unlock()
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
	return rand.Int()%2 == 0
}
