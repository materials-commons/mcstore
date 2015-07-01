package with

//import "sync"
//
//// ReadLock calls fn after acquiring a read lock.
//func ReadLock(mutex *sync.RWMutex, fn func()) {
//	defer mutex.RUnlock()
//	mutex.RLock()
//	fn()
//}
//
//// WriteLock calls fn after acquiring a write lock.
//func WriteLock(mutex *sync.RWMutex, fn func()) {
//	defer mutex.Unlock()
//	mutex.Lock()
//	fn()
//}
//
//type LockerCallFunc func(what interface{})
//
//// An RWLocker is an interface that acquires
//// a lock before calling the given func.
//type RWLocker interface {
//	WithReadLock(parm interface{}, fn LockerCallFunc)
//	WithWriteLock(parm interface{}, fn LockerCallFunc)
//}
//
//// RWLock implements the RWLocker interface. It takes a function
//// that wraps the function to call. This allows the user
//// to perform common book keeping or checks before calling
//// the function passed to WithReadLock or WithWriteLock.
//type RWLock struct {
//	mutex *sync.RWMutex
//	fn    func(parm interface{}, fn LockerCallFunc)
//}
//
//func NewRWLock(fn LockerCallFunc) *RWLock {
//	fnToCall := fn
//	if fnToCall == nil {
//		fnToCall = func(parm interface{}, fun LockerCallFunc) {
//			fun(parm)
//		}
//	}
//	return &RWLock{
//		mutex: &sync.RWMutex{},
//		fn:    fnToCall,
//	}
//}
//
//// WithReadLock acquires a read lock and then calls
//// the RWLock function passing it the fn.
//func (lock *RWLock) WithReadLock(parm interface{}, fn LockerCallFunc) {
//	defer lock.mutex.RUnlock()
//	lock.mutex.RLock()
//	lock.fn(parm, fn)
//}
//
//// WithWriteLock acquires a write lock and then calls
//// the RWLock function passing it the fn.
//func (lock *RWLock) WithWriteLock(parm interface{}, fn LockerCallFunc) {
//	defer lock.mutex.Unlock()
//	lock.mutex.Lock()
//	lock.fn(parm, fn)
//}
//
//// Use uses the passed in fn, rather than the fn that
//// that the RWLock was originally set up with.
//func (lock *RWLock) Use(fn LockerCallFunc) *RWLock {
//	rwLockToUse := &RWLock{
//		mutex: lock.mutex,
//		fn:    fn,
//	}
//	return rwLockToUse
//}
