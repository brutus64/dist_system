package lock

import (
	"6.5840/kvsrv1/rpc"
	kvtest "6.5840/kvtest1"
)

type Lock struct {
	// IKVClerk is a go interface for k/v clerks: the interface hides
	// the specific Clerk type of ck but promises that ck supports
	// Put and Get.  The tester passes the clerk in when calling
	// MakeLock().
	ck kvtest.IKVClerk
	// You may add code here
	lockKey  string
	clientID string
}

// The tester calls MakeLock() and passes in a k/v clerk; your code can
// perform a Put or Get by calling lk.ck.Put() or lk.ck.Get().
//
// Use l as the key to store the "lock state" (you would have to decide
// precisely what the lock state is).
func MakeLock(ck kvtest.IKVClerk, l string) *Lock {
	lk := &Lock{ck: ck, lockKey: l, clientID: kvtest.RandValue(8)}
	// You may add code here
	return lk
}

func (lk *Lock) Acquire() {
	/*
		idea:
		get the key first
		check if that key is empty
		if it is, put into that lock, the clientID
		else key is clientID itself
		we already acquired it just ignore dont do anymore operations
		else key is a clientID not itself
		we cannot do anything to acquire so just do nothing
	*/
	for {
		val, ver, err := lk.ck.Get(lk.lockKey)
		if err == rpc.ErrNoKey {
			err = lk.ck.Put(lk.lockKey, lk.clientID, 0)
			//should not check errMaybe, we need to know 100% that it went through to be confident in acquiring
			if err == rpc.OK {
				return
			}
			continue //put failed, have it get the value/ver of lock again
		}
		switch val {
		case "":
			err = lk.ck.Put(lk.lockKey, lk.clientID, ver)
			//should not check errMaybe, we need to know 100% that it went through to be confident in acquiring
			if err == rpc.OK {
				return
			}
		case lk.clientID: //already have lock
			return
		default:
			continue //keep getting until its not held by another client
		}
	}
}

func (lk *Lock) Release() {
	val, ver, err := lk.ck.Get(lk.lockKey)
	if val != lk.clientID || err == rpc.ErrNoKey { //not held by us
		return
	}
	err = lk.ck.Put(lk.lockKey, "", ver)
	//err returns ErrMaybe or OK only from Put
	//err.OK is for sure done
	//errMaybe is good enough too since we know val == lk.clientID, so only we can release it and if we get ErrMaybe it means we got a ErrVersion after retry, and only we can be the cause of that since only a client can release its own lock
	
	//no longer needed cause Put allows errmaybe and handles retry already in Put of client.go (it was the source of my infinite loop)
	// for err != rpc.OK {
	// 	err = lk.ck.Put(lk.lockKey, "", ver)
	// }
}
