// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package external

import (
	"context"
	"sync"
	"time"
)

// Ensure, that ServiceMock does implement Service.
// If this is not the case, regenerate this file with moq.
var _ Service = &ServiceMock{}

// ServiceMock is a mock implementation of Service.
//
//	func TestSomethingThatUsesService(t *testing.T) {
//
//		// make and configure a mocked Service
//		mockedService := &ServiceMock{
//			GetLimitsFunc: func() (uint64, time.Duration) {
//				panic("mock out the GetLimits method")
//			},
//			ProcessFunc: func(ctx context.Context, batch Batch) error {
//				panic("mock out the Process method")
//			},
//		}
//
//		// use mockedService in code that requires Service
//		// and then make assertions.
//
//	}
type ServiceMock struct {
	// GetLimitsFunc mocks the GetLimits method.
	GetLimitsFunc func() (uint64, time.Duration)

	// ProcessFunc mocks the Process method.
	ProcessFunc func(ctx context.Context, batch Batch) error

	// calls tracks calls to the methods.
	calls struct {
		// GetLimits holds details about calls to the GetLimits method.
		GetLimits []struct {
		}
		// Process holds details about calls to the Process method.
		Process []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Batch is the batch argument value.
			Batch Batch
		}
	}
	lockGetLimits sync.RWMutex
	lockProcess   sync.RWMutex
}

// GetLimits calls GetLimitsFunc.
func (mock *ServiceMock) GetLimits() (uint64, time.Duration) {
	if mock.GetLimitsFunc == nil {
		panic("ServiceMock.GetLimitsFunc: method is nil but Service.GetLimits was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetLimits.Lock()
	mock.calls.GetLimits = append(mock.calls.GetLimits, callInfo)
	mock.lockGetLimits.Unlock()
	return mock.GetLimitsFunc()
}

// GetLimitsCalls gets all the calls that were made to GetLimits.
// Check the length with:
//
//	len(mockedService.GetLimitsCalls())
func (mock *ServiceMock) GetLimitsCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetLimits.RLock()
	calls = mock.calls.GetLimits
	mock.lockGetLimits.RUnlock()
	return calls
}

// Process calls ProcessFunc.
func (mock *ServiceMock) Process(ctx context.Context, batch Batch) error {
	if mock.ProcessFunc == nil {
		panic("ServiceMock.ProcessFunc: method is nil but Service.Process was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		Batch Batch
	}{
		Ctx:   ctx,
		Batch: batch,
	}
	mock.lockProcess.Lock()
	mock.calls.Process = append(mock.calls.Process, callInfo)
	mock.lockProcess.Unlock()
	return mock.ProcessFunc(ctx, batch)
}

// ProcessCalls gets all the calls that were made to Process.
// Check the length with:
//
//	len(mockedService.ProcessCalls())
func (mock *ServiceMock) ProcessCalls() []struct {
	Ctx   context.Context
	Batch Batch
} {
	var calls []struct {
		Ctx   context.Context
		Batch Batch
	}
	mock.lockProcess.RLock()
	calls = mock.calls.Process
	mock.lockProcess.RUnlock()
	return calls
}