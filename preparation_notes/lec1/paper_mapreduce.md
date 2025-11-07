# Notes for MapReduce Paper: https://pdos.csail.mit.edu/6.824/papers/mapreduce.pdf

## Implementation
Map: accepts a key and value and emits a bunch of (key: val)

Input: Key (document name), Value (document Content)

Map(k1, v1) -> list(k2,v2) <br>Reduce (k2, list(v2)) -> list(v2)

Output: Key2, List of Values

idea is that the map emits a bunch of things associated to a k2 key, and reduce handles all values v2 for a specific k2 emitted by the distributed map function e.g. counting url access freq, reversing web-link graph, distributed grep, term-vector per host, inverted index, distributed sort

![Image of MapReduce implementation](../assets/mapreduce.png)

### General Idea:
Automatic Partition Input Data to M splits -> Machine parallel process of input runs Map() -> produces Intermediate <Key,Value> -> Reduce Invocation -> Partition Intermediate Key By Hash Function (so not 1 machine handles all keys, split keys evenly) -> Worker from Reduce grab data from workers with data from map -> Reduce Done

(1) MapReduce Library splits input file to multiple small chunks (configurable by opt param), then starts copies of program on cluster of machines (workers)

(2) 1 Node is the MASTER node, that assigns M map tasks and R reduce tasks to worker nodes

(3) Worker reads input file, parses key/value pairs out of input file, passes the intermediate pairs to Map(), buffered pairs into memory

(4) Pairs in memory written to local disk periodically, partitioned to R regions with some partition function, the location of the pairs on disk are passed back to master node (master node will forward locations to reduce workers)<br>Note: Each worker node has R regions for R reducers even if it's input data ends up not processing any keys to be processed by another reducer, so it could have empty file for certain keys/partition

(5) Master tells reducer about location of intermediate nodes, RPC to read buffered data on local disk from Map worker nodes, then sorts all data by intermediate keys (MULTIPLE KEYS handled by same reduce worker) <br>Note: Sorting can be externally done if data is too much to be put in memory to be sorted

(6) Reduce reads through all intermediate data, for each unique key and list of values with it (k2, list(v2)) -> Reduce() -> append to output file 

(7) Finish all map/reduce tasks -> master wakes up user program, and proceeds with user program (MapReduce() returned)

Output: R output files (one per reduce task, some filename specified by user program)<br>Note: Output files don't really need to be combined, usually these are passed as input for another MapReduce call or another distributed app can handle inputs that are partitioned


### Master Node:
- holds state (idle, in-progress, complete) on all map/reduce task
- holds identity of worker machines (only care about non-idle tasks)
- holds location & size of R intermediate file regions made by map tasks
- gets update from map periodically on file location and size as map is done (so it's not only sent when its done with map)
- info is pushed incrementally to reduce workers with "in-progress" (so reduce workers can start before a map is done)

### Fault Tolerance:
Thousands of machines are being used to process HUGE data, need to be able to handle failures

#### Worker Failure:
- Master pings worker nodes periodically, no response = marked failure
- map tasks complete nodes -> set back to idle, can be used again for smth else
- any map/reduce task failed -> set to idle, to be used again
- completed map tasks re-executed on failure (can tell since output on local disk is inaccessible on failed tasks, we just check the results to know if it failed or not)
- worker A does work -> fails -> worker B replaces A -> all <b>reduce workers</b> notified of re-execution, all data starts being read from worker B now.

#### Master Failure:
- Master writes periodic checkpoints to be recovered if failed, just start a new one read from checkpoint state
- Oftentimes failure is unlikely since its just 1 Master Node, usually just abort MapReduce computation if it failed, then just do a retry on it when it does
- Hard to make fault tolerant since it has state, unlikely for 1 machine to crash but workers obviously 1000 of machines 1 will crash usually

#### Semantic Failure:
- Generally the final output is same as if you ran the program sequentially on 1 machine with no failures
- Done by atomic writes and commits, in-progress map writes to R temp files, when it's done it "commits" it, sends R temp file info to master and master stores it, master ignores any future completion messages for completed task since it's all deterministic, outputs are the same given same input
- Reduce task done -> rename output file to final output file name, if multiple machine does same reduce task -> multiple rename calls for same final output -> rely on atomic rename operation to guarantee final file state just has data made by 1 execution (1 reduce will win over another doesn't matter which since same results)
- Usually deterministic so same results as running sequentially but there are cases where we have non-deterministic map/reduce functions
- Let's say R1, R2 are reducer results. If reran R1, it will always give R1 for non-deterministic, just like sequential. The difference is that if Map returns X for 1 run and Y for another run (re-execute from crash), then if R1 is based on X and R2 is based on Y, then R1 != R2. You have reasonable semantics (combination of both results) <b>consistent per reduce task but not globally consistent across all reducers</b>

#### Locality
- Network is a scarce resource when it comes to thousands of machines
- Input data being stored on local disks helps in not using network bandwidth
- Typically tries to store 3 copies of a chunk of input file on 3 different machines, and the master keeps in mind of where these copies are and tries to allocate worker nodes to work on input files that are on local disk to avoid using network bandwidth
- Worst case, master node tries to get a node near that resource (e.g. on same network switch) to avoid using too much network bandwidth

#### Task Granularity
- Ideally M map tasks and R reduce tasks are larger than # of workers 
- This means smaller map/reduce tasks so better for load balancing, (no worker runs too long for 1 large task, master can just assign another worker quickly if its smaller tasks)
- Easier to recover from failure if task is small, just re-execute on another machine
- Master makes O(M+R) scheduling decision with O(M*R) state in memory (info about map/reduce tasks each piece of state is around 1 byte for a Map/Reduce Pair so it's not that much but it still scales with O(M\*R))
- So too big M and R is bad too, can run out of memory for master node
- Try to base M on a good size for splitting a file (e.g. 64 MB per input)
- Try to base R on a small multiple of # workers since each reducer makes a output file, don't want too many small output files, but also enough that parallelizing is worth it

#### Backup Tasks
- Stragglers/Bottlenecks -> a machine takes unusually long for a Map or Reduce task so MapReduce is not done
- Hard to predict why, can be hardware issue, another program is ran on the node competing for resource, input data is far away, software issue (e.g. cache disabled)
- Fixed by duplicating the in-progress tasks in other nodes for backup, whichever finishes it first is used, makes it a lot faster and fixes <b>tail latency (time spent on very slow tasks)</b> but usually small # of tasks left that need backup or they're done before backup


## Improvements
- <b>Partition Function</b>: We specify # of R tasks so R output files, but Partition functioning should be good on the intermediate key, usually its just ```hash(key) mod R``` which is pretty even load balancing but sometimes you want to group certain keys together so make your own
- <b>Ordering</b>: Each reducer has sorted intermediate key/value before running reduce() so output file is sorted by key as well -> easier lookups, easy to merge/index results, binary search by key easy
- <b>Combiner Function</b>: (only safe for communitative/associative operations) <word, 1> if 1 million times across 100 machines, you send 1 million times to reducer, it's inefficient, can have a combiner function that adds it all up in each map worker node then send it to reduce node, it's like a <i><b> local pre-reducer function</b></i> just that the output is to an intermediate file not a final output file
- <b>Input/Output Types</b>: can have custom reader/writer for map and reduce, doesn't have to be text files
- <b>Side Effect</b>: if you want to write extra files (logs, debug info, stats), you can by writing to intermediate file then atomically renaming it but has to be <b>atomic and idempotent (safe to retry)</b>, no multi-file atomic commits, you are responsible for making multiple related output files consistent (e.g. keep it deterministic to avoid inconsistency when it's retried)
- <b>Skip Bad Records</b>: wrap MapReduce in a signal handler, if get a bad record -> tell master -> if more than 1 -> skip bad records when retried
- <b>Local Execution</b>: Debugging Dist Sys is hard, it has a local version that runs sequentially to find the issue
- <b>Status Info</b>: Master has HTTP status page to tell you job progress and metrics (e.g. #tasks complete/in-progress, bytes of input, intermediate, output, throughput, worker failures, each worker's stderr/stdout)
- <b>Counters</b>: Have map/reduce tasks keep count of certain events, piggyback ride it on when masters pings for healthy check, good for sanity checks from users when it's displayed on Status Page or when it's returned to user program