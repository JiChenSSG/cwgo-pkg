// Code generated by Kitex v0.5.1. DO NOT EDIT.

package hello

import (
	"github.com/cloudwego-contrib/cwgo-pkg/registry/consul/consulkitex/example/hello/kitex_gen/api"
	server "github.com/cloudwego/kitex/server"
)

// NewInvoker creates a server.Invoker with the given handler and options.
func NewInvoker(handler api.Hello, opts ...server.Option) server.Invoker {
	var options []server.Option

	options = append(options, opts...)

	s := server.NewInvoker(options...)
	if err := s.RegisterService(serviceInfo(), handler); err != nil {
		panic(err)
	}
	if err := s.Init(); err != nil {
		panic(err)
	}
	return s
}
