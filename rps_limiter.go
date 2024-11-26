package rate

import (
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

type Limiter interface {
	io.Closer
	Start()
	TryEnqueue(func() error) (<-chan error, bool)
}

type RPSLimiter struct {
	eg      errgroup.Group
	tick    <-chan time.Time
	tasks   chan func()
	logger  Logger
	running atomic.Bool
	closed  atomic.Bool
}

func NewRPSLimiter(rps uint) *RPSLimiter {
	rl := &RPSLimiter{
		tick:   time.Tick(time.Second / time.Duration(rps)),
		tasks:  make(chan func()),
		logger: standardLogger,
	}
	return rl
}

func (l *RPSLimiter) WithLogger(logger Logger) *RPSLimiter {
	if l.isRunningOrClosed() {
		l.logger.Errorf("can't update limiter's logger while it's running or closed")
		return l
	}
	l.logger = logger
	return l
}

func (l *RPSLimiter) MakeBuffered(n int) *RPSLimiter {
	if l.isRunningOrClosed() {
		l.logger.Errorf("can't make the limiter buffered while it's running or closed")
		return l
	}
	l.tasks = make(chan func(), n)
	return l
}

func (l *RPSLimiter) Start() {
	if l.isRunningOrClosed() {
		return
	}
	l.running.Store(true)
	go func() {
		l.executeOnTick()
		l.running.Store(false)
	}()
}

func (l *RPSLimiter) isRunningOrClosed() bool {
	return l.running.Load() || l.closed.Load()
}

// executeOnTick executes a task in a goroutine everytime the internal ticker
// sends a signal when a new execution is allowed.
func (l *RPSLimiter) executeOnTick() {
	var count int
	for task := range l.tasks {
		// We got a task, now wait for a tick to execute it.
		<-l.tick
		// Execute the task in a goroutine, in case it takes more time than
		// the current rps / tick every delay, that way we do not block
		// other tasks' execution.
		l.eg.Go(func() error {
			start := time.Now()
			task()
			l.logger.Debugf("task executed in %s", time.Since(start))
			return nil
		})
		count++
	}
	l.logger.Infof("dequeued all %d tasks", count)
}

// Close closes the [RPSLimiter]. It waits for all the currently queued or
// executing task to be done.
//
// Further [RPSLimiter.TryEnqueue] calls will end up not enqueuing any task
// once [RPSLimiter.Close] has been called.
func (l *RPSLimiter) Close() error {
	// First, mark the limiter as closed to make sure no more tasks are being
	// queued. It will also signal to the executing loop to finish.
	if !l.closed.CompareAndSwap(false, true) {
		return nil
	}
	// Wait for all tasks that are currently executing to finish and
	// produce their result.
	if err := l.eg.Wait(); err != nil {
		return fmt.Errorf("unable to close errgroup, err: %w", err)
	}
	close(l.tasks)
	return nil
}

// TryEnqueue tries to push a task in the [RPSLimiter] queue, that will be
// executed once the current rate limit allows it.
//
// It returns a chan that will receive the execution error once it is ready and
// a bool indicating whether or not the task has been queued or not. The only
// case where this can happen is when [RPSLimiter.Close] has been called.
func (l *RPSLimiter) TryEnqueue(task func() error) (<-chan error, bool) {
	if l.closed.Load() || !l.running.Load() {
		return nil, false
	}
	// The returned chan is buffered for a single result, that way if no
	// goroutine is ready to receive on the calling side, the rate limiter can
	// still send the execution's result without waiting.
	resultCh := make(chan error, 1)
	l.tasks <- func() {
		resultCh <- task()
		close(resultCh)
	}
	return resultCh, true
}

var (
	_ Limiter = &RPSLimiter{}
)
