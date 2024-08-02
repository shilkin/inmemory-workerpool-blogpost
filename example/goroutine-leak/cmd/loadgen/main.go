package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const requestTimeout = 1 * time.Second

func main() {
	concurrency := flag.Int("c", 3, "Number of concurrent workers")
	target := flag.String("t", "http://localhost:8080/create?delay=1s", "Target URL")
	interval := flag.Duration("i", 10*time.Millisecond, "Request interval")
	flag.Parse()

	fmt.Printf("target: %s\n", *target)
	fmt.Printf("concurrency: %d\n", *concurrency)
	fmt.Printf("interval: %s\n", *interval)

	stopC := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(*concurrency)

	for i := 0; i < *concurrency; i++ {
		go func() {
			defer wg.Done()

			ticker := time.NewTicker(*interval)
			defer ticker.Stop()

			client := http.Client{Timeout: requestTimeout}
			url := *target

			for {
				select {
				case <-stopC:
					return
				case <-ticker.C:
					if _, err := client.Get(url); err != nil {
						fmt.Printf("error: %s\n", err)
						continue
					}

					fmt.Printf("request to %s done\n", url)
				}
			}
		}()
	}

	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	<-signalC
	close(stopC)

	doneC := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneC)
	}()

	select {
	case <-doneC:
		fmt.Println("done (OK)")
	case <-time.After(5 * time.Second):
		fmt.Println("done (stop timeout)")
	}
}
