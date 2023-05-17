package servicegrpc

import (
	"context"
	"testing"

	of "github.com/open-feature/go-sdk/pkg/openfeature"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"

	offlipt "go.flipt.io/flipt-openfeature-provider/pkg/service/flipt"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	reqID    = "987654321"
	entityID = "123456789"
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
				address: "localhost:9000",
			},
		},
		{
			name: "with host",
			opts: []Option{WithAddress("foo:9000")},
			expected: Service{
				address: "foo:9000",
			},
		},
		{
			name: "with certificate path",
			opts: []Option{WithCertificatePath("foo")},
			expected: Service{
				address:         "localhost:9000",
				certificatePath: "foo",
			},
		},
		{
			name: "with unix address path",
			opts: []Option{WithAddress("unix://bar")},
			expected: Service{
				address: "passthrough:///unix://bar",
			},
		},
	}

	//nolint (copylocks)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.opts...)

			assert.NotNil(t, s)
			assert.Equal(t, tt.expected.address, s.address)
		})
	}
}

func TestGetFlag(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedErr error
		expected    *flipt.Flag
	}{
		{
			name: "success",
			expected: &flipt.Flag{
				Key:          "foo",
				NamespaceKey: "foo-namespace",
			},
		},
		{
			name:        "flag not found",
			err:         status.Error(codes.NotFound, `flag "foo" not found`),
			expectedErr: of.NewFlagNotFoundResolutionError(`flag "foo" not found`),
		},
		{
			name:        "other error",
			err:         status.Error(codes.Internal, "internal error"),
			expectedErr: of.NewGeneralResolutionError("internal error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := offlipt.NewMockClient(t)

			mockClient.On("GetFlag", mock.Anything, &flipt.GetFlagRequest{
				Key:          "foo",
				NamespaceKey: "foo-namespace",
			}).Return(tt.expected, tt.err)

			s := &Service{
				client: mockClient,
			}

			actual, err := s.GetFlag(context.Background(), "foo-namespace", "foo")
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
		err         error
		expectedErr error
		expected    *flipt.EvaluationResponse
	}{
		{
			name: "success",
			expected: &flipt.EvaluationResponse{
				FlagKey:    "foo",
				Match:      true,
				SegmentKey: "foo-segment",
			},
		},
		{
			name:        "flag not found",
			err:         status.Error(codes.NotFound, `flag "foo" not found`),
			expectedErr: of.NewFlagNotFoundResolutionError(`flag "foo" not found`),
		},
		{
			name:        "other error",
			err:         status.Error(codes.Internal, "internal error"),
			expectedErr: of.NewGeneralResolutionError("internal error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := offlipt.NewMockClient(t)

			mockClient.EXPECT().Evaluate(mock.Anything, &flipt.EvaluationRequest{
				FlagKey:      "foo",
				NamespaceKey: "foo-namespace",
				RequestId:    reqID,
				EntityId:     entityID,
				Context: map[string]string{
					"requestID":    reqID,
					"targetingKey": entityID,
				},
			}).Return(tt.expected, tt.err)

			s := &Service{
				client: mockClient,
			}

			evalCtx := map[string]interface{}{
				"requestID":     reqID,
				of.TargetingKey: entityID,
			}

			actual, err := s.Evaluate(context.Background(), "foo-namespace", "foo", evalCtx)
			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.FlagKey, actual.FlagKey)
				assert.Equal(t, tt.expected.Match, actual.Match)
				assert.Equal(t, tt.expected.SegmentKey, actual.SegmentKey)
			}
		})
	}
}

func TestEvaluateInvalidContext(t *testing.T) {
	s := &Service{}

	_, err := s.Evaluate(context.Background(), "foo-namespace", "foo", nil)
	assert.EqualError(t, err, of.NewInvalidContextResolutionError("evalCtx is nil").Error())

	_, err = s.Evaluate(context.Background(), "foo-namespace", "foo", map[string]interface{}{})
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
