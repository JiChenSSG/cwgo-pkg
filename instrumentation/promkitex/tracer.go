/*
 * Copyright 2021 CloudWeGo Authors
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

// Package prometheus provides the extend implement of prometheus.
package prometheus

import (
	"context"
	"github.com/cloudwego-contrib/cwgo-pkg/meter/label"
	cwmetric "github.com/cloudwego-contrib/cwgo-pkg/meter/metric"
	"github.com/cloudwego-contrib/cwgo-pkg/semantic"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/stats"
)

// Labels
const (
	labelKeyCaller = semantic.LabelKeyCaller
	labelKeyMethod = semantic.LabelMethodProm
	labelKeyCallee = semantic.LabelKeyCallee
	labelKeyStatus = semantic.LabelKeyStatus
	labelKeyRetry  = semantic.LabelKeyRetry

	// status
	statusSucceed = "succeed"
	statusError   = "error"

	unknownLabelValue = "unknown"
)

// genLabels make labels values.
func genLabels(ri rpcinfo.RPCInfo) prom.Labels {
	var (
		labels = make(prom.Labels)

		caller = ri.From()
		callee = ri.To()
	)
	labels[labelKeyCaller] = defaultValIfEmpty(caller.ServiceName(), unknownLabelValue)
	labels[labelKeyCallee] = defaultValIfEmpty(callee.ServiceName(), unknownLabelValue)
	labels[labelKeyMethod] = defaultValIfEmpty(callee.Method(), unknownLabelValue)

	labels[labelKeyStatus] = statusSucceed
	if ri.Stats().Error() != nil {
		labels[labelKeyStatus] = statusError
	}

	labels[labelKeyRetry] = "0"
	if retriedCnt, ok := callee.Tag(rpcinfo.RetryTag); ok {
		labels[labelKeyRetry] = retriedCnt
	}

	return labels
}

func genCwLabels(ri rpcinfo.RPCInfo) []label.CwLabel {
	labels := genLabels(ri)
	return label.ToCwLabelFromPromelabel(labels)
}

type clientTracer struct {
	clientHandledCounter   *prom.CounterVec
	clientHandledHistogram *prom.HistogramVec
	promMetric             *cwmetric.PrometheusMetrics
}

// Start record the beginning of an RPC invocation.
func (c *clientTracer) Start(ctx context.Context) context.Context {
	return ctx
}

// Finish record after receiving the response of server.
func (c *clientTracer) Finish(ctx context.Context) {
	ri := rpcinfo.GetRPCInfo(ctx)
	if ri.Stats().Level() == stats.LevelDisabled {
		return
	}
	rpcStart := ri.Stats().GetEvent(stats.RPCStart)
	rpcFinish := ri.Stats().GetEvent(stats.RPCFinish)
	cost := rpcFinish.Time().Sub(rpcStart.Time())

	extraLabels := make(prom.Labels)
	extraLabels[labelKeyStatus] = statusSucceed
	if ri.Stats().Error() != nil {
		extraLabels[labelKeyStatus] = statusError
	}
	c.promMetric.Inc(ctx, genCwLabels(ri))
	c.promMetric.Record(ctx, float64(cost.Microseconds()), genCwLabels(ri))
}

// NewClientTracer provide tracer for client call, addr and path is the scrape_configs for prometheus server.
func NewClientTracer(addr, path string, options ...Option) stats.Tracer {
	cfg := defaultConfig()
	for _, opt := range options {
		opt.apply(cfg)
	}

	if !cfg.disableServer {
		cfg.serveMux.Handle(path, promhttp.HandlerFor(cfg.registry, promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
			Registry:      cfg.registry,
		}))
		go func() {
			if err := http.ListenAndServe(addr, cfg.serveMux); err != nil {
				log.Fatal("Unable to start a promhttp server, err: " + err.Error())
			}
		}()
	}

	clientHandledCounter := prom.NewCounterVec(
		prom.CounterOpts{
			Name: semantic.ClientThroughput,
			Help: "Total number of RPCs completed by the client, regardless of success or failure.",
		},
		[]string{labelKeyCaller, labelKeyCallee, labelKeyMethod, labelKeyStatus, labelKeyRetry},
	)
	cfg.registry.MustRegister(clientHandledCounter)

	clientHandledHistogram := prom.NewHistogramVec(
		prom.HistogramOpts{
			Name:    semantic.ClientDuration,
			Help:    "Latency (microseconds) of the RPC until it is finished.",
			Buckets: cfg.buckets,
		},
		[]string{labelKeyCaller, labelKeyCallee, labelKeyMethod, labelKeyStatus, labelKeyRetry},
	)
	cfg.registry.MustRegister(clientHandledHistogram)
	if cfg.enableGoCollector {
		cfg.registry.MustRegister(collectors.NewGoCollector(collectors.WithGoCollectorRuntimeMetrics(cfg.runtimeMetricRules...)))
	}
	promMetric := cwmetric.NewPrometheusMetrics(clientHandledCounter, clientHandledHistogram)
	return &clientTracer{
		promMetric: promMetric,
	}
}

type serverTracer struct {
	promMetric *cwmetric.PrometheusMetrics
}

// Start record the beginning of server handling request from client.
func (c *serverTracer) Start(ctx context.Context) context.Context {
	return ctx
}

// Finish record the ending of server handling request from client.
func (c *serverTracer) Finish(ctx context.Context) {
	ri := rpcinfo.GetRPCInfo(ctx)
	if ri.Stats().Level() == stats.LevelDisabled {
		return
	}

	rpcStart := ri.Stats().GetEvent(stats.RPCStart)
	rpcFinish := ri.Stats().GetEvent(stats.RPCFinish)
	cost := rpcFinish.Time().Sub(rpcStart.Time())

	extraLabels := make(prom.Labels)
	extraLabels[labelKeyStatus] = statusSucceed
	if ri.Stats().Error() != nil {
		extraLabels[labelKeyStatus] = statusError
	}

	c.promMetric.Inc(ctx, genCwLabels(ri))
	c.promMetric.Record(ctx, float64(cost.Microseconds()), genCwLabels(ri))
}

// NewServerTracer provides tracer for server access, addr and path is the scrape_configs for prometheus server.
func NewServerTracer(addr, path string, options ...Option) stats.Tracer {
	cfg := defaultConfig()
	for _, opt := range options {
		opt.apply(cfg)
	}

	if !cfg.disableServer {
		cfg.serveMux.Handle(path, promhttp.HandlerFor(cfg.registry, promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
			Registry:      cfg.registry,
		}))
		go func() {
			if err := http.ListenAndServe(addr, cfg.serveMux); err != nil {
				log.Fatal("Unable to start a promhttp server, err: " + err.Error())
			}
		}()
	}

	serverHandledCounter := prom.NewCounterVec(
		prom.CounterOpts{
			Name: "kitex_server_throughput",
			Help: "Total number of RPCs completed by the server, regardless of success or failure.",
		},
		[]string{labelKeyCaller, labelKeyCallee, labelKeyMethod, labelKeyStatus, labelKeyRetry},
	)
	cfg.registry.MustRegister(serverHandledCounter)

	serverHandledHistogram := prom.NewHistogramVec(
		prom.HistogramOpts{
			Name:    "kitex_server_latency_us",
			Help:    "Latency (microseconds) of RPC that had been application-level handled by the server.",
			Buckets: cfg.buckets,
		},
		[]string{labelKeyCaller, labelKeyCallee, labelKeyMethod, labelKeyStatus, labelKeyRetry},
	)
	cfg.registry.MustRegister(serverHandledHistogram)

	if cfg.enableGoCollector {
		cfg.registry.MustRegister(collectors.NewGoCollector(collectors.WithGoCollectorRuntimeMetrics(cfg.runtimeMetricRules...)))
	}
	promMetric := cwmetric.NewPrometheusMetrics(serverHandledCounter, serverHandledHistogram)
	return &serverTracer{
		promMetric: promMetric,
	}
}

func defaultValIfEmpty(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
