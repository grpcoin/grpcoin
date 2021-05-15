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
	"testing"

	"github.com/go-redis/redismock/v8"
)

type mockToken string

func (g mockToken) V() string { return string(g) }

func TestAccountCache(t *testing.T) {
	rc, mock := redismock.NewClientMock()
	c := &AccountCache{cache: rc}

	tok := mockToken("123")
	tokHash := "token_a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3"

	// cache miss
	mock.ExpectGet(tokHash).RedisNil()
	v, ok, err := c.Get(tok)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("wasn't expecting 123 in the cache")
	}

	// cache hit
	mock.ExpectGet(tokHash).SetVal("abc")
	v, ok, err = c.Get(tok)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("was expecting 123 in the cache")
	}
	if v != "abc" {
		t.Fatalf("unexpected value %q", v)
	}

	// set
	tok2 := mockToken("345")
	tok2key := c.cacheKey(tok2)
	mock.ExpectSet(tok2key, "cde", 0).SetVal("cde")
	err = c.Set(tok2, "cde")
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAccountCache_cacheKey(t *testing.T) {
	tok := "123"
	got := (&AccountCache{}).cacheKey(mockToken(tok))
	expected := "token_a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3"
	if got != expected {
		t.Fatalf("got:%s expected:%s", got, expected)
	}
}
