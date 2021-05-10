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
	"context"
	"log"
	"time"

	"github.com/ahmetb/grpcoin/gdax"
)

func main() {
	ch, err := gdax.StartWatch(context.Background(), "BTC-USD")
	if err != nil {
		panic(err)
	}
	ch = gdax.RateLimited(ch, time.Millisecond*5)
	log.SetFlags(log.Lmicroseconds)
	for v := range ch {
		log.Printf("[%s] %s: %v", v.Time, v.Product, v.Price)
	}
}
