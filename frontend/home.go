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
	"html/template"
	"net/http"

	"github.com/yuin/goldmark"
)

type StaticPage struct {
	Content template.HTML
}

var (
	//go:embed content/join.md
	joinMD []byte
	//go:embed content/game.md
	gameMD []byte

	gameContent, joinContent template.HTML
)

func init() {
	for _, p := range []struct {
		src []byte
		out *template.HTML
	}{
		{joinMD, &joinContent},
		{gameMD, &gameContent},
	} {
		var b bytes.Buffer
		if err := goldmark.Convert(p.src, &b); err != nil {
			panic(err)
		}
		*p.out = template.HTML(b.String())
	}
}

func (fe *frontend) home(w http.ResponseWriter, r *http.Request) error {
	return tpl.ExecuteTemplate(w, "home.tmpl", nil)
}

func (fe *frontend) join(w http.ResponseWriter, _ *http.Request) error {
	return tpl.ExecuteTemplate(w, "join.tmpl", StaticPage{Content: joinContent})
}

func (fe *frontend) rules(w http.ResponseWriter, _ *http.Request) error {
	return tpl.ExecuteTemplate(w, "game.tmpl", StaticPage{Content: gameContent})
}
