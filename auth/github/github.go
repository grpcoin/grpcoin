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

package github

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GitHubUser struct {
	ID       uint64
	Username string
}

func (g GitHubUser) DBKey() string       { return fmt.Sprintf("github_%v", g.ID) }
func (g GitHubUser) DisplayName() string { return g.Username }
func (g GitHubUser) ProfileURL() string  { return "https://github.com/" + g.Username }

func VerifyUser(token string) (GitHubUser, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Set("accept", "application/vnd.github.v3+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return GitHubUser{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var v struct {
			Message string `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&v)
		return GitHubUser{}, fmt.Errorf("github: failed to authenticate (%d): %s", resp.StatusCode, v.Message)
	}
	var user struct {
		Login string `json:"login"`
		Id    uint64 `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&user)
	return GitHubUser{ID: user.Id, Username: user.Login}, err
}
