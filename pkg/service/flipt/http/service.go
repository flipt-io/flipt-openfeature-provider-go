package servicehttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	of "github.com/open-feature/go-sdk/pkg/openfeature"
	"google.golang.org/grpc/codes"

	service "github.com/flipt-io/openfeature-provider-go/pkg/service/flipt"
	"go.flipt.io/flipt-grpc"
)

const (
	requestID = "requestID"

	defaultHost = "localhost"
	defaultPort = 8080
)

var _ service.Service = (*Service)(nil)

type Protocol int

const (
	HTTP Protocol = iota
	HTTPS
)

func (p Protocol) String() string {
	switch p {
	case HTTPS:
		return "https"
	default:
		return "http"
	}
}

//go:generate mockery --name=httpClient --case=underscore --inpackage --filename=service_support_test.go --testonly --with-expecter --disable-version-string
type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type errorBody struct {
	Error string     `json:"error"`
	Code  codes.Code `json:"code"`
}

type Service struct {
	client   httpClient
	host     string
	port     uint
	protocol Protocol
}

type Option func(*Service)

func WithHTTPClient(client httpClient) Option {
	return func(s *Service) {
		if client != nil {
			s.client = client
		}
	}
}

func WithHost(host string) Option {
	return func(s *Service) {
		s.host = host
	}
}

func WithPort(port uint) Option {
	return func(s *Service) {
		s.port = port
	}
}

func WithHTTPS() Option {
	return func(s *Service) {
		s.protocol = HTTPS
	}
}

func New(opts ...Option) *Service {
	s := &Service{
		host:     defaultHost,
		port:     defaultPort,
		protocol: HTTP,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// this never returns an error but wanted to make it similar to the grpc service.
func (s *Service) instance() (httpClient, error) { //nolint
	if s.client != nil {
		return s.client, nil
	}

	s.client = &http.Client{}

	return s.client, nil
}

func (s *Service) GetFlag(ctx context.Context, flagKey string) (*flipt.Flag, error) {
	url := fmt.Sprintf("%s://%s:%d/api/v1/flags/%s", s.protocol, s.host, s.port, flagKey)
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
		if err := json.Unmarshal(b, f); err != nil {
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

	url := fmt.Sprintf("%s://%s:%d/api/v1/evaluate", s.protocol, s.host, s.port)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(b))

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
		if err := json.Unmarshal(b, e); err != nil {
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
