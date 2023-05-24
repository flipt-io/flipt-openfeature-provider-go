package flipt

import (
	"context"

	flipt "go.flipt.io/flipt/rpc/flipt"
)

//go:generate mockery --name=Client --case=underscore --inpackage --filename=service_support.go --testonly --with-expecter --disable-version-string
type Client interface {
	GetFlag(ctx context.Context, c *flipt.GetFlagRequest) (*flipt.Flag, error)
	Evaluate(ctx context.Context, v *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error)
}
