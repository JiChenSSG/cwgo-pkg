/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package promprovider

import (
	"net/http"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/semantic"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	defaultBuckets = []float64{5000, 10000, 25000, 50000, 100000, 250000, 500000, 1000000}
	retryBuckets   = []float64{0, 5, 10, 50, 100, 1000, 5000, 10000, 50000}
)

// Option opts for opentelemetry tracer provider
type Option interface {
	apply(cfg *config)
}

type option func(cfg *config)

func (fn option) apply(cfg *config) {
	fn(cfg)
}

type config struct {
	buckets  []float64
	serveMux *http.ServeMux
	registry *prometheus.Registry

	disableServer bool
	path          string

	serviceType semantic.ServiceType
}

func newConfig(opts []Option) *config {
	cfg := defaultConfig()

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return cfg
}

func defaultConfig() *config {
	return &config{
		buckets:       defaultBuckets,
		registry:      prometheus.NewRegistry(),
		serveMux:      http.DefaultServeMux,
		disableServer: false,
		serviceType:   semantic.Hertz,
		path:          "/prometheus",
	}
}

// WithRegistry define your custom registry
func WithRegistry(registry *prometheus.Registry) Option {
	return option(func(cfg *config) {
		if registry != nil {
			cfg.registry = registry
		}
	})
}

// WithServeMux define your custom serve mux
func WithServeMux(serveMux *http.ServeMux) Option {
	return option(func(cfg *config) {
		if serveMux != nil {
			cfg.serveMux = serveMux
		}
	})
}

// WithDisableServer disable prometheus server
func WithDisableServer(disable bool) Option {
	return option(func(cfg *config) {
		cfg.disableServer = disable
	})
}

func WithPath(path string) Option {
	return option(func(cfg *config) {
		cfg.path = path
	})
}

func WithHttpServer() Option {
	return option(func(cfg *config) {
		cfg.serviceType = semantic.Hertz
	})
}

func WithRPCServer() Option {
	return option(func(cfg *config) {
		cfg.serviceType = semantic.Kitex
	})
}

// WithHistogramBuckets define your custom histogram buckets base on your biz
func WithHistogramBuckets(buckets []float64) Option {
	return option(func(cfg *config) {
		if len(buckets) > 0 {
			cfg.buckets = buckets
		}
	})
}
