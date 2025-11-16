package utils

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu         sync.Mutex
	tokens     map[string]int
	capacity   int
	interval   time.Duration
	lastRefill map[string]time.Time
}

func NewRateLimiter(capacity int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     make(map[string]int),
		capacity:   capacity,
		interval:   interval,
		lastRefill: make(map[string]time.Time),
	}
}

func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	last, ok := r.lastRefill[key]
	if !ok || now.Sub(last) >= r.interval {
		r.tokens[key] = r.capacity
		r.lastRefill[key] = now
	}

	if r.tokens[key] > 0 {
		r.tokens[key]--
		return true
	}
	return false
}
