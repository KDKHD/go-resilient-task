## Uber FX example

### Steps to run example
- Launch the docker containers `docker compose up`
- run `go run .`

Observe how before running `go run .` the task in the database was in a `WAITING` state. After starting the program, the task should be in a `DONE` state.