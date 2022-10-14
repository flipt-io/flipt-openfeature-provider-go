package flipt

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	service "github.com/flipt-io/openfeature-provider-go/pkg/service/flipt"
	of "github.com/open-feature/go-sdk/pkg/openfeature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.flipt.io/flipt-grpc"
)

func TestServiceType(t *testing.T) {
	tests := []struct {
		name        string
		serviceType ServiceType
	}{
		{
			name:        "http",
			serviceType: ServiceHTTP,
		},
		{
			name:        "https",
			serviceType: ServiceHTTPS,
		},
		{
			name:        "grpc",
			serviceType: ServiceGRPC,
		},
		{
			name:        "unknown",
			serviceType: ServiceType(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.name, tt.serviceType.String())
		})
	}
}

func TestNew(t *testing.T) {
	type want = struct {
		config Config
		svc    service.Service
	}
	tests := []struct {
		name string
		opts []Option
		want want
	}{
		{
			name: "default",
			want: want{
				config: DefaultConfig,
			},
		},
		{
			name: "with service type",
			opts: []Option{WithServiceType(ServiceGRPC)},
			want: want{
				config: Config{
					ServiceType: ServiceGRPC,
					Port:        8080,
					Host:        "localhost",
				},
			},
		},
		{
			name: "with port",
			opts: []Option{WithPort(8081)},
			want: want{
				config: Config{
					ServiceType: ServiceHTTP,
					Port:        8081,
					Host:        "localhost",
				},
			},
		},
		{
			name: "with host",
			opts: []Option{WithHost("github.com")},
			want: want{
				config: Config{
					ServiceType: ServiceHTTP,
					Port:        8080,
					Host:        "github.com",
				},
			},
		},
		{
			name: "with certificate path",
			opts: []Option{WithCertificatePath("/path/to/cert")},
			want: want{
				config: Config{
					ServiceType:     ServiceHTTP,
					Port:            8080,
					Host:            "localhost",
					CertificatePath: "/path/to/cert",
				},
			},
		},
		{
			name: "with socket path",
			opts: []Option{WithSocketPath("/path/to/socket")},
			want: want{
				config: Config{
					ServiceType: ServiceHTTP,
					Port:        8080,
					Host:        "localhost",
					SocketPath:  "/path/to/socket",
				},
			},
		},
		{
			name: "with config",
			opts: []Option{WithConfig(Config{
				ServiceType:     ServiceHTTPS,
				Port:            8081,
				Host:            "github.com",
				CertificatePath: "/path/to/cert",
				SocketPath:      "/path/to/socket",
			})},
			want: want{
				config: Config{
					ServiceType:     ServiceHTTPS,
					Port:            8081,
					Host:            "github.com",
					CertificatePath: "/path/to/cert",
					SocketPath:      "/path/to/socket",
				},
			},
		},
		{
			name: "with service",
			opts: []Option{WithService(&mockService{})},
			want: want{
				config: DefaultConfig,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(tt.opts...)

			assert.Equal(t, tt.want.config, p.config)
			assert.NotNil(t, p.svc)
		})
	}
}

func TestMetadata(t *testing.T) {
	p := NewProvider()
	assert.Equal(t, "flipt-provider", p.Metadata().Name)
}

func TestBooleanEvaluation(t *testing.T) {
	tests := []struct {
		name            string
		flagKey         string
		defaultValue    bool
		mockRespFlag    *flipt.Flag
		mockRespFlagErr error
		expected        of.BoolResolutionDetail
	}{
		{
			name:         "true",
			flagKey:      "boolean-true",
			defaultValue: false,
			mockRespFlag: &flipt.Flag{
				Key:     "boolean-true",
				Enabled: true,
			},
			expected: of.BoolResolutionDetail{Value: true, ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.TargetingMatchReason}},
		},
		{
			name:         "false",
			flagKey:      "boolean-false",
			defaultValue: true,
			mockRespFlag: &flipt.Flag{
				Key:     "boolean-false",
				Enabled: false,
			},
			expected: of.BoolResolutionDetail{Value: false, ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.DisabledReason}},
		},
		{
			name:            "flag not found",
			flagKey:         "boolean-not-found",
			defaultValue:    true,
			mockRespFlagErr: of.NewFlagNotFoundResolutionError("flag not found"),
			expected: of.BoolResolutionDetail{
				Value: true,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewFlagNotFoundResolutionError("flag not found"),
				},
			},
		},
		{
			name:            "resolution error",
			flagKey:         "boolean-res-error",
			defaultValue:    true,
			mockRespFlagErr: of.NewInvalidContextResolutionError("boom"),
			expected: of.BoolResolutionDetail{
				Value: true,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewInvalidContextResolutionError("boom"),
				},
			},
		},
		{
			name:            "error",
			flagKey:         "boolean-error",
			defaultValue:    true,
			mockRespFlagErr: errors.New("boom"),
			expected: of.BoolResolutionDetail{
				Value: true,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewGeneralResolutionError("boom"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := newMockService(t)
			mockSvc.On("GetFlag", mock.Anything, tt.flagKey).Return(tt.mockRespFlag, tt.mockRespFlagErr)

			p := NewProvider(WithService(mockSvc))
			actual := p.BooleanEvaluation(context.Background(), tt.flagKey, tt.defaultValue, map[string]interface{}{})

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestStringEvaluation(t *testing.T) {
	tests := []struct {
		name                  string
		flagKey               string
		defaultValue          string
		mockRespFlag          *flipt.Flag
		mockRespFlagErr       error
		mockRespEvaluation    *flipt.EvaluationResponse
		mockRespEvaluationErr error
		expected              of.StringResolutionDetail
	}{
		{
			name:         "flag enabled",
			flagKey:      "string-true",
			defaultValue: "false",
			mockRespFlag: &flipt.Flag{
				Key:     "string-true",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "string-true",
				Match:   true,
				Value:   "true",
			},
			expected: of.StringResolutionDetail{Value: "true", ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.TargetingMatchReason}},
		},
		{
			name:         "flag disabled",
			flagKey:      "string-true",
			defaultValue: "false",
			mockRespFlag: &flipt.Flag{
				Key:     "string-true",
				Enabled: false,
			},
			expected: of.StringResolutionDetail{Value: "false", ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.DisabledReason}},
		},
		{
			name:            "flag not found",
			flagKey:         "string-not-found",
			defaultValue:    "true",
			mockRespFlagErr: of.NewFlagNotFoundResolutionError("flag not found"),
			expected: of.StringResolutionDetail{
				Value: "true",
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewFlagNotFoundResolutionError("flag not found"),
				},
			},
		},
		{
			name:         "resolution error",
			flagKey:      "string-res-error",
			defaultValue: "true",
			mockRespFlag: &flipt.Flag{
				Key:     "string-res-error",
				Enabled: true,
			},
			mockRespEvaluationErr: of.NewInvalidContextResolutionError("boom"),
			expected: of.StringResolutionDetail{
				Value: "true",
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewInvalidContextResolutionError("boom"),
				},
			},
		},
		{
			name:         "error",
			flagKey:      "string-error",
			defaultValue: "true",
			mockRespFlag: &flipt.Flag{
				Key:     "string-error",
				Enabled: true,
			},
			mockRespEvaluationErr: errors.New("boom"),
			expected: of.StringResolutionDetail{
				Value: "true",
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewGeneralResolutionError("boom"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := newMockService(t)
			mockSvc.On("GetFlag", mock.Anything, tt.flagKey).Return(tt.mockRespFlag, tt.mockRespFlagErr)
			mockSvc.On("Evaluate", mock.Anything, tt.flagKey, mock.Anything).Return(tt.mockRespEvaluation, tt.mockRespEvaluationErr).Maybe()

			p := NewProvider(WithService(mockSvc))
			actual := p.StringEvaluation(context.Background(), tt.flagKey, tt.defaultValue, map[string]interface{}{})

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestFloatEvaluation(t *testing.T) {
	tests := []struct {
		name                  string
		flagKey               string
		defaultValue          float64
		mockRespFlag          *flipt.Flag
		mockRespFlagErr       error
		mockRespEvaluation    *flipt.EvaluationResponse
		mockRespEvaluationErr error
		expected              of.FloatResolutionDetail
	}{
		{
			name:         "flag enabled",
			flagKey:      "float-one",
			defaultValue: 1.0,
			mockRespFlag: &flipt.Flag{
				Key:     "float-one",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "float-one",
				Match:   true,
				Value:   "1.0",
			},
			expected: of.FloatResolutionDetail{Value: 1.0, ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.TargetingMatchReason}},
		},
		{
			name:         "flag disabled",
			flagKey:      "float-zero",
			defaultValue: 0.0,
			mockRespFlag: &flipt.Flag{
				Key:     "float-zero",
				Enabled: false,
			},
			expected: of.FloatResolutionDetail{Value: 0.0, ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.DisabledReason}},
		},
		{
			name:            "flag not found",
			flagKey:         "float-not-found",
			defaultValue:    1.0,
			mockRespFlagErr: of.NewFlagNotFoundResolutionError("flag not found"),
			expected: of.FloatResolutionDetail{
				Value: 1.0,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewFlagNotFoundResolutionError("flag not found"),
				},
			},
		},
		{
			name:         "resolution error",
			flagKey:      "float-res-error",
			defaultValue: 0.0,
			mockRespFlag: &flipt.Flag{
				Key:     "float-res-error",
				Enabled: true,
			},
			mockRespEvaluationErr: of.NewInvalidContextResolutionError("boom"),
			expected: of.FloatResolutionDetail{
				Value: 0.0,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewInvalidContextResolutionError("boom"),
				},
			},
		},
		{
			name:         "parse error",
			flagKey:      "float-parse-error",
			defaultValue: 1.0,
			mockRespFlag: &flipt.Flag{
				Key:     "float-parse-error",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "float-parse-error",
				Match:   true,
				Value:   "not-a-float",
			},
			expected: of.FloatResolutionDetail{
				Value: 1.0,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewTypeMismatchResolutionError("value is not a float"),
				},
			},
		},
		{
			name:         "error",
			flagKey:      "float-error",
			defaultValue: 1.0,
			mockRespFlag: &flipt.Flag{
				Key:     "float-error",
				Enabled: true,
			},
			mockRespEvaluationErr: errors.New("boom"),
			expected: of.FloatResolutionDetail{
				Value: 1.0,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewGeneralResolutionError("boom"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := newMockService(t)
			mockSvc.On("GetFlag", mock.Anything, tt.flagKey).Return(tt.mockRespFlag, tt.mockRespFlagErr)
			mockSvc.On("Evaluate", mock.Anything, tt.flagKey, mock.Anything).Return(tt.mockRespEvaluation, tt.mockRespEvaluationErr).Maybe()

			p := NewProvider(WithService(mockSvc))
			actual := p.FloatEvaluation(context.Background(), tt.flagKey, tt.defaultValue, map[string]interface{}{})

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestIntEvaluation(t *testing.T) {
	tests := []struct {
		name                  string
		flagKey               string
		defaultValue          int64
		mockRespFlag          *flipt.Flag
		mockRespFlagErr       error
		mockRespEvaluation    *flipt.EvaluationResponse
		mockRespEvaluationErr error
		expected              of.IntResolutionDetail
	}{
		{
			name:         "flag enabled",
			flagKey:      "int-one",
			defaultValue: 1,
			mockRespFlag: &flipt.Flag{
				Key:     "int-one",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "int-one",
				Match:   true,
				Value:   "1",
			},
			expected: of.IntResolutionDetail{Value: 1, ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.TargetingMatchReason}},
		},
		{
			name:         "flag disabled",
			flagKey:      "int-zero",
			defaultValue: 0,
			mockRespFlag: &flipt.Flag{
				Key:     "int-zero",
				Enabled: false,
			},
			expected: of.IntResolutionDetail{Value: 0, ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.DisabledReason}},
		},
		{
			name:            "flag not found",
			flagKey:         "int-not-found",
			defaultValue:    1,
			mockRespFlagErr: of.NewFlagNotFoundResolutionError("flag not found"),
			expected: of.IntResolutionDetail{
				Value: 1,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewFlagNotFoundResolutionError("flag not found"),
				},
			},
		},
		{
			name:         "resolution error",
			flagKey:      "int-res-error",
			defaultValue: 0,
			mockRespFlag: &flipt.Flag{
				Key:     "int-res-error",
				Enabled: true,
			},
			mockRespEvaluationErr: of.NewInvalidContextResolutionError("boom"),
			expected: of.IntResolutionDetail{
				Value: 0,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewInvalidContextResolutionError("boom"),
				},
			},
		},
		{
			name:         "parse error",
			flagKey:      "int-parse-error",
			defaultValue: 1,
			mockRespFlag: &flipt.Flag{
				Key:     "int-parse-error",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "int-parse-error",
				Match:   true,
				Value:   "not-an-int",
			},
			expected: of.IntResolutionDetail{
				Value: 1,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewTypeMismatchResolutionError("value is not an integer"),
				},
			},
		},
		{
			name:         "error",
			flagKey:      "int-error",
			defaultValue: 1,
			mockRespFlag: &flipt.Flag{
				Key:     "int-error",
				Enabled: true,
			},
			mockRespEvaluationErr: errors.New("boom"),
			expected: of.IntResolutionDetail{
				Value: 1,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewGeneralResolutionError("boom"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := newMockService(t)
			mockSvc.On("GetFlag", mock.Anything, tt.flagKey).Return(tt.mockRespFlag, tt.mockRespFlagErr)
			mockSvc.On("Evaluate", mock.Anything, tt.flagKey, mock.Anything).Return(tt.mockRespEvaluation, tt.mockRespEvaluationErr).Maybe()

			p := NewProvider(WithService(mockSvc))
			actual := p.IntEvaluation(context.Background(), tt.flagKey, tt.defaultValue, map[string]interface{}{})

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestObjectEvaluation(t *testing.T) {
	attachment := map[string]interface{}{
		"foo": "bar",
	}

	b, _ := json.Marshal(attachment)
	attachmentJSON := string(b)

	tests := []struct {
		name                  string
		flagKey               string
		defaultValue          map[string]interface{}
		mockRespFlag          *flipt.Flag
		mockRespFlagErr       error
		mockRespEvaluation    *flipt.EvaluationResponse
		mockRespEvaluationErr error
		expected              of.InterfaceResolutionDetail
	}{
		{
			name:    "flag enabled",
			flagKey: "obj-enabled",
			defaultValue: map[string]interface{}{
				"baz": "qux",
			},
			mockRespFlag: &flipt.Flag{
				Key:     "obj-enabled",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey:    "obj-enabled",
				Match:      true,
				Attachment: attachmentJSON,
			},
			expected: of.InterfaceResolutionDetail{
				Value:                    attachment,
				ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.TargetingMatchReason},
			},
		},
		{
			name:    "flag disabled",
			flagKey: "obj-disabled",
			defaultValue: map[string]interface{}{
				"baz": "qux",
			}, mockRespFlag: &flipt.Flag{
				Key:     "obj-disabled",
				Enabled: false,
			},
			expected: of.InterfaceResolutionDetail{
				Value: map[string]interface{}{
					"baz": "qux",
				},
				ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.DisabledReason},
			},
		},
		{
			name:    "flag not found",
			flagKey: "obj-not-found",
			defaultValue: map[string]interface{}{
				"baz": "qux",
			}, mockRespFlagErr: of.NewFlagNotFoundResolutionError("flag not found"),
			expected: of.InterfaceResolutionDetail{
				Value: map[string]interface{}{
					"baz": "qux",
				},
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewFlagNotFoundResolutionError("flag not found"),
				},
			},
		},
		{
			name:    "resolution error",
			flagKey: "obj-res-error",
			defaultValue: map[string]interface{}{
				"baz": "qux",
			}, mockRespFlag: &flipt.Flag{
				Key:     "obj-res-error",
				Enabled: true,
			},
			mockRespEvaluationErr: of.NewInvalidContextResolutionError("boom"),
			expected: of.InterfaceResolutionDetail{
				Value: map[string]interface{}{
					"baz": "qux",
				}, ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewInvalidContextResolutionError("boom"),
				},
			},
		},
		{
			name:    "unmarshal error",
			flagKey: "obj-unmarshal-error",
			defaultValue: map[string]interface{}{
				"baz": "qux",
			},
			mockRespFlag: &flipt.Flag{
				Key:     "obj-unmarshal-error",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey:    "obj-unmarshal-error",
				Match:      true,
				Attachment: "x",
			},
			expected: of.InterfaceResolutionDetail{
				Value: map[string]interface{}{
					"baz": "qux",
				},
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewTypeMismatchResolutionError("value is not an object: \"x\""),
				},
			},
		},
		{
			name:    "error",
			flagKey: "obj-error",
			defaultValue: map[string]interface{}{
				"baz": "qux",
			},
			mockRespFlag: &flipt.Flag{
				Key:     "obj-error",
				Enabled: true,
			},
			mockRespEvaluationErr: errors.New("boom"),
			expected: of.InterfaceResolutionDetail{
				Value: map[string]interface{}{
					"baz": "qux",
				},
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewGeneralResolutionError("boom"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := newMockService(t)
			mockSvc.On("GetFlag", mock.Anything, tt.flagKey).Return(tt.mockRespFlag, tt.mockRespFlagErr)
			mockSvc.On("Evaluate", mock.Anything, tt.flagKey, mock.Anything).Return(tt.mockRespEvaluation, tt.mockRespEvaluationErr).Maybe()

			p := NewProvider(WithService(mockSvc))
			actual := p.ObjectEvaluation(context.Background(), tt.flagKey, tt.defaultValue, map[string]interface{}{})

			assert.Equal(t, tt.expected, actual)
		})
	}
}
