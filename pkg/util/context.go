package util

import (
	"context"
	"sync"
)

func WithCondition(ctx context.Context, expected int) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)
	var target int
	mutex := &sync.Mutex{}
	cancelWithCondition := func() {
		mutex.Lock()
		defer mutex.Unlock()
		target++
		if expected <= target {
			cancel()
		}
	}
	return ctx, cancelWithCondition
}
