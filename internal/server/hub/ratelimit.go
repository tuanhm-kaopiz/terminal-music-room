package hub

import (
	"sync"
	"time"
)

// LimitKind selects which token bucket to use.
type LimitKind int

const (
	LimitConnect LimitKind = iota
	LimitCreateRoom
	LimitChat
)

// RateLimitConfig defines token-bucket rates per key (usually client IP).
type RateLimitConfig struct {
	ConnectPerMinute  float64
	CreateRoomPerHour float64
	ChatPerMinute     float64
}

// DefaultRateLimitConfig matches architecture security notes.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		ConnectPerMinute:  10,
		CreateRoomPerHour: 5,
		ChatPerMinute:     20,
	}
}

// Limiter applies per-IP token buckets.
type Limiter struct {
	mu   sync.Mutex
	cfg  RateLimitConfig
	keys map[LimitKind]map[string]*tokenBucket
}

type tokenBucket struct {
	tokens  float64
	updated time.Time
	rate    float64
	burst   float64
}

// NewLimiter creates a limiter with the given config.
func NewLimiter(cfg RateLimitConfig) *Limiter {
	return &Limiter{
		cfg: cfg,
		keys: map[LimitKind]map[string]*tokenBucket{
			LimitConnect:    {},
			LimitCreateRoom: {},
			LimitChat:       {},
		},
	}
}

// Allow reports whether the action is permitted for key.
func (l *Limiter) Allow(kind LimitKind, key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	rate, burst := l.params(kind)
	buckets := l.keys[kind]
	b, ok := buckets[key]
	now := time.Now()
	if !ok {
		b = &tokenBucket{tokens: burst, updated: now, rate: rate, burst: burst}
		buckets[key] = b
	}
	elapsed := now.Sub(b.updated).Seconds()
	b.tokens += elapsed * b.rate
	if b.tokens > b.burst {
		b.tokens = b.burst
	}
	b.updated = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// RetryAfterSeconds estimates seconds until the next token is available.
func (l *Limiter) RetryAfterSeconds(kind LimitKind, key string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	buckets := l.keys[kind]
	b, ok := buckets[key]
	if !ok {
		return 1
	}
	if b.tokens >= 1 {
		return 1
	}
	need := 1 - b.tokens
	if b.rate <= 0 {
		return 60
	}
	sec := int(need / b.rate)
	if sec < 1 {
		return 1
	}
	return sec
}

func (l *Limiter) params(kind LimitKind) (rate float64, burst float64) {
	switch kind {
	case LimitConnect:
		return l.cfg.ConnectPerMinute / 60, l.cfg.ConnectPerMinute
	case LimitCreateRoom:
		return l.cfg.CreateRoomPerHour / 3600, l.cfg.CreateRoomPerHour
	case LimitChat:
		return l.cfg.ChatPerMinute / 60, l.cfg.ChatPerMinute
	default:
		return 1, 1
	}
}

// AllowCreateRoom is used by room handlers (T-006+).
func (s *Server) AllowCreateRoom(ip string) bool {
	return s.limiter.Allow(LimitCreateRoom, ip)
}

// AllowChat is used by chat handlers (T-010+).
func (s *Server) AllowChat(ip string) bool {
	return s.limiter.Allow(LimitChat, ip)
}
