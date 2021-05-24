package main

import (
	"fmt"
	"time"
)

// this function simulates the behavior of an actor sending
// communications over a channel.
// this actor is prioritized when writing output to the console
func highPriorityAgent(content chan string, done chan bool) {
	content <- "hp: a"
	content <- "hp: b"
	content <- "hp: c"
	time.Sleep(2 * time.Second)
	content <- "hp: d"
	content <- "hp: e"
	done <- true
}

// this function simulates the behavior of an actor sending
// communications over a channel.
// this actor is prioritized below highPriorityAgent
func lowPriorityAgent(content chan string, done chan bool) {
	content <- "lp: a"
	time.Sleep(1 * time.Second)
	content <- "lp: b"
	content <- "lp: c"
	content <- "lp: d"
	done <- true
}

func main() {
	// establish 'communication', done, and exit channels
	comm1 := make(chan string)
	comm2 := make(chan string)
	done := make(chan bool)
	exit := make(chan bool)

	go highPriorityAgent(comm1, done)
	go lowPriorityAgent(comm2, done)

	// we wait for both actors to have sent all content
	// before exiting the process
	go func() {
		<-done
		<-done
		exit <- true
	}()

	// establish a listener
	for {
		// recieve from the channels each half-second.
		// this minimizes printing of the "waiting" line
		time.Sleep(500 * time.Millisecond)

		select {
		// prioritize communication over comm1
		case m1 := <-comm1:
			fmt.Println(m1)
			continue
		default:
		}
		// if no values to recieve from the priorizied channel
		select {
		case m2 := <-comm2:
			fmt.Println(m2)
			continue
		case <-exit:
			fmt.Println("All communications recieved. Exiting.")
		default:
			fmt.Println("No communications currently available. Waiting.")
			continue
		}
		// only reach break on case <- exit.
		break
	}

}
