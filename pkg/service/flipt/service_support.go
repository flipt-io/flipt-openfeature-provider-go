// Code generated by mockery. DO NOT EDIT.

package flipt

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	rpcflipt "go.flipt.io/flipt/rpc/flipt"
)

// MockClient is an autogenerated mock type for the Client type
type MockClient struct {
	mock.Mock
}

type MockClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockClient) EXPECT() *MockClient_Expecter {
	return &MockClient_Expecter{mock: &_m.Mock}
}

// Evaluate provides a mock function with given fields: ctx, v
func (_m *MockClient) Evaluate(ctx context.Context, v *rpcflipt.EvaluationRequest) (*rpcflipt.EvaluationResponse, error) {
	ret := _m.Called(ctx, v)

	var r0 *rpcflipt.EvaluationResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *rpcflipt.EvaluationRequest) (*rpcflipt.EvaluationResponse, error)); ok {
		return rf(ctx, v)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *rpcflipt.EvaluationRequest) *rpcflipt.EvaluationResponse); ok {
		r0 = rf(ctx, v)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*rpcflipt.EvaluationResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *rpcflipt.EvaluationRequest) error); ok {
		r1 = rf(ctx, v)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClient_Evaluate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Evaluate'
type MockClient_Evaluate_Call struct {
	*mock.Call
}

// Evaluate is a helper method to define mock.On call
//   - ctx context.Context
//   - v *rpcflipt.EvaluationRequest
func (_e *MockClient_Expecter) Evaluate(ctx interface{}, v interface{}) *MockClient_Evaluate_Call {
	return &MockClient_Evaluate_Call{Call: _e.mock.On("Evaluate", ctx, v)}
}

func (_c *MockClient_Evaluate_Call) Run(run func(ctx context.Context, v *rpcflipt.EvaluationRequest)) *MockClient_Evaluate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*rpcflipt.EvaluationRequest))
	})
	return _c
}

func (_c *MockClient_Evaluate_Call) Return(_a0 *rpcflipt.EvaluationResponse, _a1 error) *MockClient_Evaluate_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClient_Evaluate_Call) RunAndReturn(run func(context.Context, *rpcflipt.EvaluationRequest) (*rpcflipt.EvaluationResponse, error)) *MockClient_Evaluate_Call {
	_c.Call.Return(run)
	return _c
}

// GetFlag provides a mock function with given fields: ctx, c
func (_m *MockClient) GetFlag(ctx context.Context, c *rpcflipt.GetFlagRequest) (*rpcflipt.Flag, error) {
	ret := _m.Called(ctx, c)

	var r0 *rpcflipt.Flag
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *rpcflipt.GetFlagRequest) (*rpcflipt.Flag, error)); ok {
		return rf(ctx, c)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *rpcflipt.GetFlagRequest) *rpcflipt.Flag); ok {
		r0 = rf(ctx, c)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*rpcflipt.Flag)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *rpcflipt.GetFlagRequest) error); ok {
		r1 = rf(ctx, c)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClient_GetFlag_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetFlag'
type MockClient_GetFlag_Call struct {
	*mock.Call
}

// GetFlag is a helper method to define mock.On call
//   - ctx context.Context
//   - c *rpcflipt.GetFlagRequest
func (_e *MockClient_Expecter) GetFlag(ctx interface{}, c interface{}) *MockClient_GetFlag_Call {
	return &MockClient_GetFlag_Call{Call: _e.mock.On("GetFlag", ctx, c)}
}

func (_c *MockClient_GetFlag_Call) Run(run func(ctx context.Context, c *rpcflipt.GetFlagRequest)) *MockClient_GetFlag_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*rpcflipt.GetFlagRequest))
	})
	return _c
}

func (_c *MockClient_GetFlag_Call) Return(_a0 *rpcflipt.Flag, _a1 error) *MockClient_GetFlag_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClient_GetFlag_Call) RunAndReturn(run func(context.Context, *rpcflipt.GetFlagRequest) (*rpcflipt.Flag, error)) *MockClient_GetFlag_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewMockClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockClient creates a new instance of MockClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockClient(t mockConstructorTestingTNewMockClient) *MockClient {
	mock := &MockClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}