package util

import (
	"testing"

	of "github.com/open-feature/go-sdk/pkg/openfeature"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
			err := GRPCToOpenFeatureError(tt.grpcStatus.Err())

			assert.EqualError(t, err, tt.expectedErr.Error())
		})
	}
}
