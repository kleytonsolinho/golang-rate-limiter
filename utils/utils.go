package utils

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Estrutura para armazenar informações de bloqueio
type RateLimiter struct {
	mu            sync.Mutex
	blockedIPs    map[string]time.Time
	blockedTokens map[string]time.Time
	blockDuration time.Duration
}

func NewRateLimiter(blockDuration time.Duration) *RateLimiter {
	return &RateLimiter{
		blockedIPs:    make(map[string]time.Time),
		blockedTokens: make(map[string]time.Time),
		blockDuration: blockDuration,
	}
}

func (rl *RateLimiter) BlockIP(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.blockedIPs[ip] = time.Now().Add(rl.blockDuration)
}

func (rl *RateLimiter) BlockToken(token string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.blockedTokens[token] = time.Now().Add(rl.blockDuration)
}

func (rl *RateLimiter) IsBlocked(ip, token string) (bool, time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Verificar bloqueio por IP
	if unblockTime, blocked := rl.blockedIPs[ip]; blocked {
		if now.Before(unblockTime) {
			return true, unblockTime
		}
		delete(rl.blockedIPs, ip)
	}

	// Verificar bloqueio por Token
	if unblockTime, blocked := rl.blockedTokens[token]; blocked && token != "" {
		if now.Before(unblockTime) {
			return true, unblockTime
		}
		delete(rl.blockedTokens, token)
	}

	return false, time.Time{}
}

func GetIP(r *http.Request) string {
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip, _, _ = net.SplitHostPort(ip)
	}
	return ip
}
