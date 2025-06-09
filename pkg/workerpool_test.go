package workerpool_test

import (
	"sync"
	"testing"

	workerpool "github.com/onemoreslacker/workerpool/pkg"
	"github.com/stretchr/testify/require"
)

func TestWorkerPool_AddRemove(t *testing.T) {
	wp := workerpool.New()
	require.Zero(t, wp.Running())

	const numWorkers = 1000

	for range numWorkers {
		require.NoError(t, wp.AddWorker())
	}
	require.Equal(t, wp.Running(), numWorkers)

	for range numWorkers {
		require.NoError(t, wp.RemoveWorker())
	}
}

func TestWorkerPool_Close(t *testing.T) {
	wp := workerpool.New()

	require.NoError(t, wp.AddWorker())
	require.NoError(t, wp.Submit("some job"))

	wp.Close()

	require.True(t, wp.IsClosed())

	require.ErrorIs(t, wp.AddWorker(), workerpool.ErrPoolClosed)
	require.ErrorIs(t, wp.RemoveWorker(), workerpool.ErrPoolClosed)
	require.ErrorIs(t, wp.Submit("some job"), workerpool.ErrPoolClosed)

	wp.Close()
}

func TestWorkerPool_ConcurrentAddRemove(t *testing.T) {
	wp := workerpool.New()
	defer wp.Close()

	const numWorkers = 1000
	wg := sync.WaitGroup{}

	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			require.NoError(t, wp.AddWorker())
		}()
	}

	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = wp.RemoveWorker()
		}()
	}

	wg.Wait()

	require.Greater(t, wp.Running(), -1)
}
