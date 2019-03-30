package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"
)

type (
	KubernetesClient struct {
		kubeToken string
		ghToken   string
	}
)

func (c *KubernetesClient) request(
	req *http.Request,
) error {
	req.Header.Set("Authorization", "Bearer "+c.kubeToken)
	req.Header.Set("Content-Type", "application/yaml")
	// TODO should be loading ca from pod
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Job creation response: %d %s\n", res.StatusCode, string(contents))
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}

	return fmt.Errorf("unexpected response status code %d", res.StatusCode)
}

func (c *KubernetesClient) CreateJob(
	repo string,
	sha string,
	task string,
) error {
	bodyTpl := template.Must(template.New("main").Parse(`---
apiVersion: batch/v1
kind: Job
metadata:
  name: job-{{ .Sha }}-{{ .Task }}
spec:
  ttlSecondsAfterFinished: 86400
  template:
    metadata:
      name: job-{{ .Sha }}-{{ .Task }}
    spec:
      volumes:
      - name: ci-go-cache
        hostPath:
          path: /tmp/ci-go-cache
          type: DirectoryOrCreate
      containers:
      - name: golang
        image: golang:1.12
        env:
        - name: GOMODULE111
          value: "on"
        - name: GOCACHE
          value: /go-cache
        volumeMounts:
        - name: ci-go-cache
          mountPath: /go-cache
        command:
        - '/bin/bash'
        - '-c'
        - >-
          git clone -n https://github.com/{{ .Repo }} source
          && cd source
          && git checkout {{ .Sha }}
          && make {{ .Task }}
          && export CI_STATE=success || export CI_STATE=failure
          && curl -v 
          -X POST
          -H "Content-Type: application/json"
          -H "Authorization: token {{ .GithubToken }}"
          --data "{\"state\":\"$CI_STATE\",\"context\":\"ci.nimona.io: {{ .Task }}\"\"description\":\"$CI_STATE\"}"
          https://api.github.com/repos/{{ .Repo }}/statuses/{{ .Sha }}
      restartPolicy: Never
`))

	values := struct {
		Sha         string
		Task        string
		Repo        string
		GithubToken string
	}{
		Sha:         sha,
		Task:        task,
		Repo:        repo,
		GithubToken: c.ghToken,
	}

	buf := bytes.NewBuffer(nil)
	if err := bodyTpl.Execute(buf, values); err != nil {
		return err
	}

	url := "https://kubernetes.default.svc/apis/batch/v1/namespaces/default/jobs"
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return err
	}

	return c.request(req)
}
