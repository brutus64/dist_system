questions from lec 3:

1) what does it mean when a machine level instruction is called transparent as opposed to application level operations?
```Transparent = application/disk/OS don't know replication is happening, its all controlled by the hypervisor capturing all these events and sending it through the logging channel to the hypervisor of the backup

you can also run any OS/app and it'll be replicated automatically and work fine
```

2) pros/cons of machine level instruction and application level operation being the 2 diff types of RSM approaches

```
Machine Level 
Pros:
1) can run any OS/app
2) a lot of control
3) failover correctness is usually super high, since it's instruction level correctness (memory/CPU state correct)
Cons:
1) bandwidth can be high because of logging channel
2) determinism is hard when you have a lot of non-deterministic instructions/events
3) performance is slower because of ACK requirement from backup (Output Rule)
4) hypervisor complexity is high

Application Level
Pros:
1) Easy determinism, app just defines the deterministic state machine
2) Low bandwidth since less things to log for operations (high level operations)
3) Fast performance since operations are batched asynchronously usually
Cons:
1) Need to reconstruct states via logs and can lose recent operations
2) Need to write application level code to work with FT
3) Developer complexity high
```

3) pros/cons of RSM vs state transfer (i know about the gigabyte data transfer/result change but is there more?)
```
RSM
Pros:
1) just transfers operation so not much data
2) perfect failover
3) works for continuous service
Cons:
1) hard requirement of deterministic operations
2) real time latency due to syncing logs + ACK

State Transfer
Pros:
1) Simpler to implement usually
2) asynchronous (no stalling for ACKs btwn primary & backup)
Cons:
1) failover can lose recent state since last checkpoint
2) may need to transfer huge Results (GB) -> can be large pauses & bandwidth spikes
3) cant maintain exact execution
```

4) why is hypervisor/virtualization so much easier to do RSM on than just hardware?
```
Hypervisor sees all nondeterministic inputs and can record them, but on hardware you can't just modify physical CPU to be able to intercept and grab these inputs
```
5) test and set when does it actually reset to 0?
```
returns 0 to server if can be primary and increments, else if already 1 means theres a primary.
It resets to 0 when software writes it back to 0 (when the old primary can no longer access the shared disk aynmore)
```
6) only nondeterministic instructions go over logging channel? (for linux, but what about app runs? client input is nondeterministic right? so are we just saying like a deterministic app code is just not transmitting the info of that run over the logging channel?)
```
app code is deterministic (usually) so no need to log per instruction, while things like client input, interrupts, network packets, disk access, are all usually non-deterministic, those are necessary to put through the logging channel to sync backup with primary because these are the points of divergence, no need to put anything non-diverging over the logging channel.
Idea is hypervisor just cares about events that has the primary having a choice or if it reads from the external world. (cpu nondeterministic instructions, interrupt timing, IO, multithread)
```
7) application level usually better than machine level due to bandwidth issues from machine level cause it requires ack’s so it slows down?
```
yes, latency increases and throughput lowers because of the ACK and logging channel sending, which ruins network bandwidth.

On an application level, only logical operations are replicated, and those operations can often be batched and asynchronously sent, so it has a lot less bandwidth overhead, but comes at the cost of having to write distributed-system code for application
```