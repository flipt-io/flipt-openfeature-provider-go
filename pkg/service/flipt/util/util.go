package util

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	of "github.com/open-feature/go-sdk/pkg/openfeature"
)

func GRPCToOpenFeatureError(err error) of.ResolutionError {
	s, ok := status.FromError(err)
	if !ok {
		return of.NewGeneralResolutionError("internal error")
	}

	switch s.Code() {
	case codes.NotFound:
		return of.NewFlagNotFoundResolutionError(s.Message())
	case codes.InvalidArgument:
		return of.NewInvalidContextResolutionError(s.Message())
	case codes.Unavailable:
		return of.NewProviderNotReadyResolutionError(s.Message())
	}

	return of.NewGeneralResolutionError(s.Message())
}
