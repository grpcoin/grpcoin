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
	"flag"
	"fmt"
	"net"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/grpcoin/grpcoin/apiserver/firestoreutil"
	"github.com/grpcoin/grpcoin/userdb"
)

var flUseRealDB bool
var flProjectID string
var flEmulatorHost string

func init() {
	flag.BoolVar(&flUseRealDB, "use-real-db", false, "run against production database (requires -project set)")
	flag.StringVar(&flProjectID, "project", "", "gcp project id")
	flag.StringVar(&flEmulatorHost, "emulator-addr",
		net.JoinHostPort(firestoreutil.FirestoreEmulatorHost, firestoreutil.FirestoreEmulatorPort),
		"emulator addr")
	flag.Parse()
}
func main() {
	var fs *firestore.Client
	var err error
	ctx := context.Background()
	if flUseRealDB {
		if flProjectID == "" {
			panic("empty project id")
		}
		fs, err = firestore.NewClient(ctx, flProjectID)
		fmt.Println("using actual firestore db")
	} else {
		os.Setenv("FIRESTORE_EMULATOR_HOST", flEmulatorHost)
		fs, err = firestore.NewClient(ctx, firestoreutil.FirestoreEmulatorProject)
		os.Unsetenv("FIRESTORE_EMULATOR_HOST")
		fmt.Println("using local firestore emulator")
	}
	if err != nil {
		panic(err)
	}

	var confirm string
	fmt.Print("DANGER: delete the database records? (y/N): ")
	fmt.Scanln(&confirm)
	if confirm != "y" {
		panic("cancelled")
	}

	// clear valuations
	if err := firestoreutil.BatchDeleteAll(ctx, fs, fs.CollectionGroup("valuations").Documents(ctx)); err != nil {
		panic(err)
	}

	// clear orders
	if err := firestoreutil.BatchDeleteAll(ctx, fs, fs.CollectionGroup("orders").Documents(ctx)); err != nil {
		panic(err)
	}

	// reset user portfolio
	users, err := fs.Collection("users").Documents(ctx).GetAll()
	if err != nil {
		panic(err)
	}
	batch := fs.Batch()
	for i, u := range users {
		var uv userdb.User
		if err := u.DataTo(&uv); err != nil {
			panic(err)
		}
		uv.Portfolio = userdb.Portfolio{
			CashUSD:   userdb.Amount{Units: 100_000},
			Positions: nil,
		}
		batch.Set(u.Ref, uv)
		if i == len(users)-1 || i%100 == 99 {
			if _, err := batch.Commit(ctx); err != nil {
				panic(err)
			}
		}
	}
	fmt.Println("done")
}
