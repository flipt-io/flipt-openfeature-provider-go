package servicehttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/protobuf/encoding/protojson"

	of "github.com/open-feature/go-sdk/pkg/openfeature"
	"google.golang.org/grpc/codes"

	"go.flipt.io/flipt-grpc"
)

const (
	requestID   = "requestID"
	defaultAddr = "http://localhost:8080"
)

//go:generate mockery --name=httpClient --case=underscore --inpackage --filename=service_support_test.go --testonly --with-expecter --disable-version-string
type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type httpClientFunc func(req *http.Request) (*http.Response, error)

func (h httpClientFunc) Do(req *http.Request) (*http.Response, error) {
	return h(req)
}

// otelPropagationClient uses the provide TextMapPropagator to propagate any
// trace context from the requests context onto its outgoing headers
func otelPropagationClient(client httpClient, propagator propagation.TextMapPropagator) httpClient {
	return httpClientFunc(func(req *http.Request) (*http.Response, error) {
		if propagator != nil {
			propagator.Inject(req.Context(), propagation.HeaderCarrier(req.Header))
		}

		return client.Do(req)
	})
}

type errorBody struct {
	Error string     `json:"error"`
	Code  codes.Code `json:"code"`
}

// Service is a http service.
type Service struct {
	client     httpClient
	address    string
	propagator propagation.TextMapPropagator
}

// Option is a service option.
type Option func(*Service)

// WithHTTPClient sets the HTTP client to use.
func WithHTTPClient(client httpClient) Option {
	return func(s *Service) {
		if client != nil {
			s.client = client
		}
	}
}

// WithAddress sets the address for the target Flipt service.
func WithAddress(address string) Option {
	return func(s *Service) {
		s.address = address
	}
}

// WithPropagator overrides the default propagation.TextMapPropagator used
// to propagate trace context through the service calls.
func WithPropagator(propagator propagation.TextMapPropagator) Option {
	return func(s *Service) {
		s.propagator = propagator
	}
}

// New creates a new http(s) service.
func New(opts ...Option) *Service {
	s := &Service{
		address: defaultAddr,
		// default is to use globally registered propagator
		propagator: otel.GetTextMapPropagator(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Service) url(path string) string {
	address := s.address
	if strings.HasPrefix(s.address, "unix://") {
		address = "http://unix"
	}

	return address + path
}

// this never returns an error but wanted to make it similar to the grpc service.
func (s *Service) instance() (client httpClient, _ error) { //nolint
	// defer decorating resulting client with middleware
	defer func() { client = otelPropagationClient(client, s.propagator) }()

	if s.client != nil {
		return s.client, nil
	}

	// the following dialer and transport defaults are copied
	// from net/http.DefaultTransport setup
	// with the addition of `unix` socket support
	dialer := (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext

	// support unix:// scheme addresses
	if strings.HasPrefix(s.address, "unix://") {
		dialer = func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", strings.TrimPrefix(s.address, "unix://"))
		}
	}

	s.client = &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialer,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	return s.client, nil
}

// GetFlag returns a flag if it exists for the given key.
func (s *Service) GetFlag(ctx context.Context, flagKey string) (*flipt.Flag, error) {
	url := s.url("/api/v1/flags/" + flagKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return nil, fmt.Errorf("creating request %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	conn, err := s.instance()
	if err != nil {
		return nil, err
	}

	resp, err := conn.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request %w", err)
	}

	b, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	// reset the body/buffer incase we need to read it again
	resp.Body = io.NopCloser(bytes.NewBuffer(b))

	if err != nil {
		return nil, fmt.Errorf("reading response body %w", err)
	}

	if resp.StatusCode == http.StatusOK {
		f := &flipt.Flag{}
		if err := protojson.Unmarshal(b, f); err != nil {
			return nil, fmt.Errorf("unmarshalling response body %w", err)
		}

		return f, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		// check if content-type is json and if so, unmarshal the error
		if resp.Header.Get("Content-Type") == "application/json" {
			errorBody := &errorBody{}
			if err := json.Unmarshal(b, errorBody); err != nil {
				return nil, fmt.Errorf("unmarshalling response body %w", err)
			}

			// here we can guarantee that the error is a grpc error and that it is a NotFound error
			if errorBody.Code == codes.NotFound {
				return nil, of.NewFlagNotFoundResolutionError(fmt.Sprintf("flag %q not found", flagKey))
			}
		}
	}

	return nil, fmt.Errorf("getting flag: status=%d %s", resp.StatusCode, string(b))
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

	body := &flipt.EvaluationRequest{
		FlagKey:   flagKey,
		EntityId:  targetingKey,
		RequestId: ec[requestID],
		Context:   ec,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshalling request body %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url("/api/v1/evaluate"), bytes.NewBuffer(b))

	if err != nil {
		return nil, fmt.Errorf("creating request %w", err)
	}

	req.Method = http.MethodPost
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	conn, err := s.instance()
	if err != nil {
		return nil, err
	}

	resp, err := conn.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request %w", err)
	}

	b, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	// reset the body/buffer incase we need to read it again
	resp.Body = io.NopCloser(bytes.NewBuffer(b))

	if err != nil {
		return nil, fmt.Errorf("reading response body %w", err)
	}

	if resp.StatusCode == http.StatusOK {
		e := &flipt.EvaluationResponse{}
		if err := protojson.Unmarshal(b, e); err != nil {
			return nil, fmt.Errorf("unmarshalling response body %w", err)
		}

		return e, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		// check if content-type is json and if so, unmarshal the error
		if resp.Header.Get("Content-Type") == "application/json" {
			errorBody := &errorBody{}
			if err := json.Unmarshal(b, errorBody); err != nil {
				return nil, fmt.Errorf("unmarshalling response body %w", err)
			}

			// here we can guarantee that the error is a grpc error and that it is a NotFound error
			if errorBody.Code == codes.NotFound {
				return nil, of.NewFlagNotFoundResolutionError(fmt.Sprintf("flag %q not found", flagKey))
			}
		}
	}

	// here it could be that the endpoint is not found (ie the server url is configured wrong), so we return a generic error
	return nil, fmt.Errorf("evaluating: status=%d %s", resp.StatusCode, string(b))
}

func convertMapInterface(m map[string]interface{}) map[string]string {
	ee := make(map[string]string)
	for k, v := range m {
		ee[k] = fmt.Sprintf("%v", v)
	}

	return ee
}
