package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	ghToken := os.Getenv("GH_TOKEN")
	if ghToken == "" {
		log.Fatal("Missing Github secret")
	}

	kubeSecret := os.Getenv("KUBERNETES_SECRET")
	if kubeSecret == "" {
		sPath := "/var/run/secrets/kubernetes.io/serviceaccount/token"
		if _, err := os.Stat(sPath); err == nil {
			d, err := ioutil.ReadFile(sPath)
			if err != nil {
				log.Println("could not load kube token", err)
			}
			kubeSecret = string(d)
		}
	}
	if kubeSecret == "" {
		log.Fatal("Missing Kubernetes secret")
	}

	ghClient := &GithubClient{
		ghToken: ghToken,
	}

	kubeClient := &KubernetesClient{
		kubeToken: kubeSecret,
		ghToken:   ghToken,
	}

	ghWebhookHandler := &GithubWebhookHandler{
		githubClient: ghClient,
		kubeClient:   kubeClient,
		tasks: []string{
			"test",
			"build",
		},
	}

	http.HandleFunc("/webhooks", ghWebhookHandler.Handle)

	fmt.Println("Listening on :8000")
	http.ListenAndServe(":8000", nil)
}
