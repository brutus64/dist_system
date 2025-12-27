package mr

import (
	// "encoding/json"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"net/rpc"
	"os"
	"sort"
	"strconv"
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

func runMap(mapf func(string, string) []KeyValue, content string, reply AssignTaskReply) error {
	/*
	check N reduce tasks, 1 line per reduce task so mr-out-X-Y for x'th map task y'th reduce task
	have keys and values in the form of: "%v %v" format key,value e.g. fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)
	*/
	kvs := mapf(reply.InputFile, content)
	//need to read all the keys, perform ihash, put it in the right bucket for file, with the index being the R'th reduce task
	buckets := make([][]KeyValue, reply.N)
	for _, kv := range kvs {
		ind := ihash(kv.Key) % reply.N
		buckets[ind] = append(buckets[ind], kv)
	}
	//for each bucket create them then add the stuff in json to the bucket
	for i := range buckets{
		file, err := os.CreateTemp(".","map-temp") //curr dir
		if err != nil {
			fmt.Println("error creating temp file for maps: ", err)
			return err
		}
		enc := json.NewEncoder(file)
		for _, kv := range buckets[i] {
			if err := enc.Encode(&kv); err != nil {
				fmt.Println("error occurred while encoding key/val to map, removing temp file as well: ", err)
				os.Remove(file.Name())
				return err
			}
		}
		//done encoding everything so time to atomically rename it
		filename := "mr-" + strconv.Itoa(reply.MapTaskNum) + "-" + strconv.Itoa(i)
		file.Close()
		os.Rename(file.Name(), filename)
	}
	return nil
}

func runReduce(reduce func(string, []string) string, reply AssignTaskReply) error {
	//reduce takes key for 1st param and list of values for 2nd param
	n := reply.ReduceTaskNum
	// fmt.Printf("BEFORE RUNNING LOOP, n: %v, totalmaptasks:, %v", n, reply.TotalMapTasks)
	kvs := make(map[string][]string) //declare a map with key=string, val=list of string
	//start reading through all map tasks for this reducer
	for i := 0; i < reply.TotalMapTasks; i++ {
		filename := "mr-" + strconv.Itoa(i) + "-" + strconv.Itoa(n)
		// fmt.Printf("filename from map reading: %v",filename)
		file, err := os.Open(filename)
		if err != nil {
			continue //some maptasks might not have outputs for a reducetask
		}
		//kv := []string
		//need a map and a list of values for each map, array of maps(keystring,arr(valstring))
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil { //no more to decode
				break
			}
			kvs[kv.Key] = append(kvs[kv.Key], kv.Value)
		}
		file.Close()
	}
	//now sort the keys for this specific reducer
	// fmt.Printf("the map: %v", kvs)
	var keys []string
	for k := range kvs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	//now all keys are made, we can pass to reduce
	file, err := os.CreateTemp(".","reduce-temp")
	if err != nil {
		log.Fatalf("cannot read %v", err)
	}
	for _, k := range keys {
		res := reduce(k, kvs[k])
		fmt.Fprintf(file, "%v %v\n", k, res)
	}
	filename := "mr-out-" + strconv.Itoa(n)
	file.Close()
	os.Rename(file.Name(), filename)
	return nil
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
						fmt.Printf("error reading file %v, error: %v", reply.InputFile, err)
					}
					readable := string(content)
					err = runMap(mapf, readable, reply)
					if err != nil {
						//should we send a rpc back saying task failed?
					}
					//send back rpc that its done, and have it store the intermediate file in the proper reduce task
					doneArgs := FinishTaskArgs{true, reply.MapTaskNum, -1}
					doneReply := FinishTaskReply{}
					call("Coordinator.FinishTask", &doneArgs, &doneReply)
				} else {
					//for reduce need to read the intermediate files,
					runReduce(reduce, reply)
					doneArgs := FinishTaskArgs{false, -1, reply.ReduceTaskNum}
					doneReply := FinishTaskReply{}
					call("Coordinator.FinishTask", &doneArgs, &doneReply)
				}
				
			} else { //no task just sleep
				time.Sleep(time.Second * 2)
			}
		} else { //call to coordinator doesn't work
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
