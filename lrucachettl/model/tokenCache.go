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

func (tc *TokenCache) ReadMap(userString string) (*cacheItem, bool) {
	// cacheItem := tc.items[userString]
	return &cacheItem{}, false
}
