package locallock

import "sync"

type LocalLock struct {
	m sync.Map
}

func (l *LocalLock) Lock(key string) func() {
	muIface, _ := l.m.LoadOrStore(key, &sync.Mutex{})
	mu := muIface.(*sync.Mutex)
	mu.Lock()
	return func() {
		mu.Unlock()
	}
}

var GlobalLock = &LocalLock{}
