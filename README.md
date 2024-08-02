# Online In-Memory Task Processing
## Introduction
This is a clean project for exploring online in-memory task processing.

## Makefile commands
```sh
$ make help
format            Format golang code
lint              Run linter
test              Run tests
example-start     Start example of goroutine leak
example-stop      Stop example of goroutine leak
example-refresh   Build and start example of goroutine leak
example-load      Build and start example loadgen
example-restart   Restart example of goroutine leak
```

## Run and monitor

### Run example
```sh
make example-start # or example-refresh
```

### Monitor example
* Grafana: http://localhost:3000
* Pyroscope: http://localhost:4040

## Source code
* `example/goroutine-leak/cmd/webserver/main.go`: [main entry point](example/goroutine-leak/cmd/webserver/main.go)
* `example/goroutine-leak/cmd/webserver/internal/service/user.go`: [problematic user service](example/goroutine-leak/cmd/webserver/internal/service/user.go)

