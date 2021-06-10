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
	"fmt"
	"net/http"

	"github.com/blendle/zapdriver"
	"github.com/purini-to/zapmw"
	"go.uber.org/zap"
)

func withStackdriverFields(log *zap.Logger, r *http.Request) *zap.Logger {
	payload := zapdriver.NewHTTP(r, nil)
	payload.RemoteIP = r.Header.Get("x-forwarded-for")
	return log.With(zapdriver.HTTP(payload))
}

func loggerFrom(ctx context.Context) *zap.Logger {
	v := ctx.Value(zapmw.ZapKey)
	if v == nil {
		panic("request did not have a logger")
	}
	vv, ok := v.(*zap.Logger)
	if !ok {
		panic(fmt.Sprintf("req ctx had wrong type of logger (%T)", vv))
	}
	return vv
}
