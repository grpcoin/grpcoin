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
	"context"
	"testing"
	"time"
)

func TestRateLimited(t *testing.T) {
	ctx, cleanup := context.WithTimeout(context.Background(), time.Millisecond*450)
	defer cleanup()

	in := make(chan Quote)
	go func() {
		for {
			if ctx.Err() != nil {
				close(in)
				return
			}
			in <- Quote{}
		}
	}()

	// 100*4=400 < 450
	out := RateLimited(in, time.Millisecond*100)
	count := 0
	for range out {
		count++
	}
	if expected := 5; count != expected {
		t.Fatalf("wrong msg recv count:%d expected:%d", count, expected)
	}
}

func TestRateLimitedPerSymbol(t *testing.T) {
	ctx, cleanup := context.WithTimeout(context.Background(), time.Millisecond*450)
	defer cleanup()

	in := make(chan Quote)
	go func() {
		for {
			if ctx.Err() != nil {
				close(in)
				return
			}
			in <- Quote{Product: "X"}
			in <- Quote{Product: "Y"}
		}
	}()

	// 100*4=400 < 450
	out := RateLimited(in, time.Millisecond*100)
	xCount, yCount := 0, 0
	for v := range out {
		switch v.Product {
		case "X":
			xCount++
		case "Y":
			yCount++
		}
	}
	if expected := 5; xCount != expected {
		t.Fatalf("wrong X msg recv count:%d expected:%d", xCount, expected)
	}
	if expected := 5; yCount != expected {
		t.Fatalf("wrong Y msg recv count:%d expected:%d", xCount, expected)
	}
}
