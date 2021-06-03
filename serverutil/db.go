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
	"context"
	"errors"
	"fmt"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"

	"github.com/grpcoin/grpcoin/apiserver/firestoreutil"
)

func DetectDatabase(ctx context.Context, datasetFile string, onCloud, useProdDB bool) (client *firestore.Client, shutdown func(), err error) {
	log := ctxzap.Extract(ctx)
	if !onCloud && !useProdDB {
		return GetLocalDB(ctx, datasetFile)
	}
	proj := firestore.DetectProjectID
	if useProdDB {
		proj = os.Getenv("GOOGLE_CLOUD_PROJECT")
		if proj == "" {
			return nil, nil, errors.New("please set GOOGLE_CLOUD_PROJECT environment variable")
		}
		log.Debug("project id is explicitly set", zap.String("project", proj))
	}
	fs, err := GetProdDB(ctx, proj)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect firestore: %w", err)
	}
	return fs, func() {fs.Close()}, nil
}

func GetProdDB(ctx context.Context, project string) (*firestore.Client, error) {
	return firestore.NewClient(ctx, project)
}

func GetLocalDB(ctx context.Context, datasetFile string) (client *firestore.Client, shutdown func(), err error) {
	log := ctxzap.Extract(ctx)
	log.Info("starting a local firestore emulator")
	f, shutdownEmulator, err := firestoreutil.StartEmulator(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize local firestore emulator: %w", err)
	}
	closeFn := func() {
		log.Debug("shutting down firestore emulator")
		shutdownEmulator()
	}

	log.Debug("loading test data", zap.String("file", datasetFile))
	td, err := os.Open(datasetFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open local test data file: %w", err)
	}
	if err := firestoreutil.ImportData(td, f); err != nil {
		return nil, nil, fmt.Errorf("failed to load test data: %w", err)
	}
	log.Info("loaded test data into the local firestore emulator")
	return f, closeFn, nil
}
