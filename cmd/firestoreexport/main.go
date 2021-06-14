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
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"flag"
	"os"
	"regexp"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/grpcoin/grpcoin/apiserver/firestoreutil"
)

var (
	flProj string
)

func init() {
	flag.StringVar(&flProj, "project", "", "gcp project")
	flag.Parse()
}

func main() {
	ctx := context.Background()
	collections := []string{
		"users",
		"orders",
		"valuations",
	}
	fs, err := firestore.NewClient(ctx, flProj)
	if err != nil {
		panic(err)
	}

	gob.Register(map[string]interface{}{})
	gob.Register(time.Time{})
	w := json.NewEncoder(os.Stdout)
	for _, col := range collections {
		docs, err := fs.CollectionGroup(col).Documents(ctx).GetAll()
		if err != nil {
			panic(err)
		}
		for _, d := range docs {
			data := d.Data()
			var b bytes.Buffer
			if err := gob.NewEncoder(&b).Encode(&data); err != nil {
				panic(err)
			}
			w.Encode(firestoreutil.Doc{Path: clearPathPrefix(d.Ref.Path), Value: b.Bytes()})
		}
	}
}

func clearPathPrefix(s string) string {
	return regexp.MustCompile(`^projects/[^/]*/databases/[^/]*/documents/`).ReplaceAllString(s, "")
}
