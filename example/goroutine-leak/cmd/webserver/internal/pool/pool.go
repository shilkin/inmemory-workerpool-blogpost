package service

import (
	"context"
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"sync/atomic"
	"time"
)

// import std -> external -> internal
// const excl: type ConstType -> const ( Foo ConstType = iota )
// var
// type
// public func
// private func

var (
	successfulTasks = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "successful_tasks_total",
			Help: "Total number of successful tasks.",
		},
	)
	failedTasks = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "failed_tasks_total",
			Help: "Total number of failed tasks.",
		},
	)
	taskDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "task_duration_seconds",
			Help:    "Duration of task execution in seconds.",
			Buckets: prometheus.DefBuckets,
		},
	)
	enqueueTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "task_enqueue_time_seconds",
			Help:    "Time a task spends in the queue before execution in seconds.",
			Buckets: prometheus.DefBuckets,
		},
	)
	queueSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "task_queue_size",
			Help: "Current size of the task queue.",
		},
	)
)

// 1. Graceful shutdown
// 2. *Pass singal about pool stopping to task handlers
// 3. General purpose pool
type Task func(ctx context.Context) // TODO: think how to propagate pool stop signal

// TODO: Homework
//  1. prometheus metrics
//     - count succesful tasks
//     - count failed tasks
//     - time of task execution
//     - time of enqueueing*
//     - queue size
//
// 2. opentelemetry* -- source code + read about architecture and metrics
// 3. structred logging: slog
type PoolService struct {
	wg      sync.WaitGroup
	ch      chan TaskMessage
	running atomic.Bool
	stopC   chan struct{}
}

type TaskMessage struct {
	task Task
	ctx  context.Context
	inQ  time.Time
}

func NewPoolService(maxWorkers int) *PoolService {
	prometheus.MustRegister(successfulTasks, failedTasks, taskDuration, enqueueTime, queueSize)

	ps := PoolService{
		ch:    make(chan TaskMessage, 10000),
		stopC: make(chan struct{}),
	}

	ps.running.Store(true)

	for i := 0; i < maxWorkers; i++ {
		ps.wg.Add(1)

		go func() {
			defer ps.wg.Done()

			for {
				select {
				case taskMessage, ok := <-ps.ch:
					if !ok {
						return // TODO: check if it drains the channel when the channel is closed - NO
					}
					queueSize.Dec()
					enqueueTime.Observe(time.Since(taskMessage.inQ).Seconds())
					taskMessage.task(taskMessage.ctx)

				case <-ps.stopC:
					select {
					case taskMessage := <-ps.ch:
						queueSize.Dec()
						enqueueTime.Observe(time.Since(taskMessage.inQ).Seconds())
						taskMessage.task(taskMessage.ctx)
					default:
						return
					}
					return
				}
			}
		}()
	}

	return &ps
}

func (ps *PoolService) Enqueue(ctx context.Context, task Task) error {
	if ps.running.Load() != true {
		return errors.New("Worker are stopped")
	}

	select {
	case ps.ch <- TaskMessage{
		ctx:  context.WithoutCancel(ctx),
		task: recoverable(task),
		inQ:  time.Now(),
	}:
		queueSize.Inc()
	case <-ctx.Done():
		return ctx.Err() // context deadline exceeded
	}

	return nil
}

func (ps *PoolService) Stop() error {
	// TODO: read about atomic ops and sync/atomic
	if ps.running.CompareAndSwap(true, false) {
		return nil
	}

	close(ps.stopC) // <-

	timeout := 5 * time.Second // const or param

	doneC := make(chan struct{})

	go func() {
		ps.wg.Wait()
		close(doneC)
	}()

	select {
	case <-time.After(timeout):
		return errors.New("stop timout")
	case <-doneC:
		return nil
	}

}

func recoverable(task Task) Task {
	return func(ctx context.Context) {
		// TODO: how defer and recover work
		defer func() {
			if r := recover(); r != nil {
				failedTasks.Inc()
				//log.Printf("%v", r)
				//panicCounter.Inc()
				//alert()
			}
		}()

		// defer recoverPanic() doesn't recover
		start := time.Now()
		task(ctx)
		taskDuration.Observe(time.Since(start).Seconds())
		successfulTasks.Inc()
	}
}

// doesn't recover
// func recoverPanic() {
// 	if r := recover(); r != nil {
// 		// log.Printf("%v", r)
// 		// panicCounter.Inc()
// 		// alert()
//    }
// }
