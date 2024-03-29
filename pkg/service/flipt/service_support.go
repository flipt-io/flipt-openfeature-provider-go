// Code generated by mockery. DO NOT EDIT.

package flipt

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	evaluation "go.flipt.io/flipt/rpc/flipt/evaluation"

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

// Boolean provides a mock function with given fields: ctx, v
func (_m *MockClient) Boolean(ctx context.Context, v *evaluation.EvaluationRequest) (*evaluation.BooleanEvaluationResponse, error) {
	ret := _m.Called(ctx, v)

	var r0 *evaluation.BooleanEvaluationResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *evaluation.EvaluationRequest) (*evaluation.BooleanEvaluationResponse, error)); ok {
		return rf(ctx, v)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *evaluation.EvaluationRequest) *evaluation.BooleanEvaluationResponse); ok {
		r0 = rf(ctx, v)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*evaluation.BooleanEvaluationResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *evaluation.EvaluationRequest) error); ok {
		r1 = rf(ctx, v)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClient_Boolean_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Boolean'
type MockClient_Boolean_Call struct {
	*mock.Call
}

// Boolean is a helper method to define mock.On call
//   - ctx context.Context
//   - v *evaluation.EvaluationRequest
func (_e *MockClient_Expecter) Boolean(ctx interface{}, v interface{}) *MockClient_Boolean_Call {
	return &MockClient_Boolean_Call{Call: _e.mock.On("Boolean", ctx, v)}
}

func (_c *MockClient_Boolean_Call) Run(run func(ctx context.Context, v *evaluation.EvaluationRequest)) *MockClient_Boolean_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*evaluation.EvaluationRequest))
	})
	return _c
}

func (_c *MockClient_Boolean_Call) Return(_a0 *evaluation.BooleanEvaluationResponse, _a1 error) *MockClient_Boolean_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClient_Boolean_Call) RunAndReturn(run func(context.Context, *evaluation.EvaluationRequest) (*evaluation.BooleanEvaluationResponse, error)) *MockClient_Boolean_Call {
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

// Variant provides a mock function with given fields: ctx, v
func (_m *MockClient) Variant(ctx context.Context, v *evaluation.EvaluationRequest) (*evaluation.VariantEvaluationResponse, error) {
	ret := _m.Called(ctx, v)

	var r0 *evaluation.VariantEvaluationResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *evaluation.EvaluationRequest) (*evaluation.VariantEvaluationResponse, error)); ok {
		return rf(ctx, v)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *evaluation.EvaluationRequest) *evaluation.VariantEvaluationResponse); ok {
		r0 = rf(ctx, v)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*evaluation.VariantEvaluationResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *evaluation.EvaluationRequest) error); ok {
		r1 = rf(ctx, v)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClient_Variant_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Variant'
type MockClient_Variant_Call struct {
	*mock.Call
}

// Variant is a helper method to define mock.On call
//   - ctx context.Context
//   - v *evaluation.EvaluationRequest
func (_e *MockClient_Expecter) Variant(ctx interface{}, v interface{}) *MockClient_Variant_Call {
	return &MockClient_Variant_Call{Call: _e.mock.On("Variant", ctx, v)}
}

func (_c *MockClient_Variant_Call) Run(run func(ctx context.Context, v *evaluation.EvaluationRequest)) *MockClient_Variant_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*evaluation.EvaluationRequest))
	})
	return _c
}

func (_c *MockClient_Variant_Call) Return(_a0 *evaluation.VariantEvaluationResponse, _a1 error) *MockClient_Variant_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClient_Variant_Call) RunAndReturn(run func(context.Context, *evaluation.EvaluationRequest) (*evaluation.VariantEvaluationResponse, error)) *MockClient_Variant_Call {
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
