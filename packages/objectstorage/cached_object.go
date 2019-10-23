package objectstorage

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/iotaledger/goshimmer/packages/typeutils"
)

type CachedObject struct {
	objectStorage *ObjectStorage
	value         StorableObject
	err           error
	consumers     int32
	published     int32
	persisted     int32
	deleted       int32
	wg            sync.WaitGroup
	valueMutex    sync.RWMutex
}

func newCachedObject(database *ObjectStorage) (result *CachedObject) {
	result = &CachedObject{
		objectStorage: database,
	}

	result.wg.Add(1)

	return
}

func (cachedObject *CachedObject) Get() (result StorableObject) {
	if !cachedObject.isDeleted() {
		cachedObject.valueMutex.RLock()
		result = cachedObject.value
		cachedObject.valueMutex.RUnlock()
	}

	return
}

func (cachedObject *CachedObject) Release() {
	if consumers := atomic.AddInt32(&(cachedObject.consumers), -1); consumers == 0 {
		if cachedObject.objectStorage.options.cacheTime != 0 {
			time.AfterFunc(cachedObject.objectStorage.options.cacheTime, cachedObject.release)
		} else {
			cachedObject.release()
		}
	}
}

func (cachedObject *CachedObject) Consume(consumer func(object StorableObject)) {
	if cachedObject.isDeleted() {
		consumer(nil)
	} else {
		consumer(cachedObject.Get())
	}

	cachedObject.Release()
}

func (cachedObject *CachedObject) Delete() {
	cachedObject.setDeleted(true)
}

func (cachedObject *CachedObject) RegisterConsumer() {
	atomic.AddInt32(&(cachedObject.consumers), 1)
}

func (cachedObject *CachedObject) Exists() bool {
	return cachedObject.Get() != nil
}

func (cachedObject *CachedObject) setDeleted(deleted bool) {
	if deleted {
		atomic.StoreInt32(&(cachedObject.deleted), 1)
	} else {
		atomic.StoreInt32(&(cachedObject.deleted), 0)
	}
}

func (cachedObject *CachedObject) isDeleted() bool {
	return atomic.LoadInt32(&(cachedObject.deleted)) == 1
}

func (cachedObject *CachedObject) release() {
	cachedObject.objectStorage.cacheMutex.Lock()
	if consumers := atomic.LoadInt32(&(cachedObject.consumers)); consumers == 0 && atomic.AddInt32(&(cachedObject.persisted), 1) == 1 {
		if cachedObject.isDeleted() {
			if err := cachedObject.objectStorage.deleteObjectFromBadger(cachedObject.value.GetId()); err != nil {
				panic(err)
			}
		} else {
			if err := cachedObject.objectStorage.persistObjectToBadger(cachedObject.value.GetId(), cachedObject.value); err != nil {
				panic(err)
			}
		}

		delete(cachedObject.objectStorage.cachedObjects, typeutils.BytesToString(cachedObject.value.GetId()))
	} else if consumers < 0 {
		panic("too many unregistered consumers of cached object")
	}
	cachedObject.objectStorage.cacheMutex.Unlock()
}

func (cachedObject *CachedObject) updateValue(value StorableObject) {
	cachedObject.valueMutex.Lock()
	cachedObject.value = value
	cachedObject.valueMutex.Unlock()
}

func (cachedObject *CachedObject) publishResult(result StorableObject, err error) bool {
	if atomic.AddInt32(&(cachedObject.published), 1) == 1 {
		cachedObject.value = result
		cachedObject.err = err
		cachedObject.wg.Done()

		return true
	}

	return false
}

func (cachedObject *CachedObject) waitForResult() (*CachedObject, error) {
	if atomic.LoadInt32(&(cachedObject.published)) != 1 {
		cachedObject.wg.Wait()
	}

	if err := cachedObject.err; err != nil {
		return nil, err
	} else {
		return cachedObject, nil
	}
}
