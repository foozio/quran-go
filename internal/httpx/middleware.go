package httpx

import (
    "net"
    "net/http"
    "os"
    "strconv"
    "strings"
    "sync"

    "golang.org/x/time/rate"
)

func CORS(next http.Handler) http.Handler {
	origins := strings.Split(os.Getenv("QURAN_ALLOWED_ORIGINS"), ",")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o := "*"
		if len(origins) > 0 && origins[0] != "" { o = origins[0] }
		w.Header().Set("Access-Control-Allow-Origin", o)
		w.Header().Set("Vary", "Origin")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
			w.WriteHeader(http.StatusNoContent); return
		}
		next.ServeHTTP(w, r)
	})
}

// RateLimit applies a simple per-IP token bucket.
// Not production-hardened (no eviction); good enough for small deployments.
func RateLimit(next http.Handler) http.Handler {
    perMin, _ := strconv.Atoi(os.Getenv("QURAN_RATE_PER_MIN"))
    if perMin <= 0 { perMin = 120 }
    rps := rate.Limit(float64(perMin) / 60.0)
    burst := perMin

    var (
        mu       sync.Mutex
        limiters = map[string]*rate.Limiter{}
    )
    getLimiter := func(ip string) *rate.Limiter {
        mu.Lock(); defer mu.Unlock()
        if l, ok := limiters[ip]; ok { return l }
        l := rate.NewLimiter(rps, burst)
        limiters[ip] = l
        return l
    }
    clientIP := func(r *http.Request) string {
        // Honor X-Forwarded-For (first IP) and X-Real-IP, else RemoteAddr
        if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
            if i := strings.IndexByte(xff, ','); i >= 0 {
                return strings.TrimSpace(xff[:i])
            }
            return strings.TrimSpace(xff)
        }
        if xr := r.Header.Get("X-Real-IP"); xr != "" { return strings.TrimSpace(xr) }
        host, _, err := net.SplitHostPort(r.RemoteAddr)
        if err != nil { return r.RemoteAddr }
        return host
    }

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := clientIP(r)
        if !getLimiter(ip).Allow() {
            http.Error(w, "rate limit", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
