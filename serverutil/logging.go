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

package serverutil

import (
	"os"

	stackdriver "github.com/tommy351/zap-stackdriver"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetLogging(onCloud bool) (*zap.Logger, error) {
	if !onCloud {
		z, err := zap.NewDevelopment()
		if err != nil {
			return nil, err
		}
		z = z.With(zap.String("env", "dev"))
		return z, nil

	}
	c := zap.NewProductionConfig()
	c.EncoderConfig = stackdriver.EncoderConfig
	c.OutputPaths = []string{"stdout"}
	return c.Build(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &stackdriver.Core{
			Core: core,
		}
	}), zap.Fields(
		stackdriver.LogServiceContext(&stackdriver.ServiceContext{
			Service: os.Getenv("K_SERVICE"),
			Version: os.Getenv("K_REVISION"),
		}),
	))
}
