package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/grpcoin/grpcoin/ratelimiter"
	"go.uber.org/zap"
)

const (
	frontendPerIPPerMinRateLimit = 60
)

func rateLimiting(rl ratelimiter.RateLimiter, log *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ip, err := findIP(req)
			if err != nil {
				handleErr(log, w, 0, fmt.Errorf("failed to find peer IP: %w", err))
				return
			}
			rateKey := "frontend_ip__" + ip
			if err := rl.Hit(req.Context(), rateKey, frontendPerIPPerMinRateLimit); err != nil {
				handleErr(log, w, http.StatusTooManyRequests, err)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}

// findIP returns the client's IP address from x-real-ip or x-forwarded-for headers,
// or falling back to connected peer IP.
func findIP(req *http.Request) (string, error) {
	if xri := req.Header.Get("x-real-ip"); xri != "" {
		return xri, nil
	}
	if xff := req.Header.Get("x-forwarded-for"); xff != "" {
		ips := strings.Split(xff, ",")
		return ips[0], nil
	}
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	return ip, err

}
