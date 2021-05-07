package github

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type AuthenticatedUser interface {
	Key() string
	DisplayName()
}

type GitHubUser struct {
	ID       uint64
	Username string
}

func (g GitHubUser) Key() string         { return fmt.Sprintf("%v", g.ID) }
func (g GitHubUser) DisplayName() string { return g.Username }

func VerifyUser(token string) (GitHubUser, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Set("accept", "application/vnd.github.v3+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return GitHubUser{}, err
	}
	var user struct {
		Login string `json:"login"`
		Id    uint64 `json:"id"`
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&user)
	return GitHubUser{ID: user.Id, Username: user.Login}, err
}
