package rate

import (
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestRPSRateLimiterExecutionDuration(t *testing.T) {
	const (
		roundsPerSecond = 16
		nbTasks         = 128
	)
	logger := zaptest.NewLogger(t)

	task := func() error {
		time.Sleep(1 * time.Second)
		if rand.Int32()%3 == 0 { // Error out one third of the responses.
			return fmt.Errorf("something went wrong")
		}
		return nil
	}

	start := time.Now()
	limiter := NewRPSLimiter(roundsPerSecond).
		MakeBuffered(nbTasks).
		WithLogger(logger.Sugar())
	limiter.Start()

	errorChans := make([]<-chan error, 0, nbTasks)
	for range nbTasks {
		errorChan, enqueued := limiter.TryEnqueue(task)
		require.NotNil(t, errorChan)
		require.True(t, enqueued)
		errorChans = append(errorChans, errorChan)
	}

	t.Logf("enqueued all %d tasks", nbTasks)

	require.NoError(t, limiter.Close())

	var errCount, nilCount int
	for _, errorCh := range errorChans {
		err := <-errorCh
		// Each chan must be closed after the err value is read.
		_, ok := <-errorCh
		require.False(t, ok)
		if err == nil {
			nilCount++
		} else {
			errCount++
		}
	}

	elapsed := time.Since(start)

	t.Logf("rate limited %d tasks in %s at %d rps", nbTasks, elapsed, roundsPerSecond)

	require.Equal(t, nbTasks, errCount+nilCount)
	require.GreaterOrEqual(t, elapsed, time.Duration(nbTasks/roundsPerSecond))
}
