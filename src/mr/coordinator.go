package mr

import (
	"log"
	"net"
	"os"
	"net/rpc"
	"net/http"
	"sync"
)

//need to keep track of tasks and certain info on them like start time for timeout
type TaskTrack struct {
	jobType string
	workerNum string //can be mapnumber or reducenumber
	timeStart string
	status string //can be either in-progress, idle, or
	inputFile string
}

type Coordinator struct {
	// Your definitions here.
	tasks []TaskTrack
	inputFiles []string
	mu sync.Mutex

	//rpc under the hood runs each RPC handler in own goroutine so need to use a mutex
	//need to store all worker tasks (number, when they started, so we know 10 second timer or have it be a timeout in a "go")
	//keep an idea of the files that have assigned tasks


}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) AssignTask(args *AssignTaskArgs, reply *AssignTaskReply) error {
	//to assign a task, need to look for map
	c.mu.Lock()
	defer c.mu.Unlock()
	
}

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}


//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	ret := false

	// Your code here.


	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.


	c.server()
	return &c
}
