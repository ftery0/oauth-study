package handlers

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ipLimiter: IP 별 token bucket. brute force / 토큰 추측 공격 차단.
//
// 디폴트 정책 (학습 + 운영 균형):
//   - 로그인/회원가입: 분당 10 회 (burst 5) — 사용자 반복 시도 허용 + 봇 차단
//   - 토큰 / 인증 엔드포인트: 분당 60 회 (burst 20) — 정상 클라이언트 트래픽 여유
//
// 메모리만 사용 (Redis 등 외부 의존 없음). 단일 인스턴스 가정 — HA 가면 외부 store 필요.
type ipLimiter struct {
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
	clients  map[string]*entry
}

type entry struct {
	limiter *rate.Limiter
	seen    time.Time
}

func newIPLimiter(perMinute float64, burst int) *ipLimiter {
	return &ipLimiter{
		rate:    rate.Every(time.Minute / time.Duration(perMinute)),
		burst:   burst,
		clients: make(map[string]*entry),
	}
}

func (l *ipLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.clients[ip]
	if !ok {
		e = &entry{limiter: rate.NewLimiter(l.rate, l.burst)}
		l.clients[ip] = e
	}
	e.seen = time.Now()
	return e.limiter.Allow()
}

// gc: 오래 안 본 IP 항목 제거 — 메모리 무한 증가 방지.
func (l *ipLimiter) gc(ttl time.Duration) {
	for {
		time.Sleep(ttl)
		cutoff := time.Now().Add(-ttl)
		l.mu.Lock()
		for ip, e := range l.clients {
			if e.seen.Before(cutoff) {
				delete(l.clients, ip)
			}
		}
		l.mu.Unlock()
	}
}

var (
	loginLimiter    = newIPLimiter(10, 5)
	tokenLimiter    = newIPLimiter(60, 20)
	registerLimiter = newIPLimiter(5, 3)
)

func init() {
	go loginLimiter.gc(10 * time.Minute)
	go tokenLimiter.gc(10 * time.Minute)
	go registerLimiter.gc(10 * time.Minute)
}

// clientIP: X-Forwarded-For 우선, 없으면 RemoteAddr.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RateLimitedFunc: HandlerFunc 래퍼. 초과 시 429 + Retry-After.
func RateLimitedFunc(lim *ipLimiter, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !lim.Allow(clientIP(r)) {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		h(w, r)
	}
}

// LoginLimiter / TokenLimiter / RegisterLimiter: 외부 노출 (router 가 wrap 할 때 사용).
func LoginLimiter() *ipLimiter    { return loginLimiter }
func TokenLimiter() *ipLimiter    { return tokenLimiter }
func RegisterLimiter() *ipLimiter { return registerLimiter }
