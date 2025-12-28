package mr

import (
	// "fmt"
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
	taskNum 	int //mapnumber, good for
	timeStart 	time.Time
	status 		string //can be either working, idle, or done
	inputFile 	string
}

type ReduceTask struct {
	taskNum 	int //reducenumber to know what file to look out for in reducer
	timeStart 	time.Time
	status 		string //can be either working, idle,

}
type Coordinator struct {
	// Your definitions here.
	mapTasks 	[]MapTask
	reduceTasks []ReduceTask
	inputFiles 	[]string
	n 			int
	mu 			sync.Mutex

	//rpc under the hood runs each RPC handler in own goroutine so need to use a mutex
	//need to store all worker tasks (number, when they started, so we know 10 second timer or have it be a timeout in a "go")
	//keep an idea of the files that have assigned tasks


}

//check if maptasks are all done
func (c *Coordinator) IsMapDone() bool {
	inputMap := make(map[string]bool)
	for _, v := range c.inputFiles {
		inputMap[v] = false
	}
	numDone := 0
	for _, v := range c.mapTasks {
		if v.status == "done" && inputMap[v.inputFile] == false {
			inputMap[v.inputFile] = true
			numDone++
		}
	}
	return numDone == len(inputMap) 
}

// Your code here -- RPC handlers for the worker to call.

//assign a maptask or reducetask depending on what phase we're in
func (c *Coordinator) AssignTask(args *AssignTaskArgs, reply *AssignTaskReply) error {
	//to assign a task, need to look for map
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.IsMapDone() { //map task assign
		//assign an input file that no worker is working on or is marked as done.
		for i, v := range c.mapTasks {
			if v.status == "idle" {
				c.mapTasks[i].status = "working"
				reply.IsMap = true
				reply.MapTaskNum = v.taskNum
				reply.InputFile = v.inputFile
				reply.N = c.n
				reply.TaskAvail = true
				//need to also start it now
				c.mapTasks[i].timeStart = time.Now()
				return nil
			}
		}
	} else { //reduce task assign
		for i, v := range c.reduceTasks {
			if v.status == "idle" {
				c.reduceTasks[i].status = "working"
				reply.IsMap = false
				reply.ReduceTaskNum = v.taskNum
				reply.N = c.n
				reply.TotalMapTasks = len(c.mapTasks)
				reply.TaskAvail = true
				c.reduceTasks[i].timeStart = time.Now()
				return nil
			}
		}
	}
	return nil
}

//marks a task as done
func (c *Coordinator) FinishTask(args *FinishTaskArgs, reply *FinishTaskReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if args.IsMap {
		c.mapTasks[args.MapTaskNum].status = "done"
	} else {
		c.reduceTasks[args.ReduceTaskNum].status = "done"
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


//check for tasks that take longer than 10 seconds that's still working and put them to idle
func (c *Coordinator) ConsistentCheck() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, v := range c.mapTasks {
		if time.Since(v.timeStart) > time.Second * 10 && v.status == "working" {
			c.mapTasks[i].status = "idle"
		}
	}
	for i, v := range c.reduceTasks {
		if time.Since(v.timeStart) > time.Second * 10 && v.status == "working" {
			c.reduceTasks[i].status = "idle"
		}
	}
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
	//initializing the maptasks and reducetasks
	for i := 0; i < len(c.inputFiles); i++ {
		c.mapTasks = append(c.mapTasks, MapTask{i, time.Time{}, "idle", c.inputFiles[i]})
	}
	for i := 0; i < c.n; i++ {
		c.reduceTasks = append(c.reduceTasks, ReduceTask{i, time.Time{}, "idle"})
	}
	//need to check for crashes and timeouts
	go func() {
		for {
			c.ConsistentCheck()
			time.Sleep(2*time.Second)
		}
	}()
	c.server()
	
	return &c
}
