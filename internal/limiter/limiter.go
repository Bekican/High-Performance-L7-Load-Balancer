package limiter

import (
	"sync"
	"sync/atomic"
	"time"
)

type Bucket struct {
	tokens     float64
	lastRefill time.Time
}

type RateLimiter struct {
	buckets      map[string]*Bucket
	mu           sync.Mutex
	rate         float64
	capacity     float64
	BlockedCount uint64
}

func NewLimiter(rate float64, capacity float64) *RateLimiter {
	return &RateLimiter{
		buckets:  make(map[string]*Bucket),
		rate:     rate,
		capacity: capacity,
	}
}

func (l *RateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, exists := l.buckets[ip]
	if !exists {
		b = &Bucket{
			tokens:     l.capacity,
			lastRefill: time.Now(),
		}
		l.buckets[ip] = b
	}
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	tokensToAdd := elapsed * l.rate

	if tokensToAdd > 0 {
		b.tokens = b.tokens + tokensToAdd
		if b.tokens > l.capacity {
			b.tokens = l.capacity
		}
		b.lastRefill = now
	}
	if b.tokens >= 1.0 {
		b.tokens--
		return true
	}
	atomic.AddUint64(&l.BlockedCount, 1)
	return false
}

func (l *RateLimiter) GetBlockedCount() uint64 {
	return atomic.LoadUint64(&l.BlockedCount)
}
