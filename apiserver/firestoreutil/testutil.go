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

package firestoreutil

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	"testing"
)

func StartTestEmulator(t *testing.T, ctx context.Context) *firestore.Client {
	t.Helper()
	f, cleanup, err := StartEmulator(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		t.Log("shutting down firestore operator")
		cleanup()
	})
	return f
}
