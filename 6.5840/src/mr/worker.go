package mr

import (
	// "encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"net/rpc"
	"os"
	"sort"
	"time"
)

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func runMap(mapf func(string, string) []KeyValue, content string, reply AssignTaskReply) {
	/*
	check N reduce tasks, 1 line per reduce task so mr-out-X-Y for x'th map task y'th reduce task
	have keys and values in the form of: "%v %v" format key,value e.g. fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)
	*/
	kvs := mapf(reply.InputFile, content)
	//need to read all the keys, check for the ihash, put it in the right bucket for file
	buckets := make([][]KeyValue, reply.N)
	for _, kv := range kvs {
		ind := ihash(kv.Key) % reply.N
		buckets[ind] = append(buckets[ind], kv)
	}
	for i := 0; i < len(buckets); i++ {
		file, err := os.CreateTemp("",)
		
	}
}

func runReduce(reduce func(string, []string) string) {
	
}
//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue, reduce func(string, []string) string) {

	// Your worker implementation here.
	
	// uncomment to send the Example RPC to the coordinator.
	// CallExample()
	for {
		args := AssignTaskArgs{}
		reply := AssignTaskReply{}
		ok := call("Coordinator.AssignTask", &args, &reply)
		if ok {
			if reply.TaskAvail {
				if reply.IsMap {
					content, err := os.ReadFile(reply.InputFile)
					if err != nil {
						fmt.Println("error reading file %v, error: %v", reply.InputFile, err)
					}
					
					//map function
					//read file before calling runmap
					// dec := json.NewDecoder()
					// for {
						
					// }
					readable = string(content)
					runMap(mapf, readable, reply)
				} else {
					runReduce(reduce)
				}
			} else {
				time.Sleep(time.Second * 2)
			}
		} else {
			time.Sleep(time.Second * 2)
		}
	}
}

//
// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
//
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args any, reply any) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
