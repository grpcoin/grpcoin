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
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"github.com/yuin/goldmark"
)

type HomeResponse struct {
}

func (fe *frontend) home(w http.ResponseWriter, r *http.Request) error {

	return tpl.ExecuteTemplate(w, "home.tmpl", HomeResponse{})
}

var (
	//go:embed templates/join.md
	joinContent []byte

	joinRender   sync.Once
	joinRendered string
)

type JoinHandlerData struct {
	Content template.HTML
}

func (fe *frontend) join(w http.ResponseWriter, _ *http.Request) error {
	var b bytes.Buffer
	var err error
	joinRender.Do(func() {
		err = goldmark.Convert(joinContent, &b)
		joinRendered = b.String()
	})
	if err != nil {
		return fmt.Errorf("markdown rendering error: %w", err)
	}
	if joinRendered == "" {
		return fmt.Errorf("markdown didn't render for this page")
	}
	return tpl.ExecuteTemplate(w, "join.tmpl", JoinHandlerData{Content: template.HTML(joinRendered)})
}
