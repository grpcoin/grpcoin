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

package common

import (
	"strconv"
	"strings"

	"github.com/grpcoin/grpcoin/api/grpcoin"
)

func ParsePrice(p string) *grpcoin.Amount {
	out := strings.SplitN(p, ".", 2)
	if len(out) == 0 {
		return &grpcoin.Amount{}
	}
	if out[0] == "" {
		out[0] = "0"
	}
	i, _ := strconv.ParseInt(out[0], 10, 64)
	if len(out) == 1 {
		return &grpcoin.Amount{Units: i}
	}
	out[1] += strings.Repeat("0", 9-len(out[1]))
	j, _ := strconv.Atoi(out[1])
	return &grpcoin.Amount{Units: i, Nanos: int32(j)}
}
