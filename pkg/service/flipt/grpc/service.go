package servicegrpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"sync"

	of "github.com/open-feature/go-sdk/pkg/openfeature"
	"go.flipt.io/flipt-grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	requestID   = "requestID"
	defaultAddr = "localhost:9000"
)

//go:generate mockery --name=grpcClient --case=underscore --inpackage --filename=service_support_test.go --testonly --with-expecter --disable-version-string

type grpcClient interface {
	GetFlag(ctx context.Context, in *flipt.GetFlagRequest, opts ...grpc.CallOption) (*flipt.Flag, error)
	Evaluate(ctx context.Context, in *flipt.EvaluationRequest, opts ...grpc.CallOption) (*flipt.EvaluationResponse, error)
}

// Service is a GRPC service.
type Service struct {
	client          grpcClient
	address         string
	certificatePath string
	once            sync.Once
}

// Option is a service option.
type Option func(*Service)

// WithGRPCClient sets the GRPC client to use.
func WithGRPCClient(client grpcClient) Option {
	return func(s *Service) {
		if client != nil {
			s.client = client
		}
	}
}

// WithAddress sets the address for the remote Flipt gRPC API.
func WithAddress(address string) Option {
	if strings.HasPrefix(address, "unix://") {
		address = "passthrough:///" + address
	}

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

// New creates a new GRPC service.
func New(opts ...Option) *Service {
	s := &Service{
		address: defaultAddr,
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

	conn, err := grpc.Dial(
		s.address,
		grpc.WithTransportCredentials(credentials),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("dialing %w", err)
	}

	return conn, nil
}

func (s *Service) instance() (grpcClient, error) {
	if s.client != nil {
		return s.client, nil
	}

	var err error

	s.once.Do(func() {
		conn, cerr := s.connect()
		if cerr != nil {
			err = fmt.Errorf("connecting %w", cerr)
		}

		s.client = flipt.NewFliptClient(conn)
	})

	return s.client, err
}

// GetFlag returns a flag if it exists for the given key.
func (s *Service) GetFlag(ctx context.Context, flagKey string) (*flipt.Flag, error) {
	conn, err := s.instance()
	if err != nil {
		return nil, err
	}

	f, err := conn.GetFlag(ctx, &flipt.GetFlagRequest{Key: flagKey})

	if err != nil {
		if s, ok := status.FromError(err); ok {
			return nil, grpcToOpenFeatureError(*s)
		}

		return nil, fmt.Errorf("getting flag %q %w", flagKey, err)
	}

	return f, nil
}

// Evaluate evaluates a flag with the given context.
func (s *Service) Evaluate(ctx context.Context, flagKey string, evalCtx map[string]interface{}) (*flipt.EvaluationResponse, error) {
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

	resp, err := conn.Evaluate(ctx, &flipt.EvaluationRequest{FlagKey: flagKey, EntityId: targetingKey, RequestId: ec[requestID], Context: ec})

	if err != nil {
		if s, ok := status.FromError(err); ok {
			return nil, grpcToOpenFeatureError(*s)
		}

		return nil, fmt.Errorf("evaluating flag %q %w", flagKey, err)
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

func grpcToOpenFeatureError(s status.Status) of.ResolutionError {
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
