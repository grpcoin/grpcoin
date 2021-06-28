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

package realtimequote

import (
	"time"

	"golang.org/x/time/rate"
)

// RateLimited provides per-symbol rate limiting on ch, limiting to
// one output per d.
func RateLimited(ch <-chan Quote, d time.Duration) <-chan Quote {
	// TODO write tests
	out := make(chan Quote)

	limiters := make(map[string]*rate.Limiter)

	go func() {
		for m := range ch {
			var l *rate.Limiter
			if v, ok := limiters[m.Product]; !ok {
				l = rate.NewLimiter(rate.Every(d), 1)
				limiters[m.Product] = l
			} else {
				l = v
			}
			if l.Allow() {
				out <- m
			}
			continue
		}
		close(out)
	}()
	return out
}
