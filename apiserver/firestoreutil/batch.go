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
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// BatchDeleteAll deletes all returned results in batches of sizes allowed by firestore.
func BatchDeleteAll(ctx context.Context, cl *firestore.Client, it *firestore.DocumentIterator) error {
	const maxBatchSize = 500 // hardcoded on firestore API
	wb := cl.Batch()
	var deleted int
	commit := func() error {
		if _, err := wb.Commit(ctx); err != nil {
			return err
		}
		wb = cl.Batch()
		deleted = 0
		return nil
	}
	for {
		doc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		wb.Delete(doc.Ref)
		deleted++

		if deleted == maxBatchSize {
			if err := commit(); err != nil {
				return err
			}
		}
	}
	if deleted == 0 {
		return nil
	}
	return commit()
}
