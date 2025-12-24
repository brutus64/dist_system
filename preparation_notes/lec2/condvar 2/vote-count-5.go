package main

import "time"
import "math/rand"

func main() {
	rand.Seed(time.Now().UnixNano())
	//bad code since at count == 5 it will just stop receiving from the channels and the other channels are just blocked sending data, leaking threads (wasting mem, better to just receive all 10)
	count := 0
	finished := 0
	ch := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			ch <- requestVote()
		}()
	}
	for count < 5 && finished < 10 {
		v := <-ch
		if v {
			count += 1
		}
		finished += 1
	}
	if count >= 5 {
		println("received 5+ votes!")

	} else {
		println("lost")
	}
}

func requestVote() bool {
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	return rand.Int()%2 == 0
}
