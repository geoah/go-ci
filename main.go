package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	ghSecret := os.Getenv("GITHUB_SECRET")
	if ghSecret == "" {
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
		secret: ghSecret,
	}

	kubeClient := &KubernetesClient{
		secret: kubeSecret,
	}

	ghWebhookHandler := &GithubWebhookHandler{
		githubClient: ghClient,
		kubeClient:   kubeClient,
		tasks: []string{
			"lint",
			"test",
			"build",
		},
	}

	http.HandleFunc("/webhooks", ghWebhookHandler.Handle)

	fmt.Println("Listening on :8000")
	http.ListenAndServe(":8000", nil)
}
