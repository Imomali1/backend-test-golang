package ratelimiter

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.Mutex
	maxReqs  int
	window   time.Duration
	requests []time.Time
}

func New(maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		maxReqs:  maxRequests,
		window:   window,
		requests: make([]time.Time, 0, maxRequests),
	}
}

func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	validRequests := make([]time.Time, 0, len(r.requests))
	for _, reqTime := range r.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	r.requests = validRequests

	if len(r.requests) >= r.maxReqs {
		return false
	}

	r.requests = append(r.requests, now)
	return true
}

func (r *RateLimiter) RetryAfter() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.requests) == 0 {
		return 0
	}

	oldest := r.requests[0]
	resetTime := oldest.Add(r.window)

	untilReset := time.Until(resetTime)
	if untilReset < 0 {
		return 0
	}

	return untilReset
}

func (r *RateLimiter) Remaining() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	var count int
	for _, reqTime := range r.requests {
		if reqTime.After(cutoff) {
			count++
		}
	}

	remaining := r.maxReqs - count
	if remaining < 0 {
		return 0
	}

	return remaining
}

func (r *RateLimiter) ForceFill() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	r.requests = make([]time.Time, r.maxReqs)
	for i := 0; i < r.maxReqs; i++ {
		r.requests[i] = now
	}
}

func (r *RateLimiter) BlockUntil(until time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.requests = make([]time.Time, r.maxReqs)
	for i := 0; i < r.maxReqs; i++ {
		r.requests[i] = until.Add(-r.window)
	}
}
