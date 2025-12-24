1) create rpc definitions for assign tasks structs
2) figure out what the coordinator needs to hold in its state
3) have the workers use call() on the coordinator for tasks
4) have coordinators assign tasks
5) have coordinators keep periodic track of the workers (10 second)
6) use the mapf passed into worker where we use ihash and have it run map
7) use the reducef passed into worker where we have it run reduce task