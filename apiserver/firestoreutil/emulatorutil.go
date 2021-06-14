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

// Package firestoreutil contains test utilities for starting a firestore
// emulator locally.
package firestoreutil

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
)

const FirestoreEmulatorProject = "dummy-emulator-firestore-project"
const FirestoreEmulatorHost = "localhost"
const FirestoreEmulatorPort = "8010"

// cBuffer is a buffer safe for concurrent use.
type cBuffer struct {
	b bytes.Buffer
	sync.Mutex
}

func (c *cBuffer) Write(p []byte) (n int, err error) {
	c.Lock()
	defer c.Unlock()
	return c.b.Write(p)
}

func StartEmulator(ctx context.Context) (*firestore.Client, func(), error) {
	addr := net.JoinHostPort(FirestoreEmulatorHost, FirestoreEmulatorPort)
	ctx, cancel := context.WithCancel(ctx)
	_ = cancel // to shush the linter

	// TODO investigate why there are still java processes hanging around
	// despite we kill the exec'd command, suspecting /bin/bash wrapper that gcloud
	// applies around the java process.
	cmd := exec.CommandContext(ctx, "gcloud", "beta", "emulators", "firestore", "start", "--host-port="+addr)
	out := &cBuffer{b: bytes.Buffer{}}
	cmd.Stderr, cmd.Stdout = out, out
	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start firestore emulator: %v -- out:%s", err, out.b.String())
	}
	dialCtx, clean := context.WithTimeout(ctx, time.Second*10)
	defer clean()
	var connected bool
	for !connected {
		select {
		case <-dialCtx.Done():
			return nil, nil, fmt.Errorf("emulator did not come up timely: %v -- output: %s", dialCtx.Err(), out.b.String())
		default:
			c, err := (&net.Dialer{Timeout: time.Millisecond * 200}).DialContext(ctx, "tcp", addr)
			if err == nil {
				c.Close()
				connected = true
				break
			}
			time.Sleep(time.Millisecond * 200) //before retrying
		}
	}
	os.Setenv("FIRESTORE_EMULATOR_HOST", addr)
	cl, err := firestore.NewClient(ctx, FirestoreEmulatorProject)
	if err != nil {
		return nil, nil, err
	}
	os.Unsetenv("FIRESTORE_EMULATOR_HOST")
	err = truncateDB(addr)
	return cl, cancel, err
}

func truncateDB(addr string) error {
	// technique adopted from https://stackoverflow.com/a/58866194/54929
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s/emulator/v1/projects/%s/databases/(default)/documents",
		addr, FirestoreEmulatorProject), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to clear db: %v", resp.Status)
	}
	return nil
}
