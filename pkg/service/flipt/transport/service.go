package transport

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	of "github.com/open-feature/go-sdk/pkg/openfeature"
	offlipt "go.flipt.io/flipt-openfeature-provider/pkg/service/flipt"
	"go.flipt.io/flipt-openfeature-provider/pkg/service/flipt/util"
	flipt "go.flipt.io/flipt/rpc/flipt"
	sdk "go.flipt.io/flipt/sdk/go"
	sdkgrpc "go.flipt.io/flipt/sdk/go/grpc"
	sdkhttp "go.flipt.io/flipt/sdk/go/http"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	requestID   = "requestID"
	defaultAddr = "http://localhost:8080"
)

// Service is a Transport service.
type Service struct {
	client            offlipt.Client
	address           string
	certificatePath   string
	unaryInterceptors []grpc.UnaryClientInterceptor
	once              sync.Once
	tokenProvider     sdk.ClientTokenProvider
}

// Option is a service option.
type Option func(*Service)

// WithAddress sets the address for the remote Flipt gRPC API.
func WithAddress(address string) Option {
	return func(s *Service) {
		s.address = address
	}
}

// WithCertificatePath sets the certificate path for the service.
func WithCertificatePath(certificatePath string) Option {
	return func(s *Service) {
		s.certificatePath = certificatePath
	}
}

// WithUnaryClientInterceptor sets the provided unary client interceptors
// to be applied to the established gRPC client connection.
func WithUnaryClientInterceptor(unaryInterceptors ...grpc.UnaryClientInterceptor) Option {
	return func(s *Service) {
		s.unaryInterceptors = unaryInterceptors
	}
}

// WithClientTokenProvider sets the token provider for auth to support client
// auth needs.
func WithClientTokenProvider(tokenProvider sdk.ClientTokenProvider) Option {
	return func(s *Service) {
		s.tokenProvider = tokenProvider
	}
}

// New creates a new Transport service.
func New(opts ...Option) *Service {
	s := &Service{
		address: defaultAddr,
		unaryInterceptors: []grpc.UnaryClientInterceptor{
			// by default this establishes the otel.TextMapPropagator
			// registers to the otel package.
			otelgrpc.UnaryClientInterceptor(),
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Service) connect() (*grpc.ClientConn, error) {
	var (
		err         error
		credentials = insecure.NewCredentials()
	)

	if s.certificatePath != "" {
		credentials, err = loadTLSCredentials(s.certificatePath)
		if err != nil {
			// TODO: log error?
			credentials = insecure.NewCredentials()
		}
	}

	var address = s.address

	if strings.HasPrefix(s.address, "unix://") {
		address = "passthrough:///" + s.address
	}

	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(credentials),
		grpc.WithBlock(),
		grpc.WithChainUnaryInterceptor(s.unaryInterceptors...),
	)
	if err != nil {
		return nil, fmt.Errorf("dialing %w", err)
	}

	return conn, nil
}

func (s *Service) instance() (offlipt.Client, error) {
	if s.client != nil {
		return s.client, nil
	}

	var err error

	s.once.Do(func() {
		u, uerr := url.Parse(s.address)
		if uerr != nil {
			err = fmt.Errorf("connecting %w", uerr)
		}

		opts := []sdk.Option{}

		if s.tokenProvider != nil {
			opts = append(opts, sdk.WithClientTokenProvider(s.tokenProvider))
		}

		if u.Scheme == "https" || u.Scheme == "http" {
			s.client = sdk.New(sdkhttp.NewTransport(s.address), opts...).Flipt()

			return
		}

		conn, cerr := s.connect()
		if cerr != nil {
			err = fmt.Errorf("connecting %w", cerr)
		}

		s.client = sdk.New(sdkgrpc.NewTransport(conn), opts...).Flipt()
	})

	return s.client, err
}

// GetFlag returns a flag if it exists for the given namespace/flag key pair.
func (s *Service) GetFlag(ctx context.Context, namespaceKey, flagKey string) (*flipt.Flag, error) {
	conn, err := s.instance()
	if err != nil {
		return nil, err
	}

	flag, err := conn.GetFlag(ctx, &flipt.GetFlagRequest{
		Key:          flagKey,
		NamespaceKey: namespaceKey,
	})
	if err != nil {
		return nil, util.GRPCToOpenFeatureError(err)
	}

	return flag, nil
}

// Evaluate evaluates a flag with the given context and namespace/flag key pair.
func (s *Service) Evaluate(ctx context.Context, namespaceKey, flagKey string, evalCtx map[string]interface{}) (*flipt.EvaluationResponse, error) {
	if evalCtx == nil {
		return nil, of.NewInvalidContextResolutionError("evalCtx is nil")
	}

	ec := convertMapInterface(evalCtx)

	targetingKey := ec[of.TargetingKey]
	if targetingKey == "" {
		return nil, of.NewTargetingKeyMissingResolutionError("targetingKey is missing")
	}

	conn, err := s.instance()
	if err != nil {
		return nil, err
	}

	resp, err := conn.Evaluate(ctx, &flipt.EvaluationRequest{FlagKey: flagKey, NamespaceKey: namespaceKey, EntityId: targetingKey, RequestId: ec[requestID], Context: ec})
	if err != nil {
		return nil, util.GRPCToOpenFeatureError(err)
	}

	return resp, nil
}

func convertMapInterface(m map[string]interface{}) map[string]string {
	ee := make(map[string]string)
	for k, v := range m {
		ee[k] = fmt.Sprintf("%v", v)
	}

	return ee
}

func loadTLSCredentials(serverCertPath string) (credentials.TransportCredentials, error) {
	pemServerCA, err := os.ReadFile(serverCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	config := &tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
	}

	return credentials.NewTLS(config), nil
}
