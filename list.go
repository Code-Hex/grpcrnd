package grpcrnd

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

type reflectClient struct {
	client      *grpcreflect.Client
	stub        grpcdynamic.Stub
	marshaler   *jsonpb.Marshaler
	unmarshaler *jsonpb.Unmarshaler
}

// NewReflectionGRPCClient returns grpc reflection client.
func NewReflectionGRPCClient(conn *grpc.ClientConn) *reflectClient {
	return &reflectClient{
		client: grpcreflect.NewClient(
			context.Background(),
			grpc_reflection_v1alpha.NewServerReflectionClient(conn),
		),
		stub: grpcdynamic.NewStub(conn),
		marshaler: &jsonpb.Marshaler{
			OrigName:     true,
			EmitDefaults: true,
		},
		unmarshaler: &jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		},
	}
}

func (c *reflectClient) ListServices() error {
	const reflectionServiceName = "grpc.reflection.v1alpha.ServerReflection"
	ssvcs, err := c.client.ListServices()
	if err != nil {
		return errors.Wrap(err, "failed to get services list")
	}

	for _, s := range ssvcs {
		if s == reflectionServiceName {
			continue
		}
		svc, err := c.client.ResolveService(s)
		if err != nil {
			return errors.Wrap(err, "failed to resolve service")
		}
		for _, method := range svc.GetMethods() {
			fmt.Println(s + "." + method.GetName())
		}
	}
	return nil
}
