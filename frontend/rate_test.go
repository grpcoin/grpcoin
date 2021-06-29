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
	"net/http"
	"net/textproto"
	"testing"
)

func TestFindIP(t *testing.T) {
	tests := []struct {
		name    string
		req     *http.Request
		want    string
		wantErr bool
	}{
		{name: "xri",
			req: &http.Request{Header: map[string][]string{
				textproto.CanonicalMIMEHeaderKey("x-real-ip"): {"1.2.3.4"}}},
			want:    "1.2.3.4",
			wantErr: false},
		{name: "xff",
			req: &http.Request{Header: map[string][]string{
				textproto.CanonicalMIMEHeaderKey("x-forwarded-for"): {"1.2.3.4"}}},
			want:    "1.2.3.4",
			wantErr: false},
		{name: "xff multi",
			req: &http.Request{Header: map[string][]string{
				textproto.CanonicalMIMEHeaderKey("x-forwarded-for"): {"1.0.0.0, 1.1.1.1"}}},
			want:    "1.0.0.0",
			wantErr: false},
		{name: "xri over xff",
			req: &http.Request{Header: map[string][]string{
				textproto.CanonicalMIMEHeaderKey("x-real-ip"):       {"2.2.2.2"},
				textproto.CanonicalMIMEHeaderKey("x-forwarded-for"): {"0.0.0.0, 1.1.1.1"}}},
			want:    "2.2.2.2",
			wantErr: false},
		{name: "peer",
			req:     &http.Request{RemoteAddr: "1.2.3.4:44444"},
			want:    "1.2.3.4",
			wantErr: false},
		{name: "peer malformed",
			req:     &http.Request{RemoteAddr: "hello"},
			wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findIP(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("findIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("findIP() got = %v, want %v", got, tt.want)
			}
		})
	}
}
