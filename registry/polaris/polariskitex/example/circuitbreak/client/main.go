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

package main

import (
	"context"
	"log"
	"time"

	polaris "github.com/cloudwego-contrib/cwgo-pkg/registry/polaris/polariskitex"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api/hello"
	"github.com/cloudwego/kitex/client"
)

const (
	Namespace = "Polaris"
	// At present,polaris server tag is v1.4.0，can't support auto create namespace,
	// if you want to use a namespace other than default,Polaris ,before you register an instance,
	// you should create the namespace at polaris console first.
)

func main() {
	option := polaris.ClientOptions{}
	r, err := polaris.NewPolarisResolver(option)
	if err != nil {
		log.Fatal(err)
	}

	pb, err := polaris.NewPolarisBalancer()
	if err != nil {
		log.Fatal(err)
	}

	suite := &polaris.ClientSuite{
		DstNameSpace:       Namespace,
		Resolver:           r,
		Balancer:           pb,
		ReportCallResultMW: polaris.NewUpdateServiceCallResultMW(),
	}

	newClient := hello.MustNewClient("polaris.circuitbreak.echo",
		client.WithSuite(suite),
		client.WithRPCTimeout(time.Second*1),
	)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
		resp, err := newClient.Echo(ctx, &api.Request{Message: "Hi,polaris!"})
		cancel()
		if err != nil {
			log.Println(err)
		}
		log.Println(resp)
		time.Sleep(1 * time.Second)
	}
}
