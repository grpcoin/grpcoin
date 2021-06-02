package firestoreutil

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"time"

	firestore "cloud.google.com/go/firestore"
)

// ImportData loads data into firestore from a newline-separated json doc
// file with objects in format:
//      {"p": "col/<NAME>/<SUBCOL>/<NAME>", "v": {base64 then gob-encoded map[string]interface{}}
//      {"p": "col/<NAME>/<SUBCOL>/<NAME>", "v": {base64 then gob-encoded map[string]interface{}}
func ImportData(r io.Reader, c *firestore.Client) error {
	type Doc struct {
		Path  string `json:"p"`
		Value []byte `json:"v"`
	}

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
