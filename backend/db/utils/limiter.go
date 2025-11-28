package utils

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	MAX_FILES     = 5
	WINDOW_LENGTH = 5 * time.Second
)

type RateLimiter struct {
	mu  sync.Mutex
	mem map[uuid.UUID][]time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		mem: make(map[uuid.UUID][]time.Time),
	}
}

func (r *RateLimiter) Allow(userId uuid.UUID) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	times := r.mem[userId]
	valid := times[:0]

	for _, t := range times {
		if now.Sub(t) <= WINDOW_LENGTH {
			valid = append(valid, t)
		}
	}
	r.mem[userId] = valid

	if len(valid) >= MAX_FILES {
		return false
	}

	r.mem[userId] = append(r.mem[userId], now)

	return true
}
