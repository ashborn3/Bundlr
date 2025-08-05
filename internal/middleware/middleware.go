package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type clientKey struct {
	IP     string
	UserID string
}

var (
	mu        sync.Mutex
	limiters  = make(map[clientKey]*rate.Limiter)
	cleanupAt = time.Now().Add(10 * time.Minute)
)

func getKey(r *http.Request) clientKey {
	ip := getIP(r)
	userID := ""

	if uid, ok := r.Context().Value("userID").(string); ok {
		userID = uid
	}

	return clientKey{IP: ip, UserID: userID}
}

func getLimiter(key clientKey) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	var r rate.Limit
	var b int

	if key.UserID != "" {
		// Authenticated user
		r = rate.Every(time.Minute / 60) // 60 req/min
		b = 10
	} else {
		// IP-based unauthenticated
		r = rate.Every(time.Minute / 10) // 10 req/min
		b = 5
	}

	limiter, exists := limiters[key]
	if !exists {
		limiter = rate.NewLimiter(r, b)
		limiters[key] = limiter
	}

	if time.Now().After(cleanupAt) {
		for k := range limiters {
			if !limiters[k].Allow() {
				delete(limiters, k)
			}
		}
		cleanupAt = time.Now().Add(10 * time.Minute)
	}

	return limiter
}

func getIP(r *http.Request) string {
	// Handle X-Forwarded-For if behind proxy
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := getKey(r)
		limiter := getLimiter(key)

		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
