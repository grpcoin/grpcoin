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

package userdb

import (
	"context"
	"time"
)

type MockProfileCache struct{}

func (m MockProfileCache) GetTrades(_ context.Context, _ string) ([]TradeRecord, bool, error) {
	return nil, false, nil
}

func (m MockProfileCache) SaveTrades(_ context.Context, _ string, v []TradeRecord) error { return nil }

func (m MockProfileCache) InvalidateTrades(_ context.Context, _ string) error { return nil }

func (m MockProfileCache) GetValuation(_ context.Context, _ string, _ time.Time) ([]ValuationHistory, bool, error) {
	return nil, false, nil
}

func (m MockProfileCache) SaveValuation(_ context.Context, _ string, _ time.Time, v []ValuationHistory) error {
	return nil
}
