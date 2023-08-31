// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otlptracegrpc // import "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"

import (
	"fmt"
	"git.ddxq.mobi/css-oss-internal/otlptracegrpc/retry"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"go.opentelemetry.io/otel"
)

// Option applies an option to the gRPC driver.
type Option interface {
	applyGRPCOption(Config) Config
}

func asGRPCOptions(opts []Option) []GRPCOption {
	converted := make([]GRPCOption, len(opts))
	for i, o := range opts {
		converted[i] = NewGRPCOption(o.applyGRPCOption)
	}
	return converted
}

// RetryConfig defines configuration for retrying export of span batches that
// failed to be received by the target endpoint.
//
// This configuration does not define any network retry strategy. That is
// entirely handled by the gRPC ClientConn.
type RetryConfig retry.Config

type wrappedOption struct {
	GRPCOption
}

func (w wrappedOption) applyGRPCOption(cfg Config) Config {
	return w.ApplyGRPCOption(cfg)
}

// WithInsecure disables client transport security for the exporter's gRPC
// connection just like grpc.WithInsecure()
// (https://pkg.go.dev/google.golang.org/grpc#WithInsecure) does. Note, by
// default, client security is required unless WithInsecure is used.
//
// This option has no effect if WithGRPCConn is used.
func WithInsecure() Option {
	return wrappedOption{newGenericOption(func(cfg Config) Config {
		cfg.Traces.Insecure = true
		return cfg
	})}
}

// WithEndpoint sets the target endpoint the exporter will connect to. If
// unset, localhost:4317 will be used as a default.
//
// This option has no effect if WithGRPCConn is used.
func WithEndpoint(endpoint string) Option {
	return wrappedOption{newGenericOption(func(cfg Config) Config {
		cfg.Traces.Endpoint = endpoint
		return cfg
	})}
}

// WithReconnectionPeriod set the minimum amount of time between connection
// attempts to the target endpoint.
//
// This option has no effect if WithGRPCConn is used.
func WithReconnectionPeriod(rp time.Duration) Option {
	return wrappedOption{NewGRPCOption(func(cfg Config) Config {
		cfg.ReconnectionPeriod = rp
		return cfg
	})}
}

func compressorToCompression(compressor string) Compression {
	switch compressor {
	case "gzip":
		return GzipCompression
	}

	otel.Handle(fmt.Errorf("invalid compression type: '%s', using no compression as default", compressor))
	return NoCompression
}

// WithCompressor sets the compressor for the gRPC client to use when sending
// requests. It is the responsibility of the caller to ensure that the
// compressor set has been registered with google.golang.org/grpc/encoding.
// This can be done by encoding.RegisterCompressor. Some compressors
// auto-register on import, such as gzip, which can be registered by calling
// `import _ "google.golang.org/grpc/encoding/gzip"`.
//
// This option has no effect if WithGRPCConn is used.
func WithCompressor(compressor string) Option {
	return wrappedOption{
		newGenericOption(func(cfg Config) Config {
			cfg.Traces.Compression = compressorToCompression(compressor)
			return cfg
		})}
}

// WithHeaders will send the provided headers with each gRPC requests.
func WithHeaders(headers map[string]string) Option {
	return wrappedOption{newGenericOption(func(cfg Config) Config {
		cfg.Traces.Headers = headers
		return cfg
	})}
}

// WithTLSCredentials allows the connection to use TLS credentials when
// talking to the server. It takes in grpc.TransportCredentials instead of say
// a Certificate file or a tls.Certificate, because the retrieving of these
// credentials can be done in many ways e.g. plain file, in code tls.Config or
// by certificate rotation, so it is up to the caller to decide what to use.
//
// This option has no effect if WithGRPCConn is used.
func WithTLSCredentials(creds credentials.TransportCredentials) Option {
	return wrappedOption{NewGRPCOption(func(cfg Config) Config {
		cfg.Traces.GRPCCredentials = creds
		return cfg
	})}
}

// WithServiceConfig defines the default gRPC service config used.
//
// This option has no effect if WithGRPCConn is used.
func WithServiceConfig(serviceConfig string) Option {
	return wrappedOption{NewGRPCOption(func(cfg Config) Config {
		cfg.ServiceConfig = serviceConfig
		return cfg
	})}
}

// WithDialOption sets explicit grpc.DialOptions to use when making a
// connection. The options here are appended to the internal grpc.DialOptions
// used so they will take precedence over any other internal grpc.DialOptions
// they might conflict with.
//
// This option has no effect if WithGRPCConn is used.
func WithDialOption(opts ...grpc.DialOption) Option {
	return wrappedOption{NewGRPCOption(func(cfg Config) Config {
		cfg.DialOptions = opts
		return cfg
	})}
}

// WithGRPCConn sets conn as the gRPC ClientConn used for all communication.
//
// This option takes precedence over any other option that relates to
// establishing or persisting a gRPC connection to a target endpoint. Any
// other option of those types passed will be ignored.
//
// It is the callers responsibility to close the passed conn. The client
// Shutdown method will not close this connection.
func WithGRPCConn(conn *grpc.ClientConn) Option {
	return wrappedOption{NewGRPCOption(func(cfg Config) Config {
		cfg.GRPCConn = conn
		return cfg
	})}
}

// WithTimeout sets the max amount of time a client will attempt to export a
// batch of spans. This takes precedence over any retry settings defined with
// WithRetry, once this time limit has been reached the export is abandoned
// and the batch of spans is dropped.
//
// If unset, the default timeout will be set to 10 seconds.
func WithTimeout(duration time.Duration) Option {
	return wrappedOption{newGenericOption(func(cfg Config) Config {
		cfg.Traces.Timeout = duration
		return cfg
	})}
}

type grpcOption struct {
	fn func(Config) Config
}

func (h *grpcOption) ApplyGRPCOption(cfg Config) Config {
	return h.fn(cfg)
}

func (grpcOption) private() {}

func NewGRPCOption(fn func(cfg Config) Config) GRPCOption {
	return &grpcOption{fn: fn}
}

type genericOption struct {
	fn func(Config) Config
}

func (g *genericOption) ApplyGRPCOption(cfg Config) Config {
	return g.fn(cfg)
}

func (g *genericOption) ApplyHTTPOption(cfg Config) Config {
	return g.fn(cfg)
}

func (genericOption) private() {}

func newGenericOption(fn func(cfg Config) Config) GenericOption {
	return &genericOption{fn: fn}
}
