package proto

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"math"
)

import (
	"context"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/server"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ client.Option
var _ server.Option

// Client API for Hello service

type HelloService interface {
	Hello(ctx context.Context, in *HelloRequest, opts ...client.CallOption) (*HelloResponse, error)
}

type helloService struct {
	c    client.Client
	name string
}

func NewHelloService(name string, c client.Client) HelloService {
	if c == nil {
		c = client.NewClient()
	}
	if len(name) == 0 {
		name = "hello"
	}
	return &helloService{
		c:    c,
		name: name,
	}
}

func (c *helloService) Hello(ctx context.Context, in *HelloRequest, opts ...client.CallOption) (*HelloResponse, error) {
	req := c.c.NewRequest(c.name, "SayHello.Hello", in)
	resp := new(HelloResponse)
	err := c.c.Call(ctx, req, resp, opts...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Server API for Hello service

type HelloHandler interface {
	Hello(context.Context, *HelloRequest, *HelloResponse) error
}

func RegisterHelloHandler(s server.Server, hdlr HelloHandler, opts ...server.HandlerOption) error {
	return s.Handle(s.NewHandler(hdlr, opts...))
}

