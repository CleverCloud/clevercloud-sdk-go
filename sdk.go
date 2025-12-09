// Package sdk provides a comprehensive Go SDK for the Clever Cloud API
package sdk

import (
	"go.clever-cloud.dev/client"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("go.clever-cloud.dev/sdk")

// SDK defines the interface for the Clever Cloud SDK
type SDK interface {
	// Level 2 API - Builder pattern interface
	V2() V2Builder
	V4() V4Builder
}

// sdkImpl is the concrete implementation of the SDK interface
type sdkImpl struct {
	client *client.Client
	tracer trace.Tracer
	v2     V2Builder
	v4     V4Builder
}

// Option defines configuration options for the SDK
type Option func(*sdkImpl)

// WithClient sets a custom Clever Cloud client
func WithClient(c *client.Client) Option {
	return func(s *sdkImpl) {
		s.client = c
	}
}

// WithTracer sets a custom OpenTelemetry tracer
func WithTracer(t trace.Tracer) Option {
	return func(s *sdkImpl) {
		s.tracer = t
	}
}

// NewSDK creates a new instance of the Clever Cloud SDK
func NewSDK(opts ...Option) SDK {

	impl := &sdkImpl{
		client: client.New(client.WithAutoAuthConfig()),
		tracer: tracer,
	}

	// Apply options
	for _, opt := range opts {
		opt(impl)
	}

	impl.v2 = newV2Builder(impl)
	impl.v4 = newV4Builder(impl)

	return impl
}

// V2 returns the V2 API builder
func (s *sdkImpl) V2() V2Builder { return s.v2 }

// V4 returns the V4 API builder
func (s *sdkImpl) V4() V4Builder { return s.v4 }

// Client returns the underlying Clever Cloud client (internal use only)
func (s *sdkImpl) Client() *client.Client { return s.client }

// Tracer returns the OpenTelemetry tracer (internal use only)
func (s *sdkImpl) Tracer() trace.Tracer { return s.tracer }
