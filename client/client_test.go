package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"omtask/external"
)

func TestClient(t *testing.T) {
	const sleepDuration = time.Millisecond * 100
	const limit = 10

	ctx := context.Background()
	ext := &external.ServiceMock{}

	ext.GetLimitsFunc = func() (uint64, time.Duration) {
		return limit, sleepDuration
	}
	ext.ProcessFunc = func(ctx context.Context, batch external.Batch) error {
		return nil
	}

	const size = limit + 5
	batch := make(external.Batch, size)

	c := New(ext)

	c.Run()

	unhandled, err := c.Process(ctx, batch)
	require.NoError(t, err)
	require.Len(t, unhandled, size-limit)
	require.Equal(t, ext.ProcessCalls()[0].Batch, batch[:limit])

	unhandled, err = c.Process(ctx, batch)
	require.NoError(t, err)
	require.Len(t, unhandled, size)

	time.Sleep(sleepDuration)

	unhandled, err = c.Process(ctx, batch)
	require.NoError(t, err)
	require.Len(t, unhandled, size-limit)
	require.Equal(t, ext.ProcessCalls()[1].Batch, batch[:limit])

	require.Equal(t, len(ext.ProcessCalls()), 2)
}

func TestClient_blocked(t *testing.T) {
	const sleepDuration = time.Millisecond * 100
	const limit = 10

	ctx := context.Background()
	ext := &external.ServiceMock{}

	ext.GetLimitsFunc = func() (uint64, time.Duration) {
		return limit, sleepDuration
	}
	ext.ProcessFunc = func(ctx context.Context, batch external.Batch) error {
		return external.ErrBlocked
	}

	const size = limit + 5
	batch := make(external.Batch, size)

	c := New(ext)

	c.Run()

	unhandled, err := c.Process(ctx, batch)
	require.NoError(t, err)
	require.Len(t, unhandled, size)
	require.Equal(t, ext.ProcessCalls()[0].Batch, batch[:limit])
}

func TestClient_stop(t *testing.T) {
	const sleepDuration = time.Millisecond * 100
	const limit = 10

	ctx := context.Background()
	ext := &external.ServiceMock{}

	ext.GetLimitsFunc = func() (uint64, time.Duration) {
		return limit, sleepDuration
	}
	ext.ProcessFunc = func(ctx context.Context, batch external.Batch) error {
		return nil
	}

	const size = limit + 5
	batch := make(external.Batch, size)

	c := New(ext)

	c.Run()
	c.Stop()

	unhandled, err := c.Process(ctx, batch)
	require.ErrorIs(t, err, ErrClientStopped)
	require.Len(t, unhandled, size)

	require.Equal(t, len(ext.ProcessCalls()), 0)
}

func TestClient_waitUpdateLimit(t *testing.T) {
	const sleepDuration = time.Millisecond * 10
	const limit = 10

	ctx := context.Background()
	ext := &external.ServiceMock{}

	ext.GetLimitsFunc = func() (uint64, time.Duration) {
		time.Sleep(sleepDuration)
		return limit, sleepDuration
	}
	ext.ProcessFunc = func(ctx context.Context, batch external.Batch) error {
		return nil
	}

	const size = limit + 5
	batch := make(external.Batch, size)

	c := New(ext)

	c.Run()

	time.Sleep(sleepDuration)
	unhandled, err := c.Process(ctx, batch)
	require.NoError(t, err)
	require.Len(t, unhandled, size-limit)
	require.Equal(t, len(ext.ProcessCalls()), 1)
}

func TestClient_waitUpdateLimit_timeout(t *testing.T) {
	const sleepDuration = time.Millisecond * 10
	const limit = 10

	ctx := context.Background()
	ext := &external.ServiceMock{}

	ext.GetLimitsFunc = func() (uint64, time.Duration) {
		time.Sleep(sleepDuration * 5)
		return limit, sleepDuration
	}
	ext.ProcessFunc = func(ctx context.Context, batch external.Batch) error {
		return nil
	}

	const size = limit + 5
	batch := make(external.Batch, size)

	c := New(ext)

	c.Run()

	time.Sleep(sleepDuration)
	ctx, cancel := context.WithTimeout(ctx, sleepDuration/10)
	defer cancel()
	unhandled, err := c.Process(ctx, batch)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Len(t, unhandled, size)

	require.Equal(t, len(ext.ProcessCalls()), 0)
}
