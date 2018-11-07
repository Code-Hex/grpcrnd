package reflect

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"

	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

// Client struct.
type Client struct {
	client *grpcreflect.Client
	stub   grpcdynamic.Stub
}

// NewGRPCClient returns grpc reflection client.
func NewGRPCClient(ctx context.Context, conn *grpc.ClientConn) *Client {
	return &Client{
		client: grpcreflect.NewClient(ctx,
			grpc_reflection_v1alpha.NewServerReflectionClient(conn),
		),
		stub: grpcdynamic.NewStub(conn),
	}
}

// ListServices returns services information from prvided gRPC server.
func (c *Client) ListServices() ([]string, error) {
	return c.client.ListServices()
}

// ResolveService asks the server to resolve the given fully-qualified service
// name into a service descriptor.
func (c *Client) ResolveService(serviceName string) (*desc.ServiceDescriptor, error) {
	return c.client.ResolveService(serviceName)
}

// InvokeRPC sends a unary RPC and returns the response. Use this for unary methods.
func (c *Client) InvokeRPC(ctx context.Context, method *desc.MethodDescriptor, request proto.Message, opts ...grpc.CallOption) (proto.Message, error) {
	return c.stub.InvokeRpc(ctx, method, request, opts...)
}
