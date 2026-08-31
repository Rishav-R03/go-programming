package ratelimiter

import "time"

type Policy struct {
	Limit  int
	Window time.Duration
}

var DefaultPolicy = Policy{
	Limit:  100,
	Window: 60 * time.Second,
}

type PolicyStore struct {
	overrides map[string]Policy
}

func NewPolicyStore() *PolicyStore {
	return &PolicyStore{overrides: make(map[string]Policy)}
}

func (s *PolicyStore) Set(clientID string, p Policy) {
	s.overrides[clientID] = p
}

func (s *PolicyStore) Get(clientID string) Policy {
	if p, ok := s.overrides[clientID]; ok {
		return p
	}
	return DefaultPolicy
}
