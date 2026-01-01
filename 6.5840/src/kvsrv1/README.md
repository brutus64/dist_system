# Instructions (all work is in src/ksrv1)

each put must be executed ```at-most-once``` even with network failures

Operations need to be linearizable, will need to implement a lock

## KV server
- client -> Clerk (RPC) -> server [to interact with key/value server]
- Client can do 2 RPC operations
    1) Put(key, value, version)
    2) Get(key)
- Key = string, Value = string, version = # times key has been written
- Put can only replace/install a value for a key if version # == server version # for key (if so when server processes, increment version #)
- Don't match version # -> server returns ```rpc.ErrVersion```
- Client creates a new key through ```Put``` with ```version # = 0``` (server will result in 1)
- Client with version # in Put > 0 and key don't exit returns ```rpc.ErrNoKey```

- ```Get(key)``` fetches ```curr_value, version``` if don't exist return ```rpc.ErrNoKey```

- version # for keys is good for implementing locks with ```Put``` to make sure at-most-once semantic applies for Put when network is unreliable and client retransmits the same value

- Linearizable key/value service from the POV of client calling ```Clerk.Get``` and ```Clerk.Put``` (concurrent operations will have return values and final states that can be explained by a history of sequential operation runs)
- ```Concurrent Operations```: if operations overlap in time
- Linearizability is good cause its a behavior of a single server that processes 1 req at a time

## kvsrv1/client.go
Implements ```Clerk``` that clients use to manage RPC interactions with server has
1) Put
2) Get

## kvsrv1/server.go
Contains server code, with handlers for 
1) Put
2) Get

## kvsrv1/rpc
Package to create RPC requests, replies, errors values seen in ```kvsrv1/rpc/rpc.go```, but don't need to modify them


## Part 1 (easy)
- implement solution that works when there are NO DROPPED MESSAGES
```
Need to complete:
client.go -> Clerk's PUT and GET methods for RPC
server.go -> PUT and GET handlers for RPC

test with:
$ go test -v -run Reliable

use go test -race to help

SAMPLE RESULT:

=== RUN   TestReliablePut
One client and reliable Put (reliable network)...
  ... Passed --   0.0  1     5    0
--- PASS: TestReliablePut (0.00s)
=== RUN   TestPutConcurrentReliable
Test: many clients racing to put values to the same key (reliable network)...
info: linearizability check timed out, assuming history is ok
  ... Passed --   3.1  1 90171 90171
--- PASS: TestPutConcurrentReliable (3.07s)
=== RUN   TestMemPutManyClientsReliable
Test: memory use many put clients (reliable network)...
  ... Passed --   9.2  1 100000    0
--- PASS: TestMemPutManyClientsReliable (16.59s)
PASS
ok  	6.5840/kvsrv1	19.681s


# after each PASSED = real time in seconds, (1 = the number of RPCs sent (including client RPCs) and # of key/value operation executed from Client GET and PUT calls)
```


## Part 2 (moderate)
Implement ```lock layered``` on client's ```Clerk.Put``` and ```Clerk.Get``` calls which supports
1) ```Acquire``` method
2) ```Release``` method

This means ONLY 1 client can get a lock at a time, other clients have to wait until 1st client releases the lock with ```Release```

Code is in ```src/kvsrv1/lock/``` need to modify ```src/kvsrv1/lock/lock.go```, ```Acquire``` and ```Release``` can talk to kv server by calling ```lk.ck.Put()``` and ```lk.ck.Get()```

Usually clients can crash while having a lock, so it'll never get released which means you need a lease (which can expire so lock server will release on behalf of client) but this is ignored for this

```Hint:```
1) need a unique ID for each lock client, ```kvtest.RandValue(8)``` to get a random string 
2) Lock service needs a specific key to store ```lock state``` (need to decide what the lock state is), the key to be used is pass through param ```l``` of ```MakeLock``` in ```src/kvsrv1/lock/lock.go```

```
test with:
$ cd lock
$ go test -v -run Reliable

SAMPLE RESULT:

=== RUN   TestOneClientReliable
Test: 1 lock clients (reliable network)...
  ... Passed --   2.0  1   974    0
--- PASS: TestOneClientReliable (2.01s)
=== RUN   TestManyClientsReliable
Test: 10 lock clients (reliable network)...
  ... Passed --   2.1  1 83194    0
--- PASS: TestManyClientsReliable (2.11s)
PASS
ok  	6.5840/kvsrv1/lock	4.120s
```


## Part 3 (moderate)
```Clerk``` needs implementation for retry if no reply

Challenge: Network can ```re-order, delay, discard RPC reqs/responses```, need to recover from such situation by having Clerk re-trying each RPC until a reply from the server is detected

- Discarding RPC req is easy, client re-sends req -> server receive & execute
    - but if RPC reply msg discarded, client has ```no reply``` -> ```re-send RPC request``` -> ```server has 2 copies of req```

        OK for GET since it doesn't modify state 
        <br><br> OK for PUT RPC with same version # -> server only runs PUT if version # is same, if it has executed already on 1st time, 2nd time respond with ```rpc.ErrVersion``` since version state no longer same 
        <br><br> ```ISSUE THOUGH -> Clerk still don't know if Clerk's PUT was executed or not```
        1) 1st RPC executed -> response droppped -> retry request from Client -> rpc.ErrVersion
        - Result: ```Clerk.Put``` has to return ```rpc.ErrMaybe``` instead of ```rpc.ErrVersion``` cause the request could've been executed, then the app handles the case. 
        2) Another Clerk updated key before Clerk's 1st RPC -> server executes NONE of the 2 Clerk's RPC -> 2 rpc.ErrVersion
        - Result: ```Clerk.Put``` responds with ```rpc.ErrVersion``` to app since we know RPC was definitely not executed by server

        Easier if ```PUT``` was ```exactly-once``` so there's no ErrMaybe, but hard to guarantee without maintaining state at server for EVERY CLERK. In last exercise -> implement lock with Clerk to see how to program with ```at-most-once``` Clerk.Put

Modify ```kvsrv1/client.go``` to continue even when dropped RPC requests/replies  
- return ```true``` from client's ```ck.clnt.Call()``` = client received RPC reply from server
- return ```false``` = not received a reply (Call() waits for reply msg for timeout interval but returned false cause it timed out)

        
Clerk has to keep re-sending RPC until reply is received, solution shouldn't require ANY changes to server.


```
test with:
tests in kvsrv1/

go test -v

SAMPLE RESULT:
=== RUN   TestReliablePut
One client and reliable Put (reliable network)...
  ... Passed --   0.0  1     5    0
--- PASS: TestReliablePut (0.00s)
=== RUN   TestPutConcurrentReliable
Test: many clients racing to put values to the same key (reliable network)...
info: linearizability check timed out, assuming history is ok
  ... Passed --   3.1  1 106647 106647
--- PASS: TestPutConcurrentReliable (3.09s)
=== RUN   TestMemPutManyClientsReliable
Test: memory use many put clients (reliable network)...
  ... Passed --   8.0  1 100000    0
--- PASS: TestMemPutManyClientsReliable (14.61s)
=== RUN   TestUnreliableNet
One client (unreliable network)...
  ... Passed --   7.6  1   251  208
--- PASS: TestUnreliableNet (7.60s)
PASS
ok  	6.5840/kvsrv1	25.319s
```

## Part 4 (easy)
Implement a lock with kv clerk and unreliable network by modifying lock implementation to work correctly with modified key/value client when network is not reliable. Need to pass all ```kvsrv1/lock/``` tests including unreliable ones

```
test with:
$ cd lock
$ go test -v

SAMPLE RESULT:
=== RUN   TestOneClientReliable
Test: 1 lock clients (reliable network)...
  ... Passed --   2.0  1   968    0
--- PASS: TestOneClientReliable (2.01s)
=== RUN   TestManyClientsReliable
Test: 10 lock clients (reliable network)...
  ... Passed --   2.1  1 10789    0
--- PASS: TestManyClientsReliable (2.12s)
=== RUN   TestOneClientUnreliable
Test: 1 lock clients (unreliable network)...
  ... Passed --   2.3  1    70    0
--- PASS: TestOneClientUnreliable (2.27s)
=== RUN   TestManyClientsUnreliable
Test: 10 lock clients (unreliable network)...
  ... Passed --   3.6  1   908    0
--- PASS: TestManyClientsUnreliable (3.62s)
PASS
ok  	6.5840/kvsrv1/lock	10.033s
```