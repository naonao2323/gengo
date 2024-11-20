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
		target++
		defer mutex.Unlock()
		if expected <= target {
			cancel()
		}
	}
	return ctx, cancelWithCondition
}
