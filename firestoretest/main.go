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

// Small program to play with firestore concurrency/tx.
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
)

func main() {
	ctx := context.Background()
	projectID := "grpcoin"

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	uid := fmt.Sprintf("user%d", r.Int())
	now := time.Now()
	doc, _, err := client.Collection("users").Add(ctx, map[string]interface{}{
		"id":   uid,
		"cash": 0,
	})
	if err != nil {
		panic(err)
	}
	log.Println(time.Since(now))
	var wg sync.WaitGroup

	var mu sync.Mutex
	completed := 0
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			now := time.Now()
			if err := run(client, doc.ID); err != nil {
				mu.Lock()
				defer mu.Unlock()
				// log.Printf("failed tx %s. completed=%d %v", uid, completed, err)
				return
			}
			mu.Lock()
			completed++
			mu.Unlock()
			log.Println(time.Since(now))
		}()
	}
	wg.Wait()
	log.Println(completed, doc.ID)
}

func run(client *firestore.Client, docID string) error {
	ref := client.Collection("users").Doc(docID)
	var cur int64
	err := client.RunTransaction(context.TODO(), func(_ context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref)
		if err != nil {
			return err
		}
		val, err := doc.DataAt("cash")
		if err != nil {
			return err
		}
		cur = val.(int64)
		return tx.Set(ref, map[string]interface{}{
			"cash": cur + 1,
		}, firestore.MergeAll)
	}, firestore.MaxAttempts(1))

	if err != nil {
		return err
	}
	log.Printf("%d->%d", cur, cur+1)
	return err
}
