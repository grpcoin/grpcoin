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
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/firestore"
)

type Doc struct {
	Path  string `json:"p"`
	Value []byte `json:"v"`
}

// ImportData loads data into firestore from a newline-separated json doc
// file with objects in format:
//      {"p": "col/<NAME>/<SUBCOL>/<NAME>", "v": {base64 then gob-encoded map[string]interface{}}
//      {"p": "col/<NAME>/<SUBCOL>/<NAME>", "v": {base64 then gob-encoded map[string]interface{}}
func ImportData(r io.Reader, c *firestore.Client) error {

	batch := c.Batch()
	written := 0
	batchSize := 400

	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	gob.Register(map[string]interface{}{})
	gob.Register(time.Time{})
	for {
		var v Doc
		for {
			err := d.Decode(&v)
			if err == io.EOF {
				break
			} else if err != nil {
				return fmt.Errorf("failed to decode json: %w", err)
			}
			var data map[string]interface{}
			if err := gob.NewDecoder(bytes.NewReader(v.Value)).Decode(&data); err != nil {
				return fmt.Errorf("failed to decode gob: %w", err)
			}
			batch.Create(c.Doc(v.Path), data)
			written++
			v = Doc{} // reset for reuse
			if written == batchSize {
				break
			}
		}
		if written == 0 {
			break
		} else {
			if _, err := batch.Commit(context.TODO()); err != nil {
				return fmt.Errorf("failed to commit batch: %w", err)
			}
			written = 0
			batch = c.Batch() // start new batch
		}
	}
	return nil
}
