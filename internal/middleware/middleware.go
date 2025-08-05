package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var (
	visitors  = make(map[string]*rate.Limiter)
	mu        sync.Mutex
	rateLimit = rate.Every(200 * time.Millisecond) // 5 req/sec
	burst     = 5
)

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rateLimit, burst)
		visitors[ip] = limiter
	}

	return limiter
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr // optionally strip port
		limiter := getVisitor(ip)

		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
