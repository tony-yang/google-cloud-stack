# google-cloud-stack
An experimental repo building app using Google cloud stack.

The project will be wrapped inside a Docker container for isolation purpose.

## Stack List
- GAE
- GCS
- Cloud SQL
- Cloud Pub/Sub
- GCE
- GKE

## Dev Guide
To build a new Docker container with everything set up, run `make`.
Run a container, and play with the various tools.

Local development uses the cloud_sql_proxy for testing. GAE by default has configuration to access the Cloud SQL through its app.yaml with the `beta_settings`.

GCE and GKE however, doesn't have those access by default. See the GKE sidecar pattern with the [Cloud SQL Proxy](https://cloud.google.com/sql/docs/mysql/connect-kubernetes-engine) Docker image for detail.

In addition, we need to create a few [secrets](https://cloud.google.com/kubernetes-engine/docs/concepts/secret) using `kubectl` since the GKE yaml references those secrets for db password and oauth secrets. This is also needed for the Cloud SQL Proxy container to work.

## References
This project references the various documentation and tutorials from `cloud.google.com`.
The demo project comes from the Go getting started tutorial app and is modified as needed.
- https://cloud.google.com/go/getting-started/tutorial-app
