// Copyright 2022 CloudWeGo Authors.
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

package otelkitex

import (
	"context"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/semantic"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/metadata"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func injectPeerServiceToMetaInfo(ctx context.Context, attrs []attribute.KeyValue) map[string]string {
	md := metainfo.GetAllValues(ctx)
	if md == nil {
		md = make(map[string]string)
	}

	serviceName, serviceNamespace, deploymentEnv := getServiceFromResourceAttributes(attrs)

	if serviceName != "" {
		md[string(semconv.ServiceNameKey)] = serviceName
	}

	if serviceNamespace != "" {
		md[string(semconv.ServiceNamespaceKey)] = serviceNamespace
	}

	if deploymentEnv != "" {
		md[string(semconv.DeploymentEnvironmentKey)] = deploymentEnv
	}

	return md
}

func extractPeerServiceAttributesFromMetaInfo(md map[string]string) []attribute.KeyValue {
	var attrs []attribute.KeyValue

	for k, v := range md {
		switch k {
		case string(semconv.ServiceNameKey):
			attrs = append(attrs, semconv.PeerServiceKey.String(v))
		case string(semconv.ServiceNamespaceKey):
			attrs = append(attrs, semantic.PeerServiceNamespaceKey.String(v))
		case string(semconv.DeploymentEnvironmentKey):
			attrs = append(attrs, semantic.PeerDeploymentEnvironmentKey.String(v))
		}
	}

	return attrs
}

func extractPeerServiceAttributesFromMetadata(md metadata.MD) []attribute.KeyValue {
	var (
		attrs      []attribute.KeyValue
		mdSupplier = metadataSupplier{metadata: &md}
	)
	if v := mdSupplier.Get(string(semconv.ServiceNameKey)); v != "" {
		attrs = append(attrs, semconv.PeerServiceKey.String(v))
	}
	if v := mdSupplier.Get(string(semconv.ServiceNamespaceKey)); v != "" {
		attrs = append(attrs, semantic.PeerServiceNamespaceKey.String(v))
	}
	if v := mdSupplier.Get(string(semconv.DeploymentEnvironmentKey)); v != "" {
		attrs = append(attrs, semantic.PeerDeploymentEnvironmentKey.String(v))
	}
	return attrs
}
