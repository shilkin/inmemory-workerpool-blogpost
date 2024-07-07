package analytics

import (
	"context"
	"time"

	"github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/delay"
)

type Analytics struct{}

func NewAnalytics() *Analytics {
	return &Analytics{}
}

func (_ *Analytics) Send(ctx context.Context, message string, args ...string) {
	d := delay.FromContext(ctx)
	println("Analytics.Send " + d.String())
	<-time.After(d)
}
