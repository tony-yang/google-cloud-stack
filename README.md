# google-cloud-stack
An experimental repo building app using Google cloud stack.

The project will be wrapped inside a Docker container for isolation purpose.

## Stack List
- GAE
- GCS
- Cloud SQL
- Cloud Pub/Sub

## Dev Guide
To build a new Docker container with everything set up, run `make`.
Run a container, and play with the various tools.

To run the test, run `make test`

## References
This project references the various documentation and tutorials from `cloud.google.com`.
The demo prject comes from the Go getting started tutorial app.
- https://cloud.google.com/go/getting-started/tutorial-app