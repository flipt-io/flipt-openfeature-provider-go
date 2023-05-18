package flipt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	of "github.com/open-feature/go-sdk/pkg/openfeature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMetadata(t *testing.T) {
	p := NewProvider()
	assert.Equal(t, "flipt-provider", p.Metadata().Name)
}

func TestGetFlag_GeneralError(t *testing.T) {
	mockSvc := newMockService(t)
	mockSvc.On("GetFlag", mock.Anything, "foo-namespace", "foo").Return(nil, status.Error(codes.Internal, "boom"))

	p := NewProvider(WithService(mockSvc))
	got, detail, err := p.getFlag(context.Background(), "foo-namespace", "foo")

	assert.Nil(t, got)
	assert.Equal(t, detail, of.ProviderResolutionDetail{
		ResolutionError: of.NewGeneralResolutionError("rpc error: code = Internal desc = boom"),
		Reason:          of.DefaultReason,
	})
	assert.Error(t, err)
}

func TestBooleanEvaluation(t *testing.T) {
	tests := []struct {
		name                  string
		flagKey               string
		namespaceKey          string
		defaultValue          bool
		mockRespFlag          *flipt.Flag
		mockRespFlagErr       error
		mockRespEvaluation    *flipt.EvaluationResponse
		mockRespEvaluationErr error
		expected              of.BoolResolutionDetail
	}{
		{
			name:         "true",
			flagKey:      "boolean-true",
			namespaceKey: "default",
			defaultValue: false,
			mockRespFlag: &flipt.Flag{
				Key:     "boolean-true",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "boolean-true",
				Match:   true,
			},
			expected: of.BoolResolutionDetail{Value: true, ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.DefaultReason}},
		},
		{
			name:         "false",
			flagKey:      "boolean-false",
			namespaceKey: "default",
			defaultValue: true,
			mockRespFlag: &flipt.Flag{
				Key:     "boolean-false",
				Enabled: false,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "boolean-true",
				Match:   false,
			},
			expected: of.BoolResolutionDetail{Value: true, ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.DisabledReason}},
		},
		{
			name:            "flag not found",
			flagKey:         "boolean-not-found",
			namespaceKey:    "default",
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
			name:         "resolution error",
			flagKey:      "boolean-res-error",
			namespaceKey: "default",
			defaultValue: false,
			mockRespFlag: &flipt.Flag{
				Key:     "boolean-true",
				Enabled: true,
			},
			mockRespEvaluationErr: of.NewInvalidContextResolutionError("boom"),
			expected: of.BoolResolutionDetail{
				Value: false,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewInvalidContextResolutionError("boom"),
				},
			},
		},
		{
			name:         "error",
			flagKey:      "boolean-error",
			namespaceKey: "default",
			defaultValue: false,
			mockRespFlag: &flipt.Flag{
				Key:     "boolean-true",
				Enabled: true,
			},
			mockRespEvaluationErr: errors.New("boom"),
			expected: of.BoolResolutionDetail{
				Value: false,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewGeneralResolutionError("boom"),
				},
			},
		},
		{
			name:         "no match",
			flagKey:      "boolean-no-match",
			namespaceKey: "default",
			defaultValue: false,
			mockRespFlag: &flipt.Flag{
				Key:     "boolean-no-match",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "boolean-no-match",
				Match:   false,
			},
			expected: of.BoolResolutionDetail{Value: false, ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.DefaultReason}},
		},
		{
			name:         "non bool",
			flagKey:      "boolean-no-bool",
			namespaceKey: "default",
			defaultValue: false,
			mockRespFlag: &flipt.Flag{
				Key:     "boolean-no-bool",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "boolean-no-bool",
				Match:   true,
				Value:   "abcd",
			},
			expected: of.BoolResolutionDetail{
				Value: false,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason:          of.DefaultReason,
					ResolutionError: of.NewTypeMismatchResolutionError("value is not a boolean"),
				},
			},
		},
		{
			name:         "match",
			flagKey:      "boolean-match",
			namespaceKey: "default",
			defaultValue: false,
			mockRespFlag: &flipt.Flag{
				Key:     "boolean-match",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "boolean-match",
				Match:   true,
				Value:   "false",
			},
			expected: of.BoolResolutionDetail{
				Value: false,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason: of.TargetingMatchReason,
				},
			},
		},
		{
			name:         "match",
			flagKey:      "boolean-match",
			namespaceKey: "flipt",
			defaultValue: false,
			mockRespFlag: &flipt.Flag{
				Key:     "boolean-match",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "boolean-match",
				Match:   true,
				Value:   "false",
			},
			expected: of.BoolResolutionDetail{
				Value: false,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason: of.TargetingMatchReason,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := newMockService(t)
			mockSvc.On("GetFlag", mock.Anything, tt.namespaceKey, tt.flagKey).Return(tt.mockRespFlag, tt.mockRespFlagErr)
			mockSvc.On("Evaluate", mock.Anything, tt.namespaceKey, tt.flagKey, mock.Anything).Return(tt.mockRespEvaluation, tt.mockRespEvaluationErr).Maybe()

			p := NewProvider(WithService(mockSvc))

			f := tt.flagKey
			if tt.namespaceKey != "default" {
				f = fmt.Sprintf("%s/%s", tt.namespaceKey, tt.flagKey)
			}

			actual := p.BooleanEvaluation(context.Background(), f, tt.defaultValue, map[string]interface{}{})

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestStringEvaluation(t *testing.T) {
	tests := []struct {
		name                  string
		flagKey               string
		namespaceKey          string
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
			namespaceKey: "default",
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
			namespaceKey: "default",
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
			namespaceKey:    "default",
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
			namespaceKey: "default",
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
			namespaceKey: "default",
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
		{
			name:         "no match",
			flagKey:      "string-no-match",
			namespaceKey: "default",
			defaultValue: "default",
			mockRespFlag: &flipt.Flag{
				Key:     "string-no-match",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "string-no-match",
				Match:   false,
			},
			expected: of.StringResolutionDetail{Value: "default", ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.DefaultReason}},
		},
		{
			name:         "match",
			flagKey:      "string-match",
			namespaceKey: "default",
			defaultValue: "default",
			mockRespFlag: &flipt.Flag{
				Key:     "string-match",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "string-match",
				Match:   true,
				Value:   "abc",
			},
			expected: of.StringResolutionDetail{
				Value: "abc",
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason: of.TargetingMatchReason,
				},
			},
		},
		{
			name:         "match",
			flagKey:      "string-match",
			namespaceKey: "flipt",
			defaultValue: "default",
			mockRespFlag: &flipt.Flag{
				Key:     "string-match",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "string-match",
				Match:   true,
				Value:   "abc",
			},
			expected: of.StringResolutionDetail{
				Value: "abc",
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason: of.TargetingMatchReason,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := newMockService(t)
			mockSvc.On("GetFlag", mock.Anything, tt.namespaceKey, tt.flagKey).Return(tt.mockRespFlag, tt.mockRespFlagErr)
			mockSvc.On("Evaluate", mock.Anything, tt.namespaceKey, tt.flagKey, mock.Anything).Return(tt.mockRespEvaluation, tt.mockRespEvaluationErr).Maybe()

			p := NewProvider(WithService(mockSvc))

			f := tt.flagKey
			if tt.namespaceKey != "default" {
				f = fmt.Sprintf("%s/%s", tt.namespaceKey, tt.flagKey)
			}

			actual := p.StringEvaluation(context.Background(), f, tt.defaultValue, map[string]interface{}{})

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestFloatEvaluation(t *testing.T) {
	tests := []struct {
		name                  string
		flagKey               string
		namespaceKey          string
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
			namespaceKey: "default",
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
			namespaceKey: "default",
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
			namespaceKey:    "default",
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
			namespaceKey: "default",
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
			namespaceKey: "default",
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
					Reason:          of.ErrorReason,
					ResolutionError: of.NewTypeMismatchResolutionError("value is not a float"),
				},
			},
		},
		{
			name:         "error",
			flagKey:      "float-error",
			namespaceKey: "default",
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
		{
			name:         "no match",
			flagKey:      "float-no-match",
			namespaceKey: "default",
			defaultValue: 1.0,
			mockRespFlag: &flipt.Flag{
				Key:     "float-no-match",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "float-no-match",
				Match:   false,
			},
			expected: of.FloatResolutionDetail{Value: 1.0, ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.DefaultReason}},
		},
		{
			name:         "match",
			flagKey:      "float-match",
			namespaceKey: "default",
			defaultValue: 1.0,
			mockRespFlag: &flipt.Flag{
				Key:     "float-match",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "float-match",
				Match:   true,
				Value:   "2.0",
			},
			expected: of.FloatResolutionDetail{
				Value: 2.0,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason: of.TargetingMatchReason,
				},
			},
		},
		{
			name:         "match",
			flagKey:      "float-match",
			namespaceKey: "flipt",
			defaultValue: 1.0,
			mockRespFlag: &flipt.Flag{
				Key:     "float-match",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "float-match",
				Match:   true,
				Value:   "2.0",
			},
			expected: of.FloatResolutionDetail{
				Value: 2.0,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason: of.TargetingMatchReason,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := newMockService(t)
			mockSvc.On("GetFlag", mock.Anything, tt.namespaceKey, tt.flagKey).Return(tt.mockRespFlag, tt.mockRespFlagErr)
			mockSvc.On("Evaluate", mock.Anything, tt.namespaceKey, tt.flagKey, mock.Anything).Return(tt.mockRespEvaluation, tt.mockRespEvaluationErr).Maybe()

			p := NewProvider(WithService(mockSvc))
			f := tt.flagKey
			if tt.namespaceKey != "default" {
				f = fmt.Sprintf("%s/%s", tt.namespaceKey, tt.flagKey)
			}

			actual := p.FloatEvaluation(context.Background(), f, tt.defaultValue, map[string]interface{}{})

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestIntEvaluation(t *testing.T) {
	tests := []struct {
		name                  string
		flagKey               string
		namespaceKey          string
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
			namespaceKey: "default",
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
			namespaceKey: "default",
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
			namespaceKey:    "default",
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
			namespaceKey: "default",
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
			namespaceKey: "default",
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
					Reason:          of.ErrorReason,
					ResolutionError: of.NewTypeMismatchResolutionError("value is not an integer"),
				},
			},
		},
		{
			name:         "error",
			flagKey:      "int-error",
			namespaceKey: "default",
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
		{
			name:         "no match",
			flagKey:      "int-no-match",
			namespaceKey: "default",
			defaultValue: 1,
			mockRespFlag: &flipt.Flag{
				Key:     "int-no-match",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "int-no-match",
				Match:   false,
			},
			expected: of.IntResolutionDetail{Value: 1, ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.DefaultReason}},
		},
		{
			name:         "match",
			flagKey:      "int-match",
			namespaceKey: "default",
			defaultValue: 1,
			mockRespFlag: &flipt.Flag{
				Key:     "int-match",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "int-match",
				Match:   true,
				Value:   "2",
			},
			expected: of.IntResolutionDetail{
				Value: 2,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason: of.TargetingMatchReason,
				},
			},
		},
		{
			name:         "match",
			flagKey:      "int-match",
			namespaceKey: "flipt",
			defaultValue: 1,
			mockRespFlag: &flipt.Flag{
				Key:     "int-match",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "int-match",
				Match:   true,
				Value:   "2",
			},
			expected: of.IntResolutionDetail{
				Value: 2,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason: of.TargetingMatchReason,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := newMockService(t)
			mockSvc.On("GetFlag", mock.Anything, tt.namespaceKey, tt.flagKey).Return(tt.mockRespFlag, tt.mockRespFlagErr)
			mockSvc.On("Evaluate", mock.Anything, tt.namespaceKey, tt.flagKey, mock.Anything).Return(tt.mockRespEvaluation, tt.mockRespEvaluationErr).Maybe()

			p := NewProvider(WithService(mockSvc))
			f := tt.flagKey
			if tt.namespaceKey != "default" {
				f = fmt.Sprintf("%s/%s", tt.namespaceKey, tt.flagKey)
			}

			actual := p.IntEvaluation(context.Background(), f, tt.defaultValue, map[string]interface{}{})

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
		namespaceKey          string
		defaultValue          map[string]interface{}
		mockRespFlag          *flipt.Flag
		mockRespFlagErr       error
		mockRespEvaluation    *flipt.EvaluationResponse
		mockRespEvaluationErr error
		expected              of.InterfaceResolutionDetail
	}{
		{
			name:         "flag enabled",
			flagKey:      "obj-enabled",
			namespaceKey: "default",
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
			name:         "flag disabled",
			flagKey:      "obj-disabled",
			namespaceKey: "default",
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
			name:         "resolution error",
			flagKey:      "obj-res-error",
			namespaceKey: "default",
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
			name:         "unmarshal error",
			flagKey:      "obj-unmarshal-error",
			namespaceKey: "default",
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
					Reason:          of.ErrorReason,
					ResolutionError: of.NewTypeMismatchResolutionError("value is not an object: \"x\""),
				},
			},
		},
		{
			name:         "error",
			flagKey:      "obj-error",
			namespaceKey: "default",
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
		{
			name:         "no match",
			flagKey:      "obj-no-match",
			namespaceKey: "default",
			defaultValue: map[string]interface{}{
				"baz": "qux",
			},
			mockRespFlag: &flipt.Flag{
				Key:     "obj-no-match",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "obj-no-match",
				Match:   false,
			},
			expected: of.InterfaceResolutionDetail{
				Value: map[string]interface{}{
					"baz": "qux",
				},
				ProviderResolutionDetail: of.ProviderResolutionDetail{Reason: of.DefaultReason},
			},
		},
		{
			name:         "match",
			flagKey:      "obj-match",
			namespaceKey: "default",
			defaultValue: map[string]interface{}{
				"baz": "qux",
			},
			mockRespFlag: &flipt.Flag{
				Key:     "obj-match",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey:    "obj-match",
				Match:      true,
				Value:      "2",
				Attachment: "{\"foo\": \"bar\"}",
			},
			expected: of.InterfaceResolutionDetail{
				Value: map[string]interface{}{
					"foo": "bar",
				},
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason: of.TargetingMatchReason,
				},
			},
		},
		{
			name:         "match no attachment",
			flagKey:      "obj-match-no-attach",
			namespaceKey: "default",
			defaultValue: map[string]interface{}{
				"baz": "qux",
			},
			mockRespFlag: &flipt.Flag{
				Key:     "obj-match-no-attach",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey: "obj-match",
				Match:   true,
				Value:   "2",
			},
			expected: of.InterfaceResolutionDetail{
				Value: map[string]interface{}{
					"baz": "qux",
				},
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason: of.DefaultReason,
				},
			},
		},
		{
			name:         "match",
			flagKey:      "obj-match",
			namespaceKey: "flipt",
			defaultValue: map[string]interface{}{
				"baz": "qux",
			},
			mockRespFlag: &flipt.Flag{
				Key:     "obj-match",
				Enabled: true,
			},
			mockRespEvaluation: &flipt.EvaluationResponse{
				FlagKey:    "obj-match",
				Match:      true,
				Value:      "2",
				Attachment: "{\"foo\": \"bar\"}",
			},
			expected: of.InterfaceResolutionDetail{
				Value: map[string]interface{}{
					"foo": "bar",
				},
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason: of.TargetingMatchReason,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := newMockService(t)
			mockSvc.On("GetFlag", mock.Anything, tt.namespaceKey, tt.flagKey).Return(tt.mockRespFlag, tt.mockRespFlagErr)
			mockSvc.On("Evaluate", mock.Anything, tt.namespaceKey, tt.flagKey, mock.Anything).Return(tt.mockRespEvaluation, tt.mockRespEvaluationErr).Maybe()

			p := NewProvider(WithService(mockSvc))
			f := tt.flagKey
			if tt.namespaceKey != "default" {
				f = fmt.Sprintf("%s/%s", tt.namespaceKey, tt.flagKey)
			}

			actual := p.ObjectEvaluation(context.Background(), f, tt.defaultValue, map[string]interface{}{})

			assert.Equal(t, tt.expected.Value, actual.Value)
		})
	}
}
