package flipt

import (
	"context"

	"go.flipt.io/flipt-grpc"
)

//go:generate mockery --name=Service --structname=mockService --case=underscore --output=../provider --outpkg=flipt --filename=provider_support_test.go --testonly --with-expecter --disable-version-string

type Service interface {
	GetFlag(ctx context.Context, flagKey string) (*flipt.Flag, error)
	Evaluate(ctx context.Context, flagKey string, evalCtx map[string]interface{}) (*flipt.EvaluationResponse, error)
}
