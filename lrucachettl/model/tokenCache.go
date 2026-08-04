package model

import (
	"sync"
	"time"
)

type UserSession struct {
	UserID string
	Role   string
}

type cacheItem struct {
	session   UserSession
	expiresAt time.Time
}

type TokenCache struct {
	items map[string]cacheItem
	Mu    sync.RWMutex
}

func NewTokenCache() *TokenCache {
	return &TokenCache{
		items: make(map[string]cacheItem),
	}
}

// Get retrieves a session from the cache.
// It uses an RLock so multiple concurrent readers do not block each other.
func (c *TokenCache) Get(token string) (UserSession, bool) {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	item, exists := c.items[token]
	if !exists {
		return UserSession{}, false
	}
	if time.Now().After(item.expiresAt) {
		return UserSession{}, false
	}
	return item.session, true
}

// Set inserts or updates a session in the cache with a specified Time-To-Live (TTL).
// It acquires a full write Lock to ensure memory safety during mutations.
func (c *TokenCache) Set(token string, session UserSession, ttl time.Duration) {
	expiresAt := time.Now().Add(ttl)
	c.Mu.Lock()
	defer c.Mu.Unlock()
	c.items[token] = cacheItem{
		session:   session,
		expiresAt: expiresAt,
	}
}

func (c *TokenCache) cleanUpExpired() {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	now := time.Now()
	for token, item := range c.items {
		if now.After(item.expiresAt) {
			delete(c.items, token)
		}
	}
	// Map "Reset" Rule: Go maps grow but never shrink their bucket arrays.
	// If the cache becomes empty, re-initialize the map to free bucket memory.
	if len(c.items) == 0 {
		c.items = make(map[string]cacheItem)
	}
}

// StartCleanup launches a long-running background goroutine that periodically sweeps
// expired tokens.

func (c *TokenCache) StartCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			c.cleanUpExpired()
		}
	}()
}

// helper function to count non-expired keys currently in cache
func (c *TokenCache) Count() int {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	return len(c.items)
}
