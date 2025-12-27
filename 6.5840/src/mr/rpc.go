package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)
//
// example to show how to declare the arguments
// and reply for an RPC.
//


type AssignTaskArgs struct {
	//not much tbh
}

//when a worker gets information from a coordinator what should we receive back? all values in a reply needs to be public
type AssignTaskReply struct {
	IsMap 			bool //True = map, False = reduce
	// WorkerNumber int //For Map/Reduce number
	InputFile 		string
	MapTaskNum 		int
	N 				int //for reduce task
	ReduceTaskNum 	int //reads from local file
	TaskAvail 		bool //need a way for worker to know if it has a task or not when rpc calls so it knows if it should sleep or work on something
}

type FinishTaskArgs struct {
	IsMap			bool
	MapTaskNum	 	int
	ReduceTaskNum	int
}

type FinishTaskReply struct {
	//not much tbh
}
type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.


// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
