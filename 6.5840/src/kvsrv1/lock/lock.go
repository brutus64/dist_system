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
			if err == rpc.OK {
				return
			}
			continue //put failed, have it get the value/ver of lock again
		}
		switch val {
		case "":
			err = lk.ck.Put(lk.lockKey, lk.clientID, ver)
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
	for err != rpc.OK {
		err = lk.ck.Put(lk.lockKey, "", ver)
	}
}
