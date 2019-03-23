package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type (
	GithubWebhookHandler struct {
		githubClient *GithubClient
		kubeClient   *KubernetesClient
		tasks        []string
	}
	GithubPayload struct {
		Action      string `json:"action"`
		Number      int64  `json:"number"`
		PullRequest struct {
			URL    string `json:"url"`
			ID     int64  `json:"id"`
			Number int64  `json:"number"`
			State  string `json:"state"`
			Locked bool   `json:"locked"`
			Title  string `json:"title"`
			Head   struct {
				Label string `json:"label"`
				Ref   string `json:"ref"`
				Sha   string `json:"sha"`
			} `json:"head"`
			Base struct {
				Label string `json:"label"`
				Ref   string `json:"ref"`
				Sha   string `json:"sha"`
			} `json:"base"`
			Merged         bool   `json:"merged"`
			Mergeable      *bool  `json:"mergeable"`
			MergeableState string `json:"mergeable_state"`
			Comments       int64  `json:"comments"`
			ReviewComments int64  `json:"review_comments"`
			Commits        int64  `json:"commits"`
			Additions      int64  `json:"additions"`
			Deletions      int64  `json:"deletions"`
			ChangedFiles   int64  `json:"changed_files"`
		} `json:"pull_request"`
		Repository struct {
			ID       int64  `json:"id"`
			Name     string `json:"name"`
			FullName string `json:"full_name"`
		} `json:"repository"`
		Sender struct {
			Login string `json:"login"`
		} `json:"sender"`
	}
)

func (wh *GithubWebhookHandler) Handle(
	w http.ResponseWriter,
	r *http.Request,
) {
	fmt.Println("New request, type " + r.Header.Get("X-GitHub-Event"))

	eventName := r.Header.Get("X-GitHub-Event")
	if eventName == "" {
		log.Println("Missing X-GitHub-Event header")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		log.Println("Could not read payload", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	event := &GithubPayload{}
	if err := json.Unmarshal(payload, event); err != nil {
		log.Println("Could not decode payload", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch eventName {
	case "pull_request":
		fmt.Printf(
			"Got Pull Request event for %s#%d %s\n",
			event.Repository.FullName,
			event.PullRequest.Number,
			event.PullRequest.Head.Sha,
		)
		if err := wh.handlePR(event); err != nil {
			log.Println("Could not handle PR", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	default:
		fmt.Printf("Got unknown event, %s\n", string(payload))
	}
}

func (wh *GithubWebhookHandler) handlePR(
	event *GithubPayload,
) error {
	for _, task := range wh.tasks {
		if err := wh.runTask(event, task); err != nil {
			return err
		}
	}
	return nil
}

func (wh *GithubWebhookHandler) runTask(
	event *GithubPayload,
	task string,
) error {
	if err := wh.githubClient.UpdateStatus(
		event.Repository.FullName,
		event.PullRequest.Head.Sha,
		task,
		"pending",
		"In progress",
	); err != nil {
		return err
	}

	return wh.kubeClient.CreateJob(
		event.Repository.FullName,
		event.PullRequest.Head.Sha,
		task,
	)
}
