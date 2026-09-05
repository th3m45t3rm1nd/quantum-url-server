package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func encode(n uint64) string {
	if n == 0 {
		return "0"
	}

	var out []byte

	for n > 0 {
		out = append(out, alphabet[n%62])
		n /= 62
	}

	for i, j := 0, len(out)-1; i < j; i, j = i+1, j+1 {
		out[i], out[j] = out[j], out[i]
	}

	return string(out)
}

func decode(str string) uint64 {
	var out int = 0
	for _, s := range str {
		pos := strings.IndexRune(alphabet, s)
		if pos == -1 {
			panic("Invalid Character")
		}

		out = out*62 + pos
	}

	return uint64(out)

}

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	bucket   map[string]*bucket
	capacity float64
	rate     float64
}

func newLimiter(capacity, rate float64) *rateLimiter {
	l := &rateLimiter{
		bucket:   make(map[string]*bucket),
		capacity: capacity,
		rate:     rate,
	}
	go l.sweep()
	return l
}

func (l *rateLimiter) sweep() {
	for {
		time.Sleep(5 * time.Minute)
		l.mu.Lock()

		for ip, b := range l.bucket {
			if time.Since(b.lastRefill) > 10*time.Minute {
				delete(l.bucket, ip)
			}
		}

		l.mu.Unlock()
	}
}

func (l *rateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, exists := l.bucket[ip]

	if !exists {
		b = &bucket{l.capacity, time.Now()}
		l.bucket[ip] = b
	}

	elapsed := time.Since(b.lastRefill).Seconds()
	b.tokens = min(l.capacity, b.tokens+elapsed*l.rate)
	b.lastRefill = time.Now()

	if b.tokens >= 1 {
		b.tokens--
		return true
	}

	return false
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

type clickEvent struct {
	Code      string
	Timestamp time.Time
	UserAgent string
	Referrer  string
}

var clickChan = make(chan clickEvent, 1000)

func startClickWorker(pool *pgxpool.Pool) {
	go func() {
		for evt := range clickChan {
			_, err := pool.Exec(context.Background(),
				"INSERT INTO clicks (code, clicked_at, user_agent, referrer) VALUES ($1, $2, $3, $4)",
				evt.Code, evt.Timestamp, evt.UserAgent, evt.Referrer)
			if err != nil {
				log.Printf("click insert failed: %v", err)
			}
		}
	}()
}
