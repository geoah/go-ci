package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type (
	GithubClient struct {
		ghToken string
	}
	UpdateStatusRequest struct {
		State       string `json:"state"`
		TargetURL   string `json:"target_url"`
		Description string `json:"description"`
		Context     string `json:"context"`
	}
)

func (c *GithubClient) request(
	req *http.Request,
) error {
	req.Header.Set("Authorization", "token "+c.ghToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	if _, err := client.Do(req); err != nil {
		return err
	}

	return nil
}

func (c *GithubClient) UpdateStatus(
	repo string,
	sha string,
	task string,
	state string,
	description string,
) error {
	status := &UpdateStatusRequest{
		State:       state,
		TargetURL:   fmt.Sprintf("https://ci.nimona.io/jobs/%s-%s", task, sha),
		Description: description,
		Context:     "ci.nimona.io: " + task,
	}

	body, err := json.Marshal(status)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/statuses/%s", repo, sha)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	return c.request(req)
}
