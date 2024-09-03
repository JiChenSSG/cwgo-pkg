// Copyright 2021 CloudWeGo Authors.
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

package nacos

import (
	"os"
	"strconv"
	"strings"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/common"

	"github.com/cloudwego/kitex/pkg/klog"
)

const (
	NACOS_ENV_TAGS = "KITEX_NACOS_ENV_TAGS"
)

// Tags providers the default tags to inject nacos.
var Tags map[string]string

func init() {
	Tags = parseTags(os.Getenv(NACOS_ENV_TAGS))
}

func parseTags(tags string) map[string]string {
	out := map[string]string{
		"cloudwego.nacos.client": "kitex",
	}
	if len(tags) == 0 {
		return out
	}
	parts := strings.Split(tags, ",")
	for _, part := range parts {
		tag := strings.Split(part, "=")
		if len(tag) == 2 {
			out[tag[0]] = tag[1]
		}
	}
	return out
}

// NacosPort Get Nacos port from environment variables
func NacosPort() int64 {
	portText := os.Getenv(common.NacosEnvServerPort)
	if len(portText) == 0 {
		return common.NacosDefaultPort
	}
	port, err := strconv.ParseInt(portText, 10, 64)
	if err != nil {
		klog.Errorf("ParseInt failed,err:%s", err.Error())
		return common.NacosDefaultPort
	}
	return port
}