package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import "os"
import "strconv"

//
// example to show how to declare the arguments
// and reply for an RPC.
//

//when a worker asks for a task, what information should we send coordinator? 
type AssignTaskArgs struct {
	//not much tbh
}

//when a worker gets information from a coordinator what should we receive back? all values in a reply needs to be public
type AssignTaskReply struct {
	MapTask bool //True = map, False = reduce
	WorkerNumber int //For Map/Reduce number
	//for map
	InputFile string
	MapTaskNum int
	N int //for reduce task
	//for reduce
	ReduceTaskNum int //reads from local file
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
