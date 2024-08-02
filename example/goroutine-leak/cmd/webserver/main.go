package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	pyroscope "github.com/grafana/pyroscope-go"
	metrics "github.com/tevjef/go-runtime-metrics"

	"github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/analytics"
	"github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/delay"
	"github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/repo"
	"github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/service"
)

// main starts a web server with two endpoints:
// (1) /ping returns "pong"
// (2) /create creates a user with a given delay for testing.
func main() {
	stopObservability := startObservability()
	defer stopObservability()

	userService := service.NewUserService(
		repo.NewUserRepo(),
		analytics.NewAnalytics(),
	)

	server := http.DefaultServeMux

	server.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})

	server.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		name := "John Doe"
		email := "john.doe@example.com"
		d, err := time.ParseDuration(q.Get("delay")) // for testing
		if err != nil {
			println(err.Error())
		}

		fmt.Printf("create %v\n", q)

		ctx := delay.ToContext(r.Context(), d) // for testing

		if err := userService.Create(ctx, name, email); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	_ = http.ListenAndServe(":8080", server)
}

// startObservability starts observability tools
// and returns a function to stop them.
//
// (1) Start collecting runtime metrics.
// (2) Start profiling the application with Pyroscope.
func startObservability() func() {
	metrics.DefaultConfig.CollectionInterval = time.Second
	metrics.DefaultConfig.BatchInterval = time.Second
	metrics.DefaultConfig.Measurement = "go.runtime.webserver"
	metrics.DefaultConfig.Host = os.Getenv("INFLUXDB_HOST")
	if err := metrics.RunCollector(metrics.DefaultConfig); err != nil {
		panic("run collector " + err.Error())
	}

	println("PYROSCOPE_HOST = " + os.Getenv("PYROSCOPE_HOST"))
	println("INFLUXDB_HOST = " + os.Getenv("INFLUXDB_HOST"))

	profiler, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: "webserver",
		ServerAddress:   os.Getenv("PYROSCOPE_HOST"),
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
	if err != nil {
		panic("start pyroscope " + err.Error())
	}

	return func() { _ = profiler.Stop() }
}
