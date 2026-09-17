// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/9/17
// 描述：
// *****************************************************************************

package lock

import "sync"

type Locker struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func NewLocker() *Locker {
	return &Locker{
		locks: make(map[string]*sync.Mutex),
	}
}

// getLock 获取指定区划对应的锁
func (r *Locker) getLock(key string) *sync.Mutex {
	r.mu.Lock()
	defer r.mu.Unlock()

	lock, ok := r.locks[key]
	if !ok {
		lock = &sync.Mutex{}
		r.locks[key] = lock
	}

	return lock
}

// Lock 锁定指定区划
func (r *Locker) Lock(key string) {
	r.getLock(key).Lock()
}

// Unlock 解锁指定区划
func (r *Locker) Unlock(key string) {
	r.getLock(key).Unlock()
}

func (r *Locker) Do(key string, fn func()) {
	lock := r.getLock(key)

	lock.Lock()
	defer lock.Unlock()

	fn()
}
