package aichat

import (
	"sync"
	"time"

	"personal-page-be/biz/infra/config"
)

type guardDenyReason string

const (
	guardDeniedMinute     guardDenyReason = "minute_rate_limit"
	guardDeniedConcurrent guardDenyReason = "concurrent_limit"
	guardDeniedDaily      guardDenyReason = "daily_budget"
)

type guardDenial struct {
	Reason     guardDenyReason
	RetryAfter time.Duration
}

type guardEntry struct {
	minuteStart time.Time
	minuteCount int
	concurrent  int
	lastSeen    time.Time
}

type requestGuard struct {
	mu          sync.Mutex
	entries     map[string]*guardEntry
	limits      config.AIChatLimits
	now         func() time.Time
	lastCleanup time.Time
}

type guardLease struct {
	guard *requestGuard
	keys  []string
	once  sync.Once
}

func newRequestGuard(limits config.AIChatLimits) *requestGuard {
	return &requestGuard{
		entries: make(map[string]*guardEntry),
		limits:  limits,
		now:     time.Now,
	}
}

func (g *requestGuard) acquire(identityKey string, ipKey string) (*guardLease, *guardDenial) {
	now := g.now()
	keys := distinctGuardKeys(identityKey, ipKey)

	g.mu.Lock()
	defer g.mu.Unlock()
	g.cleanupLocked(now)

	for _, key := range keys {
		entry := g.entryLocked(key, now)
		if entry.concurrent >= g.limits.MaxConcurrent {
			return nil, &guardDenial{Reason: guardDeniedConcurrent, RetryAfter: time.Second}
		}
		if entry.minuteCount >= g.limits.RequestsPerMinute {
			retryAfter := time.Minute - now.Sub(entry.minuteStart)
			if retryAfter < time.Second {
				retryAfter = time.Second
			}
			return nil, &guardDenial{Reason: guardDeniedMinute, RetryAfter: retryAfter}
		}
	}

	for _, key := range keys {
		entry := g.entryLocked(key, now)
		entry.minuteCount++
		entry.concurrent++
		entry.lastSeen = now
	}
	return &guardLease{guard: g, keys: keys}, nil
}

func (g *requestGuard) entryLocked(key string, now time.Time) *guardEntry {
	entry := g.entries[key]
	if entry == nil {
		entry = &guardEntry{minuteStart: now, lastSeen: now}
		g.entries[key] = entry
	}
	if now.Sub(entry.minuteStart) >= time.Minute || now.Before(entry.minuteStart) {
		entry.minuteStart = now
		entry.minuteCount = 0
	}
	return entry
}

func (g *requestGuard) cleanupLocked(now time.Time) {
	if !g.lastCleanup.IsZero() && now.Sub(g.lastCleanup) < time.Hour {
		return
	}
	for key, entry := range g.entries {
		if entry.concurrent == 0 && now.Sub(entry.lastSeen) > 48*time.Hour {
			delete(g.entries, key)
		}
	}
	g.lastCleanup = now
}

func (l *guardLease) release() {
	if l == nil || l.guard == nil {
		return
	}
	l.once.Do(func() {
		now := l.guard.now()
		l.guard.mu.Lock()
		defer l.guard.mu.Unlock()
		for _, key := range l.keys {
			if entry := l.guard.entries[key]; entry != nil {
				if entry.concurrent > 0 {
					entry.concurrent--
				}
				entry.lastSeen = now
			}
		}
	})
}

func distinctGuardKeys(identityKey string, ipKey string) []string {
	keys := make([]string, 0, 2)
	if identityKey != "" {
		keys = append(keys, "identity:"+identityKey)
	}
	if ipKey != "" {
		keys = append(keys, "ip:"+ipKey)
	}
	return keys
}
