package servicehttp

import (
	"context"
	"fmt"

	of "github.com/open-feature/go-sdk/pkg/openfeature"

	"go.flipt.io/flipt-openfeature-provider/pkg/service/flipt/util"

	offlipt "go.flipt.io/flipt-openfeature-provider/pkg/service/flipt"
	flipt "go.flipt.io/flipt/rpc/flipt"
	sdk "go.flipt.io/flipt/sdk/go"
	sdkhttp "go.flipt.io/flipt/sdk/go/http"
)

const (
	requestID   = "requestID"
	defaultAddr = "http://localhost:8080"
)

// Service is a http service.
type Service struct {
	client  offlipt.Client
	address string
}

// Option is a service option.
type Option func(*Service)

// WithAddress sets the address for the target Flipt service.
func WithAddress(address string) Option {
	return func(s *Service) {
		s.address = address
	}
}

// New creates a new http(s) service.
func New(opts ...Option) *Service {
	s := &Service{
		address: defaultAddr,
	}

	for _, opt := range opts {
		opt(s)
	}

	fliptSdk := sdk.New(sdkhttp.NewTransport(s.address))

	s.client = fliptSdk.Flipt()

	return s
}

// GetFlag returns a flag if it exists for the given namepsace/flag key pair.
func (s *Service) GetFlag(ctx context.Context, namespaceKey, flagKey string) (*flipt.Flag, error) {
	flag, err := s.client.GetFlag(ctx, &flipt.GetFlagRequest{
		Key:          flagKey,
		NamespaceKey: namespaceKey,
	})
	if err != nil {
		return nil, util.GRPCToOpenFeatureError(err)
	}

	return flag, nil
}

// Evaluate evaluates a flag with the given context.
func (s *Service) Evaluate(ctx context.Context, namespaceKey, flagKey string, evalCtx map[string]interface{}) (*flipt.EvaluationResponse, error) {
	if evalCtx == nil {
		return nil, of.NewInvalidContextResolutionError("evalCtx is nil")
	}

	ec := convertMapInterface(evalCtx)

	targetingKey := ec[of.TargetingKey]
	if targetingKey == "" {
		return nil, of.NewTargetingKeyMissingResolutionError("targetingKey is missing")
	}

	e, err := s.client.Evaluate(ctx, &flipt.EvaluationRequest{
		FlagKey:      flagKey,
		NamespaceKey: namespaceKey,
		EntityId:     targetingKey,
		RequestId:    ec[requestID],
		Context:      ec,
	})
	if err != nil {
		return nil, util.GRPCToOpenFeatureError(err)
	}

	return e, nil
}

func convertMapInterface(m map[string]interface{}) map[string]string {
	ee := make(map[string]string)
	for k, v := range m {
		ee[k] = fmt.Sprintf("%v", v)
	}

	return ee
}
