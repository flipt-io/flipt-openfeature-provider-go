package flipt

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	of "github.com/open-feature/go-sdk/pkg/openfeature"
	"go.flipt.io/flipt-grpc"
	service "go.flipt.io/flipt-openfeature-provider/pkg/service/flipt"
	serviceGRPC "go.flipt.io/flipt-openfeature-provider/pkg/service/flipt/grpc"
	serviceHTTP "go.flipt.io/flipt-openfeature-provider/pkg/service/flipt/http"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

var _ of.FeatureProvider = (*Provider)(nil)

// Config is a configuration for the FliptProvider.
type Config struct {
	ServiceType     ServiceType
	Port            uint
	Host            string
	CertificatePath string
	SocketPath      string
}

// ServiceType is the type of service.
type ServiceType int

const (
	// ServiceTypeHTTP argument, this is the default value.
	ServiceTypeHTTP ServiceType = iota + 1
	// ServiceTypeHTTPS argument, overrides the default value of http.
	ServiceTypeHTTPS
	// ServiceTypeGRPC argument, overrides the default value of http.
	ServiceTypeGRPC
)

func (s ServiceType) String() string {
	switch s {
	case ServiceTypeHTTP:
		return "http"
	case ServiceTypeHTTPS:
		return "https"
	case ServiceTypeGRPC:
		return "grpc"
	default:
		return "unknown"
	}
}

// Option is a configuration option for the provider.
type Option func(*Provider)

// WithServiceType is an Option to set the service type.
func WithServiceType(serviceType ServiceType) Option {
	return func(p *Provider) {
		p.config.ServiceType = serviceType
	}
}

// WithPort is an Option to set the port.
func WithPort(port uint) Option {
	return func(p *Provider) {
		p.config.Port = port
	}
}

// WithHost is an Option to set the host.
func WithHost(host string) Option {
	return func(p *Provider) {
		p.config.Host = host
	}
}

// WithCertificatePath is an Option to set the certificate path (grpc only).
func WithCertificatePath(certificatePath string) Option {
	return func(p *Provider) {
		p.config.CertificatePath = certificatePath
	}
}

// WithSocketPath is an Option to set the socket path (grpc only).
func WithSocketPath(socketPath string) Option {
	return func(p *Provider) {
		p.config.SocketPath = socketPath
	}
}

// WithConfig is an Option to set the entire configuration.
func WithConfig(config Config) Option {
	return func(p *Provider) {
		p.config = config
	}
}

// WithService is an Option to override the service implementation.
func WithService(svc service.Service) Option {
	return func(p *Provider) {
		p.svc = svc
	}
}

// NewProvider returns a new Flipt provider.
func NewProvider(opts ...Option) *Provider {
	p := &Provider{config: Config{
		Port:        8080,
		Host:        "localhost",
		ServiceType: ServiceTypeHTTP,
	}}

	for _, opt := range opts {
		opt(p)
	}

	if p.svc == nil {
		switch p.config.ServiceType {
		case ServiceTypeHTTP:
			opts := []serviceHTTP.Option{serviceHTTP.WithHost(p.config.Host), serviceHTTP.WithPort(p.config.Port)}
			p.svc = serviceHTTP.New(opts...)
		case ServiceTypeHTTPS:
			opts := []serviceHTTP.Option{serviceHTTP.WithHost(p.config.Host), serviceHTTP.WithPort(p.config.Port), serviceHTTP.WithHTTPS()}
			p.svc = serviceHTTP.New(opts...)
		case ServiceTypeGRPC:
			opts := []serviceGRPC.Option{serviceGRPC.WithHost(p.config.Host), serviceGRPC.WithPort(p.config.Port), serviceGRPC.WithSocketPath(p.config.SocketPath), serviceGRPC.WithCertificatePath(p.config.CertificatePath)}
			p.svc = serviceGRPC.New(opts...)
		}
	}

	return p
}

// Provider implements the FeatureProvider interface and provides functions for evaluating flags with Flipt.
type Provider struct {
	svc    service.Service
	config Config
}

// Metadata returns the metadata of the provider.
func (p Provider) Metadata() of.Metadata {
	return of.Metadata{Name: "flipt-provider"}
}

func (p Provider) getFlag(ctx context.Context, flag string) (*flipt.Flag, of.ProviderResolutionDetail, error) {
	f, err := p.svc.GetFlag(ctx, flag)
	if err != nil {
		var rerr of.ResolutionError
		if errors.As(err, &rerr) {
			return nil, of.ProviderResolutionDetail{
				ResolutionError: rerr,
				Reason:          of.DefaultReason,
			}, rerr
		}

		return nil, of.ProviderResolutionDetail{
			ResolutionError: of.NewGeneralResolutionError(err.Error()),
			Reason:          of.DefaultReason,
		}, fmt.Errorf("failed to get flag: %w", err)
	}

	return f, of.ProviderResolutionDetail{}, nil
}

// BooleanEvaluation returns a boolean flag.
func (p Provider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, evalCtx of.FlattenedContext) of.BoolResolutionDetail {
	f, res, err := p.getFlag(ctx, flag)
	if err != nil {
		return of.BoolResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: res,
		}
	}

	if !f.Enabled {
		return of.BoolResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: of.ProviderResolutionDetail{
				Reason: of.DisabledReason,
			},
		}
	}

	resp, err := p.svc.Evaluate(ctx, flag, evalCtx)
	if err != nil {
		var (
			rerr   of.ResolutionError
			detail = of.BoolResolutionDetail{
				Value: defaultValue,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason: of.DefaultReason,
				},
			}
		)

		if errors.As(err, &rerr) {
			detail.ProviderResolutionDetail.ResolutionError = rerr

			return detail
		}

		detail.ProviderResolutionDetail.ResolutionError = of.NewGeneralResolutionError(err.Error())

		return detail
	}

	if !resp.Match {
		return of.BoolResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: of.ProviderResolutionDetail{
				Reason: of.DefaultReason,
			},
		}
	}

	if resp.Value != "" {
		bv, err := strconv.ParseBool(resp.Value)
		if err != nil {
			return of.BoolResolutionDetail{
				Value: defaultValue,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					ResolutionError: of.NewTypeMismatchResolutionError("value is not a boolean"),
					Reason:          of.DefaultReason,
				},
			}
		}

		return of.BoolResolutionDetail{
			Value: bv,
			ProviderResolutionDetail: of.ProviderResolutionDetail{
				Reason: of.DefaultReason,
			},
		}
	}

	return of.BoolResolutionDetail{
		Value: true,
		ProviderResolutionDetail: of.ProviderResolutionDetail{
			Reason: of.DefaultReason,
		},
	}
}

// StringEvaluation returns a string flag.
func (p Provider) StringEvaluation(ctx context.Context, flag string, defaultValue string, evalCtx of.FlattenedContext) of.StringResolutionDetail {
	// TODO: we have to check if the flag is enabled here until https://github.com/flipt-io/flipt/issues/1060 is resolved
	f, res, err := p.getFlag(ctx, flag)
	if err != nil {
		return of.StringResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: res,
		}
	}

	if !f.Enabled {
		return of.StringResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: of.ProviderResolutionDetail{
				Reason: of.DisabledReason,
			},
		}
	}

	resp, err := p.svc.Evaluate(ctx, flag, evalCtx)
	if err != nil {
		var (
			rerr   of.ResolutionError
			detail = of.StringResolutionDetail{
				Value: defaultValue,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason: of.DefaultReason,
				},
			}
		)

		if errors.As(err, &rerr) {
			detail.ProviderResolutionDetail.ResolutionError = rerr

			return detail
		}

		detail.ProviderResolutionDetail.ResolutionError = of.NewGeneralResolutionError(err.Error())

		return detail
	}

	if !resp.Match {
		return of.StringResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: of.ProviderResolutionDetail{
				Reason: of.DefaultReason,
			},
		}
	}

	return of.StringResolutionDetail{
		Value: resp.Value,
		ProviderResolutionDetail: of.ProviderResolutionDetail{
			Reason: of.TargetingMatchReason,
		},
	}
}

// FloatEvaluation returns a float flag.
func (p Provider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, evalCtx of.FlattenedContext) of.FloatResolutionDetail {
	// TODO: we have to check if the flag is enabled here until https://github.com/flipt-io/flipt/issues/1060 is resolved
	f, res, err := p.getFlag(ctx, flag)
	if err != nil {
		return of.FloatResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: res,
		}
	}

	if !f.Enabled {
		return of.FloatResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: of.ProviderResolutionDetail{
				Reason: of.DisabledReason,
			},
		}
	}

	resp, err := p.svc.Evaluate(ctx, flag, evalCtx)
	if err != nil {
		var (
			rerr   of.ResolutionError
			detail = of.FloatResolutionDetail{
				Value: defaultValue,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason: of.DefaultReason,
				},
			}
		)

		if errors.As(err, &rerr) {
			detail.ProviderResolutionDetail.ResolutionError = rerr

			return detail
		}

		detail.ProviderResolutionDetail.ResolutionError = of.NewGeneralResolutionError(err.Error())

		return detail
	}

	if !resp.Match {
		return of.FloatResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: of.ProviderResolutionDetail{
				Reason: of.DefaultReason,
			},
		}
	}

	fv, err := strconv.ParseFloat(resp.Value, 64)
	if err != nil {
		return of.FloatResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: of.ProviderResolutionDetail{
				ResolutionError: of.NewTypeMismatchResolutionError("value is not a float"),
				Reason:          of.ErrorReason,
			},
		}
	}

	return of.FloatResolutionDetail{
		Value: fv,
		ProviderResolutionDetail: of.ProviderResolutionDetail{
			Reason: of.TargetingMatchReason,
		},
	}
}

// IntEvaluation returns an int flag.
func (p Provider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, evalCtx of.FlattenedContext) of.IntResolutionDetail {
	// TODO: we have to check if the flag is enabled here until https://github.com/flipt-io/flipt/issues/1060 is resolved
	f, res, err := p.getFlag(ctx, flag)
	if err != nil {
		return of.IntResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: res,
		}
	}

	if !f.Enabled {
		return of.IntResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: of.ProviderResolutionDetail{
				Reason: of.DisabledReason,
			},
		}
	}

	resp, err := p.svc.Evaluate(ctx, flag, evalCtx)
	if err != nil {
		var (
			rerr   of.ResolutionError
			detail = of.IntResolutionDetail{
				Value: defaultValue,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason: of.DefaultReason,
				},
			}
		)

		if errors.As(err, &rerr) {
			detail.ProviderResolutionDetail.ResolutionError = rerr

			return detail
		}

		detail.ProviderResolutionDetail.ResolutionError = of.NewGeneralResolutionError(err.Error())

		return detail
	}

	if !resp.Match {
		return of.IntResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: of.ProviderResolutionDetail{
				Reason: of.DefaultReason,
			},
		}
	}

	iv, err := strconv.ParseInt(resp.Value, 10, 64)
	if err != nil {
		return of.IntResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: of.ProviderResolutionDetail{
				ResolutionError: of.NewTypeMismatchResolutionError("value is not an integer"),
				Reason:          of.ErrorReason,
			},
		}
	}

	return of.IntResolutionDetail{
		Value: iv,
		ProviderResolutionDetail: of.ProviderResolutionDetail{
			Reason: of.TargetingMatchReason,
		},
	}
}

// ObjectEvaluation returns an object flag with attachment if any. Value is a map of key/value pairs ([string]interface{}).
func (p Provider) ObjectEvaluation(ctx context.Context, flag string, defaultValue interface{}, evalCtx of.FlattenedContext) of.InterfaceResolutionDetail {
	// TODO: we have to check if the flag is enabled here until https://github.com/flipt-io/flipt/issues/1060 is resolved
	f, res, err := p.getFlag(ctx, flag)
	if err != nil {
		return of.InterfaceResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: res,
		}
	}

	if !f.Enabled {
		return of.InterfaceResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: of.ProviderResolutionDetail{
				Reason: of.DisabledReason,
			},
		}
	}

	resp, err := p.svc.Evaluate(ctx, flag, evalCtx)
	if err != nil {
		var (
			rerr   of.ResolutionError
			detail = of.InterfaceResolutionDetail{
				Value: defaultValue,
				ProviderResolutionDetail: of.ProviderResolutionDetail{
					Reason: of.DefaultReason,
				},
			}
		)

		if errors.As(err, &rerr) {
			detail.ProviderResolutionDetail.ResolutionError = rerr

			return detail
		}

		detail.ProviderResolutionDetail.ResolutionError = of.NewGeneralResolutionError(err.Error())

		return detail
	}

	if !resp.Match {
		return of.InterfaceResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: of.ProviderResolutionDetail{
				Reason: of.DefaultReason,
			},
		}
	}

	if resp.Attachment == "" {
		return of.InterfaceResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: of.ProviderResolutionDetail{
				Reason:  of.DefaultReason,
				Variant: resp.Value,
			},
		}
	}

	out := new(structpb.Struct)
	if err := protojson.Unmarshal([]byte(resp.Attachment), out); err != nil {
		return of.InterfaceResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: of.ProviderResolutionDetail{
				ResolutionError: of.NewTypeMismatchResolutionError(fmt.Sprintf("value is not an object: %q", resp.Attachment)),
				Reason:          of.ErrorReason,
			},
		}
	}

	return of.InterfaceResolutionDetail{
		Value: out.AsMap(),
		ProviderResolutionDetail: of.ProviderResolutionDetail{
			Reason:  of.TargetingMatchReason,
			Variant: resp.Value,
		},
	}
}

// Hooks returns hooks.
func (p Provider) Hooks() []of.Hook {
	// code to retrieve hooks
	return []of.Hook{}
}
