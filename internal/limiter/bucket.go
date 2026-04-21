package limiter

import (
	"sync"
	"time"
)

type TokenBucket struct {
	tokens         float64
	lastRefillTime time.Time
	lastAccess     time.Time
	mu             sync.Mutex
}

func newTokenBucket(capacity int) *TokenBucket {
	now := time.Now()
	return &TokenBucket{
		tokens:         float64(capacity),
		lastRefillTime: now,
		lastAccess:     now,
	}
}

func (tb *TokenBucket) TryConsume(ratePerMin int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Minutes()
	tb.tokens += elapsed * float64(ratePerMin)
	if tb.tokens > float64(ratePerMin) {
		tb.tokens = float64(ratePerMin)
	}
	tb.lastRefillTime = now
	tb.lastAccess = now

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}
	return false
}

func (tb *TokenBucket) LastAccess() time.Time {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.lastAccess
}

func (tb *TokenBucket) CurrentTokens(ratePerMin int) float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Minutes()
	tokens := tb.tokens + elapsed*float64(ratePerMin)
	if tokens > float64(ratePerMin) {
		tokens = float64(ratePerMin)
	}
	return tokens
}
