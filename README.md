## go-ci

This is an attempt at creating the simplest CI tool possible on top of 
kubernetes with minimal or no external dependencies and configuration.

It is not meant to replace anyone's CI and should not be used in anything that
resembles production environment.

### What it does

It listens for webhook events from one or more Github repos, and creates a 
kubernetes batch job for each of the tasks specified. 
Tasks are just the commands you want to run, ie `make lint`, `make test`.

Each job simply clones the repo, checks out the correct commit, and executes the
task.

Once the job is done, it reports back to Github the status of the execution for
that task.

### Installation

To install it go modify the `ci.yaml` to match your needs (mainly the secret
and hostname) and apply it via `kubectl apply -f ci.yaml`.
You should now be able to add the `/webhooks` endpoint to your Github's repo
webhooks.

### Notes

There isn't a docker image for go-ci, it just uses the `golang` image, and 
installs `go-cli` from source on start.

