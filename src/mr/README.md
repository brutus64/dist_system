# Notes on The Lab

https://pdos.csail.mit.edu/6.824/labs/lab-mr.html

Sample run with wc.go sequentially
```$ cd ~/6.5840
$ cd src/main
$ go build -buildmode=plugin ../mrapps/wc.go
$ rm mr-out*
$ go run mrsequential.go wc.so pg*.txt
$ more mr-out-0
```

- need to create coordinator & worker
- 1 coordinator process
- 1+ worker processes (in parallel)
- workers run on a bunch of machines but this lab it's all on a single machine but still <b> workers talk via RPC to coordinator</b>
- <b>workers in a loop asks coordinator for tasks, read tasks input (1+ file) -> execute task -> write task to output (1+ file) then once done, asks for another task</b>
- <b>coordinator needs to be aware of worker's progress (10 second requirement to finish), if not complete give same task to a different worker</b>

# Main files
## mrcoordinator.go
- the main routine for coordinator
## mrworker.go
- main routine for workers

# MR files
## coordinator.go
- implementation for coordinator
## worker.go
- implementation for worker
## rpc.go
- implementation for rpc

## Testing out coordinator with worker
Keep coordinator running with its input files:
```
$ go build -buildmode=plugin ../mrapps/wc.go
$ rm mr-out*
$ go run mrcoordinator.go pg-*.txt
```
Run workers to accept jobs and use wc.so from word count
```
$ go run mrworker.go wc.so
```
Read the results:
```$ cat mr-out-* | sort | more```

## main/test-mr.sh
checks that <b>mrapps/wc.go</b> and <b>mrapps/indexer.go</b> produce correct outputs given ```pg-xxx.txt``` as input. Test also checks  implementation runs Map/Reduce in <b>Parallel</b> and that <b>Recovery</b> works when crashing a worker

Run with:
```
$ cd ~/6.5840/src/main
$ bash test-mr.sh
*** Starting wc test.
```
^ Will hang since coordinator never finishes (with imcomplete implementation)


if we change ``` ret := false ``` to ``` ret := true ``` in ```func Done``` in mr/coordinator.go, the coordinator exits immediately

```$ bash test-mr.sh
*** Starting wc test.
sort: No such file or directory
cmp: EOF on mr-wc-all
--- wc output is not the same as mr-correct-wc.txt
--- wc test: FAIL
$
```
Then tells you the test script expects output mr-out-X, one per reduce task. It will fail given no implementation since no output file is created so tests fail.

When finished results should look like:
```
$ bash test-mr.sh
*** Starting wc test.
--- wc test: PASS
*** Starting indexer test.
--- indexer test: PASS
*** Starting map parallelism test.
--- map parallelism test: PASS
*** Starting reduce parallelism test.
--- reduce parallelism test: PASS
*** Starting job count test.
--- job count test: PASS
*** Starting early exit test.
--- early exit test: PASS
*** Starting crash test.
--- crash test: PASS
*** PASSED ALL TESTS
$
```

You may see some errors from the Go RPC package that look like:

```2019/12/16 13:27:09 rpc.Register: method "Done" has 1 input parameters; needs exactly three```

Ignore these messages; registering the coordinator as an RPC server checks if all its methods are suitable for RPCs (have 3 inputs); we know that Done is not called via RPC.

Additionally, depending on your strategy for terminating worker processes, you may see some errors of the form

```2025/02/11 16:21:32 dialing:dial unix /var/tmp/5840-mr-501: connect: connection refused```

It is fine to see a handful of these messages per test; they arise when the worker is unable to contact the coordinator RPC server after the coordinator has exited.

# Rules
- Map divides Intermediate keys -> buckets for <b>nReduce</b> reduce tasks nReduce = # reduce tasks (arg that main/mrcoordinator.go passes to MakeCoordinator()) so each Map has n files for n Reducers that are created
- <b>(NUMBERING)</b> Worker needs to put output of X'th reduce task in mr-out-X, 1 line per reduce function output in ```"%v %v" format key,value``` e.g. ```fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)```
- can only modify files in ```mr``` repo, others are ok only for testing
- files from ```intermediate Map``` must be in same dir output, so they can be read as input for Reduce task
- mr/coordinator.go needs a ```Done()``` method implemented to return true when when MapReduce is done, this is expected by main/mrcoordinator.go which it will then exit
- When job done, all worker processes need to exit (to do this use return value from ```call()``` if a worker fails to contact coordinator, we can assume coordinator exited since job is done, so workers can then terminate too, <b>can also have a please exit pseudo task that coordinator can give to workers too</b>)

# Hints
1) mr/worker.go -> Worker() send an RPC to coordinator for tasks, modify coordinator to respond with filename input of unstarted map task, then modify worker to read file and call Map function like in mrsequnetial.go
2) Map/Reduce are loaded at runtime with Go ```plugin``` package, from files that end in ```.so```
3) Need to <b>ALWAYS</b> rebuild if change in mr/ directory with ```go build -buildmode=plugin ../mrapps/wc.go``` or something along those lines
4) lab just shares files on same machine no need for global filesystem like GFS (which would be the case if ran on diff machines)
5) intermediate file should be named ```mr-X-Y``` where X=Map task # and Y=Reduce task #
6) worker map task need to store intermediate key/val pairs in file so it can be read during reduce task, use Go's ```encoding/json``` to write key/val in JSON format in open file
```
WRITING
enc := json.NewEncoder(file)
  for _, kv := ... {
    err := enc.Encode(&kv)
```

```
READING
dec := json.NewDecoder(file)
  for {
    var kv KeyValue
    if err := dec.Decode(&kv); err != nil {
      break
    }
    kva = append(kva, kv)
  }
```
7) map should use ```ihash(key)``` in ```worker.go``` to pick reduce task for a given key (which reducer should work on this)
8) use code from mrsequential.go for reading map's input file, sorting intermediate key/value in between map/reduce runs, and storing reduce output in files
9) coordinatory as an rpc server is ```concurrent```, ```LOCK SHARED DATA```
10) use ```go run -race``` for race detection ```test-mr.sh``` has explanation in comments telling how to use it with ```-race``` flag
11) workers need to wait sometimes (reducers can't start until last map is done), so let workers ```periodically ask coordinator for work, sleeping with time.Sleep() or sync.Cond``` go runs the handler for each RPC in its own thread so one handler is waiting shouldn't stop coordinator from processing other RPC's
12) coordinator can't reliably realize a ```crashed``` worker and worker that is ```stalled```, or workers that are ```too slow``` to be useful. All that can be done is have coordinator wait ```10 seconds``` then assume it's dead and give up then reassign task to diff worker, 
13) If doing ```Backup Tasks```, its tested that code doesn't schedule useless tasks when workers execute tasks without crashing. Backup tasks need to only be scheduled after some relatively long period of time ```(10s)```
14) Testing crash recovery: use ```mrapps/crash.go``` plugin, it randomly exists Map/Reduce functions
15) Make sure nobody reads partially written files in case of crashes, MapReduce uses trick of using temp files to write then atomic rename when complete, can use ```ioutil.TempFile (or os.CreateTemp)``` to create temp files then ```os.Rename``` to atomically rename it.
16) ```test-mr.sh``` runs all processes in subdir ```mr-tmp``` so if something goes wrong in intermediate/final output files, look in there, can even modify ```test-mr.sh``` to exit after failing a test so script doesn't continue testing and overwriting output files
17) ```test-mr-many.sh``` runs test-mr.sh many times in a row to see ```low prob bugs```, takes an arg=#times to run test. (Don't run in parallel since coordinator reuses same socket)
18) Go RPC sends only struct fields whose names are ```Capitalized```, sub structs ```must also have capitalized field names```
19) RPC ```call()```, reply struct should contain all ```default values```, so it should look like 
```
reply := SomeType{}
call(..., &reply)
  ```
without setting any fields in reply b4 call, if you pass reply structs with non-default fields, RPC may silently return incorrect values