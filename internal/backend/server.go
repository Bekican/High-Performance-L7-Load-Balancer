package backend

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL          string
	Alive        bool
	mux          sync.RWMutex
	RequestCount uint64
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	alive := b.Alive
	b.mux.RUnlock()
	return alive
}

func (b *Backend) IncrementRequest() {
	atomic.AddUint64(&b.RequestCount, 1)
}

func (b *Backend) GetRequestCount() uint64 {
	return atomic.LoadUint64(&b.RequestCount)
}

func (b *Backend) HealthCheck() bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", b.URL, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
