// Copyright 2021 Ahmet Alp Balkan
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
