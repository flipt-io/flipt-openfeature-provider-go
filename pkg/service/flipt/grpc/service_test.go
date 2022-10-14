package service_grpc

import (
	"context"
	"errors"
	"testing"

	of "github.com/open-feature/go-sdk/pkg/openfeature"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"go.flipt.io/flipt-grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		opts     []Option
		expected Service
	}{
		{
			name: "default",
			expected: Service{
				host: "localhost",
				port: 9000,
			},
		},
		{
			name: "with host",
			opts: []Option{WithHost("foo")},
			expected: Service{
				host: "foo",
				port: 9000,
			},
		},
		{
			name: "with port",
			opts: []Option{WithPort(1234)},
			expected: Service{
				host: "localhost",
				port: 1234,
			},
		},
		{
			name: "with certificate path",
			opts: []Option{WithCertificatePath("foo")},
			expected: Service{
				host:            "localhost",
				port:            9000,
				certificatePath: "foo",
			},
		},
		{
			name: "with socket path",
			opts: []Option{WithSocketPath("bar")},
			expected: Service{
				host:       "localhost",
				port:       9000,
				socketPath: "foo",
			},
		},
	}

	//nolint (copylocks)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.opts...)

			assert.NotNil(t, s)
			assert.Equal(t, tt.expected.host, s.host)
			assert.Equal(t, tt.expected.port, s.port)
		})
	}
}

func TestGRPCToOpenFeatureError(t *testing.T) {
	tests := []struct {
		name        string
		grpcStatus  *status.Status
		expectedErr of.ResolutionError
	}{
		{
			name:        "invalid argument",
			grpcStatus:  status.New(codes.InvalidArgument, "invalid argument"),
			expectedErr: of.NewInvalidContextResolutionError("invalid argument"),
		},
		{
			name:        "not found",
			grpcStatus:  status.New(codes.NotFound, "not found"),
			expectedErr: of.NewFlagNotFoundResolutionError("not found"),
		},
		{
			name:        "unavailable",
			grpcStatus:  status.New(codes.Unavailable, "unavailable"),
			expectedErr: of.NewProviderNotReadyResolutionError("unavailable"),
		},
		{
			name:        "unknown",
			grpcStatus:  status.New(codes.Unknown, "unknown"),
			expectedErr: of.NewGeneralResolutionError("unknown"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := grpcToOpenFeatureError(*tt.grpcStatus)

			assert.EqualError(t, err, tt.expectedErr.Error())
		})
	}
}

func TestGetFlag(t *testing.T) {
	tests := []struct {
		name        string
		response    *flipt.Flag
		reqErr      error
		expectedErr error
		expected    *flipt.Flag
	}{
		{
			name: "success",
			response: &flipt.Flag{
				Key: "foo",
			},
			expected: &flipt.Flag{
				Key: "foo",
			},
		},
		{
			name:        "flag not found",
			reqErr:      status.New(codes.NotFound, "not found").Err(),
			expectedErr: of.NewFlagNotFoundResolutionError("not found"),
		},
		{
			name:        "error",
			reqErr:      errors.New("boom"),
			expectedErr: errors.New("getting flag \"foo\" boom"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := newMockGrpcClient(t)

			mockClient.On("GetFlag", mock.Anything, mock.Anything).Return(tt.response, tt.reqErr)

			s := &Service{
				client: mockClient,
			}

			actual, err := s.GetFlag(context.Background(), "foo")
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name        string
		response    *flipt.EvaluationResponse
		reqErr      error
		expectedErr error
		expected    *flipt.EvaluationResponse
	}{
		{
			name: "success",
			response: &flipt.EvaluationResponse{
				FlagKey: "foo",
				Match:   true,
			},
			expected: &flipt.EvaluationResponse{
				FlagKey: "foo",
				Match:   true,
			},
		},
		{
			name:        "flag not found",
			reqErr:      status.New(codes.NotFound, "not found").Err(),
			expectedErr: of.NewFlagNotFoundResolutionError("not found"),
		},
		{
			name:        "error",
			reqErr:      errors.New("boom"),
			expectedErr: errors.New("evaluating flag \"foo\" boom"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := newMockGrpcClient(t)

			mockClient.On("Evaluate", mock.Anything, mock.Anything, mock.Anything).Return(tt.response, tt.reqErr)

			s := &Service{
				client: mockClient,
			}

			evalCtx := map[string]interface{}{
				"foo":           "bar",
				of.TargetingKey: "12345",
			}

			actual, err := s.Evaluate(context.Background(), "foo", evalCtx)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}

func TestEvaluateInvalidContext(t *testing.T) {
	s := &Service{}

	_, err := s.Evaluate(context.Background(), "foo", nil)
	assert.EqualError(t, err, of.NewInvalidContextResolutionError("evalCtx is nil").Error())

	_, err = s.Evaluate(context.Background(), "foo", map[string]interface{}{})
	assert.EqualError(t, err, of.NewTargetingKeyMissingResolutionError("targetingKey is missing").Error())
}

func TestLoadTLSCredentials(t *testing.T) {
	tests := []struct {
		name           string
		certificate    string
		expectedErrMsg string
	}{
		{
			name:           "no certificate",
			certificate:    "foo",
			expectedErrMsg: "failed to load certificate: open foo: no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadTLSCredentials(tt.certificate)

			if tt.expectedErrMsg != "" {
				assert.EqualError(t, err, tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
