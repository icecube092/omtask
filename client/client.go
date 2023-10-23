package client

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"omtask/external"
)

// ErrClientStopped returns if client stopped
var ErrClientStopped = errors.New("client stopped")

type Client interface {
	Run()
	// Process handle batch and returns part of batch that cannot be handled now
	Process(ctx context.Context, batch external.Batch) ([]external.Item, error)
	Stop()
}

type client struct {
	external external.Service

	count           uint64
	mux             sync.Mutex
	stopChan        chan struct{}
	isLimitUpdating bool           // when limits on updating, set to true for lock request
	isRunning       bool           // specify is client running
	workerUpWg      sync.WaitGroup // wait that all workers up
}

func New(external external.Service) Client {
	return &client{
		external:        external,
		count:           0,
		mux:             sync.Mutex{},
		stopChan:        make(chan struct{}),
		isLimitUpdating: false,
		isRunning:       false,
		workerUpWg:      sync.WaitGroup{},
	}
}

func (c *client) Run() {
	startDuration := c.updateLimit()

	c.workerUpWg.Add(1)
	c.watchLimit(startDuration)
	c.workerUpWg.Wait()

	c.isRunning = true
}

func (c *client) watchLimit(startDuration time.Duration) {
	go func() {
		c.workerUpWg.Done()
		ticker := time.NewTicker(startDuration)
		for {
			select {
			case <-ticker.C:
				ticker.Reset(c.updateLimit())
			case <-c.stopChan:
				return
			}
		}
	}()
}

func (c *client) updateLimit() time.Duration {
	c.isLimitUpdating = true
	c.mux.Lock()
	defer c.mux.Unlock()
	defer func() {
		c.isLimitUpdating = false
	}()

	n, dur := c.external.GetLimits()
	c.count = n

	return dur
}

func (c *client) Process(ctx context.Context, batch external.Batch) (
	[]external.Item,
	error,
) {
	for c.isLimitUpdating {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("wait limit updating: %w", ctx.Err())
		default:
		}
	}

	if !c.isRunning {
		return nil, ErrClientStopped
	}

	c.mux.Lock()
	if c.count <= 0 {
		c.mux.Unlock()
		return batch, nil
	}

	forHandle, unhandled := c.prepare(batch)

	c.mux.Unlock()

	if err := c.external.Process(ctx, forHandle); err != nil {
		if errors.Is(err, external.ErrBlocked) {
			return batch, nil
		}
		return nil, fmt.Errorf("external.Process: %w", err)
	}

	return unhandled, nil
}

func (c *client) prepare(batch external.Batch) (external.Batch, []external.Item) {
	size := uint64(len(batch))
	if c.count < size {
		size = c.count
	}

	forHandle := batch[:size]
	unhandled := batch[len(forHandle):]
	c.count -= uint64(len(forHandle))

	return forHandle, unhandled
}

func (c *client) Stop() {
	close(c.stopChan)
	c.isRunning = false
}
