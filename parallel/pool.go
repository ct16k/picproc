package parallel

import (
	"runtime"
	"sync"
)

type (
	WorkerFunc func(func())
	WaitFunc   func(done bool)
	CancelFunc func()
)

type Pool struct {
	wg     sync.WaitGroup
	Do     WorkerFunc
	Wait   WaitFunc
	Cancel CancelFunc
}

func Start(numWorkers int) *Pool {
	if numWorkers < 1 {
		numWorkers = runtime.GOMAXPROCS(0)
	}

	pool := &Pool{
		Do: func(f func()) {
			f()
		},
		Wait:   func(bool) {},
		Cancel: func() {},
	}

	if numWorkers > 1 {
		workChan := make(chan func(), numWorkers)

		for range numWorkers {
			pool.wg.Go(func() {
				for {
					f, ok := <-workChan
					if !ok {
						return
					}
					f()
				}
			})
		}

		pool.Do = func(f func()) {
			workChan <- f
		}

		pool.Wait = func(done bool) {
			if done {
				pool.Cancel()
			}
			pool.wg.Wait()
		}
		pool.Cancel = sync.OnceFunc(func() { close(workChan) })
	}

	return pool
}
