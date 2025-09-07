package httpx

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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

func RateLimit(next http.Handler) http.Handler {
	perMin, _ := strconv.Atoi(os.Getenv("QURAN_RATE_PER_MIN"))
	if perMin <= 0 { perMin = 120 }
	tokens := perMin
	ticker := time.NewTicker(time.Minute)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-ticker.C:
			tokens = perMin
		default:
		}
		if tokens <= 0 {
			http.Error(w, "rate limit", http.StatusTooManyRequests); return
		}
		tokens--
		next.ServeHTTP(w, r)
	})
}
