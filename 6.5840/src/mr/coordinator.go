package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

//need to keep track of tasks and certain info on them like start time for timeout
type MapTask struct {
	workerNum string //mapnumber, good for 
	timeStart time.Time
	status string //can be either in-progress, idle, or done
	inputFile string
}

type ReduceTask struct {
	workerNum string //reducenumber to know what file to look out for in reducer
	timeStart time.Time
	status string //can be either in-progress, idle,

}
type Coordinator struct {
	// Your definitions here.
	mapTasks []MapTask
	reduceTasks []ReduceTask
	inputFiles []string
	n int
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

	return nil
}



func (c *Coordinator) ConsistentCheck() error {
	//check for tasks that take longer than 10 seconds and put them to idle if its done
	for _, v := range c.mapTasks {
		if time.Since(v.timeStart) > time.Second * 10 {

		}
	}
	for _, v := range c.reduceTasks {
		if time.Since(v.timeStart) > time.Second * 10 {

		}
	}
	return nil
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
	count := 0
	for _, v := range c.reduceTasks {
		if v.status == "done" {
			count += 1
		}
	}
	if count == c.n {
		ret = true
	}
	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}
	c.inputFiles = files
	c.n = nReduce

	go func() {
		for {
			c.ConsistentCheck()
			time.Sleep(2*time.Second)
		}
	}()
	c.server()
	
	return &c
}
