package client

import (
	"context"
	"errors"
	"sync"
	"time"

	"omtask/external"
)

type Client interface {
	Run()
	Process(ctx context.Context, batch external.Batch) ([]external.Item, error)
	Stop()
}

type client struct {
	external external.Service

	count           uint64
	mux             sync.Mutex
	stopChan        chan struct{}
	isLimitUpdating bool
	isRunning       bool
}

func New(external external.Service) Client {
	return &client{
		external:        external,
		count:           0,
		mux:             sync.Mutex{},
		stopChan:        make(chan struct{}),
		isLimitUpdating: false,
		isRunning:       false,
	}
}

func (c *client) Run() {
	n, dur := c.external.GetLimits()
	c.count = n
	c.isRunning = true

	c.watchLimit(dur)
}

func (c *client) watchLimit(startDuration time.Duration) {
	go func() {
		ticker := time.NewTicker(startDuration)
		for {
			select {
			case <-ticker.C:
				c.updateLimit(ticker)
			case <-c.stopChan:
				return
			}
		}
	}()
}

func (c *client) updateLimit(ticker *time.Ticker) {
	c.isLimitUpdating = true
	c.mux.Lock()
	defer c.mux.Unlock()

	n, dur := c.external.GetLimits()
	ticker.Reset(dur)
	c.count = n
	c.isLimitUpdating = false
}

var ErrClientStopped = errors.New("client stopped")

func (c *client) Process(ctx context.Context, batch external.Batch) (
	[]external.Item,
	error,
) {
	for c.isLimitUpdating {
		select {
		case <-ctx.Done():
			return batch, ctx.Err()
		default:
		}
	}

	if !c.isRunning {
		return batch, ErrClientStopped
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
		return nil, err
	}

	return unhandled, nil
}

func (c *client) prepare(batch external.Batch) (external.Batch, []external.Item) {
	var forHandle []external.Item
	var unhandled []external.Item
	if c.count >= uint64(len(batch)) {
		forHandle = make([]external.Item, len(batch))
	} else {
		forHandle = make([]external.Item, c.count)
		unhandled = make([]external.Item, uint64(len(batch))-c.count)
	}

	copy(forHandle, batch)
	unhandled = batch[len(forHandle):]
	c.count -= uint64(len(forHandle))

	return forHandle, unhandled
}

func (c *client) Stop() {
	close(c.stopChan)
	c.isRunning = false
}
