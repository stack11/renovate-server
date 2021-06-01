// +build !noperfhelper_tracing
// +build !noconfhelper_tracing

/*
Copyright 2020 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package perfhelper

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	otexporterotlp "go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	otexporterjaeger "go.opentelemetry.io/otel/exporters/trace/jaeger"
	otexporterzipkin "go.opentelemetry.io/otel/exporters/trace/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	otsdktrace "go.opentelemetry.io/otel/sdk/trace"
	otsemconv "go.opentelemetry.io/otel/semconv"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
)

func (c *TracingConfig) CreateIfEnabled(setGlobal bool, client *http.Client) (oteltrace.TracerProvider, error) {
	if !c.Enabled {
		return nil, nil
	}

	var exporter otsdktrace.SpanExporter

	traceProviderOpts := []otsdktrace.TracerProviderOption{
		otsdktrace.WithSampler(otsdktrace.TraceIDRatioBased(c.SampleRate)),
		otsdktrace.WithResource(resource.NewWithAttributes(
			otsemconv.ServiceNameKey.String(c.ServiceName),
		)),
	}

	tlsConfig, err := c.TLS.GetTLSConfig(true)
	if err != nil {
		return nil, fmt.Errorf("failed to create tls config: %w", err)
	}

	switch c.Format {
	case "otlp":
		opts := []otlpgrpc.Option{
			otlpgrpc.WithEndpoint(c.Endpoint),
		}

		if tlsConfig != nil {
			opts = append(opts, otlpgrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)))
		} else {
			opts = append(opts, otlpgrpc.WithInsecure())
		}

		exporter, err = otexporterotlp.NewExporter(context.Background(), otlpgrpc.NewDriver(opts...))
		if err != nil {
			return nil, fmt.Errorf("failed to create otlp exporter: %w", err)
		}
	case "zipkin":
		exporter, err = otexporterzipkin.NewRawExporter(c.Endpoint,
			otexporterzipkin.WithSDKOptions(
				otsdktrace.WithResource(resource.NewWithAttributes(
					otsemconv.ServiceNameKey.String(c.ServiceName),
				)),
			),
			otexporterzipkin.WithClient(c.newHTTPClient(client, tlsConfig)),
			otexporterzipkin.WithLogger(nil),
		)
	case "jaeger":
		var endpoint otexporterjaeger.EndpointOption
		switch c.EndpointType {
		case "agent":
			host, port, err2 := net.SplitHostPort(c.Endpoint)
			if err2 != nil {
				return nil, fmt.Errorf("invalid jaeger agent endpoint: %w", err2)
			}

			endpoint = otexporterjaeger.WithAgentEndpoint(
				otexporterjaeger.WithAgentHost(host),
				otexporterjaeger.WithAgentPort(port),
				otexporterjaeger.WithAttemptReconnectingInterval(5*time.Second),
			)
		case "collector":
			// use environ OTEL_EXPORTER_JAEGER_USER
			// and OTEL_EXPORTER_JAEGER_PASSWORD
			endpoint = otexporterjaeger.WithCollectorEndpoint(
				otexporterjaeger.WithEndpoint(c.Endpoint),
				otexporterjaeger.WithHTTPClient(c.newHTTPClient(client, tlsConfig)),
			)
		default:
			return nil, fmt.Errorf("unsupported tracing endpoint type %q", c.EndpointType)
		}

		exporter, err = otexporterjaeger.NewRawExporter(endpoint)
	default:
		return nil, fmt.Errorf("unsupported tracing format %q", c.Format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create %q exporter: %w", c.Format, err)
	}

	traceProviderOpts = append(traceProviderOpts,
		otsdktrace.WithBatcher(
			exporter,
			otsdktrace.WithBatchTimeout(5*time.Second),
		),
	)

	traceProvider := otsdktrace.NewTracerProvider(traceProviderOpts...)

	if setGlobal {
		otel.SetTracerProvider(traceProvider)
	}

	return traceProvider, nil
}

func (c *TracingConfig) newHTTPClient(client *http.Client, tlsConfig *tls.Config) *http.Client {
	if tlsConfig != nil {
		tlsConfig.NextProtos = []string{"h2", "http/1.1"}
	}

	if client == nil {
		client = &http.Client{
			// TODO: set reasonable defaults, currently using default client and transport
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:       30 * time.Second,
					KeepAlive:     30 * time.Second,
					FallbackDelay: 300 * time.Millisecond,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				TLSClientConfig:       tlsConfig,

				DisableKeepAlives:      false,
				DisableCompression:     false,
				MaxIdleConnsPerHost:    0,
				MaxConnsPerHost:        0,
				ResponseHeaderTimeout:  0,
				TLSNextProto:           nil,
				ProxyConnectHeader:     nil,
				MaxResponseHeaderBytes: 0,
				WriteBufferSize:        0,
				ReadBufferSize:         0,
			},
		}
	}

	return &http.Client{
		Transport:     client.Transport,
		CheckRedirect: client.CheckRedirect,
		Jar:           client.Jar,
		Timeout:       client.Timeout,
	}
}
