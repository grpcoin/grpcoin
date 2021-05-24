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

package frontend

import (
	"reflect"
	"testing"
	"time"

	"github.com/grpcoin/grpcoin/server/userdb"
)

func TestPortfolioSnaphotAt(t *testing.T) {
	now := time.Date(2020, time.March, 31, 23, 00, 00, 00, time.UTC)
	type args struct {
		arr []userdb.ValuationHistory
		ago time.Duration
		now time.Time
	}
	tests := []struct {
		name string
		args args
		want *userdb.ValuationHistory
	}{
		{
			name: "not enough data",
			args: args{
				arr: nil,
				ago: time.Hour * 24,
				now: now,
			},
			want: nil,
		},
		{
			name: "data points less than filter duration",
			args: args{
				arr: []userdb.ValuationHistory{
					{Date: now.Add(-time.Hour * 3)},
					{Date: now.Add(-time.Hour * 2)},
					{Date: now.Add(-time.Hour * 1)},
				},
				ago: time.Hour * 24,
				now: now,
			},
			want: &userdb.ValuationHistory{Date: now.Add(-time.Hour * 3)},
		},
		{
			name: "more data points than selected timespan",
			args: args{
				arr: []userdb.ValuationHistory{
					{Date: now.Add(-time.Hour * 4)},
					{Date: now.Add(-time.Hour * 25)},
					{Date: now.Add(-time.Hour * 24)},
					{Date: now.Add(-time.Hour * 23)},
				},
				ago: time.Hour * 24,
				now: now,
			},
			want: &userdb.ValuationHistory{Date: now.Add(-time.Hour * 24)},
		},
		{
			name: "long durations",
			args: args{
				arr: []userdb.ValuationHistory{
					{Date: now.Add(-time.Hour * 24 * 366)},
					{Date: now.Add(-time.Hour * 24 * 365)},
					{Date: now.Add(-time.Hour * 24 * 364)},
					{Date: now.Add(-time.Hour * 24)},
					{Date: now},
				},
				ago: time.Hour * 24 * 365,
				now: now,
			},
			want: &userdb.ValuationHistory{Date: now.Add(-time.Hour * 24 * 365)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := portfolioSnaphotAt(tt.args.arr, tt.args.ago, tt.args.now); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("portfolioSnaphotAt() = %v, want %v", got, tt.want)
			}
		})
	}
}
