package core

import (
	"sync"
	"time"
)

type AgentRegistry struct {
	mu      sync.RWMutex
	entries map[string]*registryEntry
}

type registryEntry struct {
	agent    *Agent
	lastUsed time.Time
}

func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		entries: make(map[string]*registryEntry),
	}
}

func (r *AgentRegistry) Get(sid string) (*Agent, bool) {
	r.mu.Lock()
	e, ok := r.entries[sid]
	if ok {
		e.lastUsed = time.Now()
	}
	r.mu.Unlock()
	if !ok {
		return nil, false
	}
	return e.agent, true
}

func (r *AgentRegistry) Set(sid string, a *Agent) {
	r.mu.Lock()
	r.entries[sid] = &registryEntry{agent: a, lastUsed: time.Now()}
	r.mu.Unlock()
}

func (r *AgentRegistry) Delete(sid string) {
	r.mu.Lock()
	delete(r.entries, sid)
	r.mu.Unlock()
}

// EvictExpired removes entries that have not been used within ttl.
func (r *AgentRegistry) EvictExpired(ttl time.Duration) {
	cutoff := time.Now().Add(-ttl)
	r.mu.Lock()
	for sid, e := range r.entries {
		if e.lastUsed.Before(cutoff) {
			delete(r.entries, sid)
		}
	}
	r.mu.Unlock()
}
