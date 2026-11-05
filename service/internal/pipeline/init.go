package pipeline

import (
	"context"
	"sync"
)

var (
	cancel context.CancelFunc
	wait   sync.WaitGroup
)

// Initalize 启动任务调度。
func Initalize() {
	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())
	wait.Go(func() { schedule(ctx) })
}

// Dispose 回收任务调度。
func Dispose() {
	cancel()
	wait.Wait()
}
