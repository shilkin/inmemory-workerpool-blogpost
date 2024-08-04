package analytics

import (
	"context"
)

type Analytics struct{}

func NewAnalytics() *Analytics {
	return &Analytics{}
}

func (_ *Analytics) Send(ctx context.Context, message string, args ...string) {
	// ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	// defer cancel()
	// d := delay.FromContext(ctx)
	// println("Analytics.Send " + d.String())
	// select {
	// case <-ctx.Done():
	// 	println("Analytics.Send " + ctx.Err().Error())
	// case <-time.After(d):
	// } // simulate slow analytics
}
