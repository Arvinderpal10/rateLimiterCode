package limiter

import (
	"log"
	"sync"
	"time"

	"github.com/arvinderpal10/ratelimiter/internal/config"
)

type Store struct {
	users   map[string]*TokenBucket
	mu      sync.RWMutex
	cfg     config.Config
	cleanup *time.Ticker
	stopCh  chan struct{}
}

func NewStore(cfg config.Config) *Store {
	store := &Store{
		users:   make(map[string]*TokenBucket),
		cfg:     cfg,
		cleanup: time.NewTicker(cfg.CleanupInterval),
		stopCh:  make(chan struct{}),
	}
	go store.cleanupLoop()
	return store
}

func (s *Store) Allow(userID string) bool {
	s.mu.RLock()
	bucket, exists := s.users[userID]
	s.mu.RUnlock()

	if !exists {
		s.mu.Lock()
		bucket, exists = s.users[userID]
		if !exists {
			bucket = newTokenBucket(s.cfg.RateLimitPerMin)
			s.users[userID] = bucket
		}
		s.mu.Unlock()
	}

	return bucket.TryConsume(s.cfg.RateLimitPerMin)
}

func (s *Store) Stats() map[string]float64 {
	s.mu.RLock()
	buckets := make(map[string]*TokenBucket, len(s.users))
	for id, b := range s.users {
		buckets[id] = b
	}
	s.mu.RUnlock()

	stats := make(map[string]float64, len(buckets))
	for id, bucket := range buckets {
		stats[id] = bucket.CurrentTokens(s.cfg.RateLimitPerMin)
	}
	return stats
}

func (s *Store) cleanupLoop() {
	for {
		select {
		case <-s.cleanup.C:
			s.cleanupInactive()
		case <-s.stopCh:
			s.cleanup.Stop()
			return
		}
	}
}

func (s *Store) cleanupInactive() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, bucket := range s.users {
		if now.Sub(bucket.LastAccess()) > s.cfg.UserTTL {
			delete(s.users, id)
			log.Printf("cleaned up inactive user: %s", id)
		}
	}
}

func (s *Store) Shutdown() {
	close(s.stopCh)
}
