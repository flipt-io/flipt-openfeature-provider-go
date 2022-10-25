package servicehttp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	of "github.com/open-feature/go-sdk/pkg/openfeature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.flipt.io/flipt-grpc"
)

func TestProtocol(t *testing.T) {
	tests := []struct {
		name     string
		protocol Protocol
	}{
		{
			name:     "http",
			protocol: HTTP,
		},
		{
			name:     "https",
			protocol: HTTPS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.name, tt.protocol.String())
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		opts     []Option
		expected Service
	}{
		{
			name: "default",
			expected: Service{
				host:     "localhost",
				port:     8080,
				protocol: HTTP,
			},
		},
		{
			name: "with host",
			opts: []Option{WithHost("foo")},
			expected: Service{
				host:     "foo",
				port:     8080,
				protocol: HTTP,
			},
		},
		{
			name: "with port",
			opts: []Option{WithPort(1234)},
			expected: Service{
				host:     "localhost",
				port:     1234,
				protocol: HTTP,
			},
		},
		{
			name: "with https",
			opts: []Option{WithHTTPS()},
			expected: Service{
				host:     "localhost",
				port:     8080,
				protocol: HTTPS,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.opts...)

			assert.NotNil(t, s)
			assert.Equal(t, tt.expected.host, s.host)
			assert.Equal(t, tt.expected.port, s.port)
			assert.Equal(t, tt.expected.protocol, s.protocol)
		})
	}
}

func TestGetFlag(t *testing.T) {
	tests := []struct {
		name         string
		responseBody []byte
		responseCode int
		reqErr       error
		expectedErr  error
		expected     *flipt.Flag
	}{
		{
			name:         "success",
			responseBody: []byte(`{"key":"foo","name":"Flag Name","description":"Flag Description","enabled":true}`),
			responseCode: http.StatusOK,
			expected: &flipt.Flag{
				Key:         "foo",
				Name:        "Flag Name",
				Description: "Flag Description",
				Enabled:     true,
			},
		},
		{
			name:         "flag not found",
			responseBody: []byte(`{"error":"flag not found","code":5}`),
			responseCode: http.StatusNotFound,
			expectedErr:  of.NewFlagNotFoundResolutionError(`flag "foo" not found`),
		},
		{
			name:         "invalid json",
			responseBody: []byte(`{"invalid}`),
			responseCode: http.StatusOK,
			expectedErr:  errors.New("unmarshalling response body"),
		},
		{
			name:        "request error",
			reqErr:      errors.New("request error"),
			expectedErr: errors.New("making request request error"),
		},
		{
			name:         "unexpected status code",
			responseCode: http.StatusInternalServerError,
			expectedErr:  errors.New("getting flag: status=500 "),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := newMockHttpClient(t)

			mockClient.EXPECT().Do(mock.Anything).Run(func(req *http.Request) {
				assert.Equal(t, "GET", req.Method)
				assert.Equal(t, "/api/v1/flags/foo", req.URL.Path)
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
				assert.Equal(t, "application/json", req.Header.Get("Accept"))
			}).Return(&http.Response{
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				StatusCode: tt.responseCode,
				Body:       io.NopCloser(bytes.NewReader(tt.responseBody)),
			}, tt.reqErr)

			s := &Service{
				client: mockClient,
			}

			actual, err := s.GetFlag(context.Background(), "foo")
			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Key, actual.Key)
				assert.Equal(t, tt.expected.Name, actual.Name)
				assert.Equal(t, tt.expected.Description, actual.Description)
				assert.Equal(t, tt.expected.Enabled, actual.Enabled)
			}
		})
	}
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name         string
		responseBody []byte
		responseCode int
		reqErr       error
		expectedErr  error
		expected     *flipt.EvaluationResponse
	}{
		{
			name:         "success",
			responseBody: []byte(`{"flag_key":"foo","match":true}`),
			responseCode: http.StatusOK,
			expected: &flipt.EvaluationResponse{
				FlagKey: "foo",
				Match:   true,
			},
		},
		{
			name:         "flag not found",
			responseBody: []byte(`{"error":"flag not found","code":5}`),
			responseCode: http.StatusNotFound,
			expectedErr:  of.NewFlagNotFoundResolutionError(`flag "foo" not found`),
		},
		{
			name:         "invalid json",
			responseBody: []byte(`{"invalid}`),
			responseCode: http.StatusOK,
			expectedErr:  errors.New("unmarshalling response body"),
		},
		{
			name:        "request error",
			reqErr:      errors.New("request error"),
			expectedErr: errors.New("making request request error"),
		},
		{
			name:         "unexpected status code",
			responseCode: http.StatusInternalServerError,
			expectedErr:  errors.New("evaluating: status=500 "),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := newMockHttpClient(t)

			mockClient.EXPECT().Do(mock.Anything).Run(func(req *http.Request) {
				assert.Equal(t, "POST", req.Method)
				assert.Equal(t, "/api/v1/evaluate", req.URL.Path)
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
				assert.Equal(t, "application/json", req.Header.Get("Accept"))
			}).Return(&http.Response{
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				StatusCode: tt.responseCode,
				Body:       io.NopCloser(bytes.NewReader(tt.responseBody)),
			}, tt.reqErr)

			s := &Service{
				client: mockClient,
			}

			evalCtx := map[string]interface{}{
				"foo":           "bar",
				of.TargetingKey: "12345",
			}

			actual, err := s.Evaluate(context.Background(), "foo", evalCtx)
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

	_, err := s.Evaluate(context.Background(), "foo", nil)
	assert.EqualError(t, err, of.NewInvalidContextResolutionError("evalCtx is nil").Error())

	_, err = s.Evaluate(context.Background(), "foo", map[string]interface{}{})
	assert.EqualError(t, err, of.NewTargetingKeyMissingResolutionError("targetingKey is missing").Error())
}
