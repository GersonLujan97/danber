package transport

import (
	"sync"
	"time"
)

type bucket struct {
	until time.Time
	used  int
}

// One-process fixed windows, bounded memory, no goroutine or per-request timer.
// A full table fails closed until entries expire; it never evicts active keys.
type limiter struct {
	mu                           sync.Mutex
	entries                      map[string]bucket
	limit, globalLimit, capacity int
	global                       bucket
	now                          func() time.Time
}

func newLimiter(limit, global, capacity int) *limiter {
	return &limiter{entries: make(map[string]bucket), limit: limit, globalLimit: global, capacity: capacity, now: time.Now}
}
func (l *limiter) allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	if !now.Before(l.global.until) {
		l.global = bucket{until: now.Add(time.Minute)}
	}
	if l.global.used >= l.globalLimit {
		return false, l.global.until.Sub(now)
	}
	b, exists := l.entries[key]
	if !exists || !now.Before(b.until) {
		if !exists && len(l.entries) >= l.capacity {
			for k, v := range l.entries {
				if !now.Before(v.until) {
					delete(l.entries, k)
				}
			}
			if len(l.entries) >= l.capacity {
				return false, time.Minute
			}
		}
		b = bucket{until: now.Add(time.Minute)}
	}
	if b.used >= l.limit {
		return false, b.until.Sub(now)
	}
	b.used++
	l.global.used++
	l.entries[key] = b
	return true, 0
}
