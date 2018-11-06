package grpcrnd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/k0kubun/pp"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (c *reflectClient) Call(headers []string, reflectionMethod string) error {
	n := strings.LastIndex(reflectionMethod, ".")
	if n < 0 {
		return errors.Errorf("invalid reflection method name: %s", reflectionMethod)
	}
	service := reflectionMethod[0:n]
	method := reflectionMethod[n+1:]
	svc, err := c.client.ResolveService(service)
	if err != nil {
		return errors.Wrapf(err, "failed to resolve service %s", service)
	}
	mdesc := svc.FindMethodByName(method)
	if mdesc == nil {
		return errors.New("method couldn't be found")
	}
	msg, err := c.createMessage(mdesc)
	if err != nil {
		return errors.Wrap(err, "failed to create message")
	}

	// NOTE: DEBUG
	if false {
		pp.Println(reflectionMethod)
		reqJSON, err := msg.MarshalJSONPB(c.marshaler)
		if err != nil {
			return err
		}
		pp.Println(string(reqJSON))
	}

	ctx := metadata.NewOutgoingContext(context.Background(), buildOutgoingMetadata(headers))

	var headerMD metadata.MD
	var trailerMD metadata.MD
	resp, err := c.stub.InvokeRpc(ctx, mdesc, msg, grpc.Header(&headerMD), grpc.Trailer(&trailerMD))
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return errors.Wrap(err, "failed to get error from proto")
		}
		resp = st.Proto()
	}

	respJSON, err := c.marshaler.MarshalToString(resp)
	if err != nil {
		return errors.Wrap(err, "failed to marshal json response")
	}

	fmt.Println(respJSON)
	return nil
}

func buildOutgoingMetadata(header []string) metadata.MD {
	var pairs []string
	for i := range header {
		parts := strings.SplitN(header[i], ":", 2)
		if len(parts) < 2 {
			continue
		}
		k, v := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		pairs = append(pairs, k, v)
	}
	return metadata.Pairs(pairs...)
}

func (c *reflectClient) createMessage(mdesc *desc.MethodDescriptor) (*dynamic.Message, error) {
	msg := dynamic.NewMessage(mdesc.GetInputType())
	m := c.retriveFields(msg.GetKnownFields())
	param, err := json.Marshal(&m)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create param json")
	}
	if err := msg.UnmarshalJSONPB(c.unmarshaler, param); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal to protobuf json")
	}
	return msg, nil
}

func (c *reflectClient) retriveFields(fields []*desc.FieldDescriptor) map[string]interface{} {
	m := make(map[string]interface{}, 0)
	for _, field := range fields {
		key := field.GetJSONName()
		r := NewRand()
		// https://github.com/golang/protobuf/blob/157d9c53be5810dd5a0fac4a467f7d5f400042ea/protoc-gen-go/descriptor/descriptor.pb.go#L51-L81
		switch *field.GetType().Enum() {
		case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
			m[key] = r.double()
		case descriptor.FieldDescriptorProto_TYPE_FLOAT:
			m[key] = r.float()
		case descriptor.FieldDescriptorProto_TYPE_UINT32:
			m[key] = r.uint32()
		case descriptor.FieldDescriptorProto_TYPE_UINT64:
			m[key] = r.uint64()
		case descriptor.FieldDescriptorProto_TYPE_INT32,
			descriptor.FieldDescriptorProto_TYPE_FIXED32,
			descriptor.FieldDescriptorProto_TYPE_SFIXED32,
			descriptor.FieldDescriptorProto_TYPE_SINT32:
			m[key] = r.int32()
		case descriptor.FieldDescriptorProto_TYPE_INT64,
			descriptor.FieldDescriptorProto_TYPE_FIXED64,
			descriptor.FieldDescriptorProto_TYPE_SFIXED64,
			descriptor.FieldDescriptorProto_TYPE_SINT64:
			m[key] = r.int64()
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			m[key] = r.bool()
		case descriptor.FieldDescriptorProto_TYPE_BYTES:
			m[key] = r.bytes()
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			m[key] = r.string()
		// Group is deprecated in proto3.
		// case descriptor.FieldDescriptorProto_TYPE_GROUP:
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			msg := field.GetMessageType()
			m[key] = c.retriveFields(msg.GetFields())
		case descriptor.FieldDescriptorProto_TYPE_ENUM:
			enum := field.GetEnumType().GetValues()
			num := len(enum)
			idx := r.pickupEnum(num)
			m[key] = enum[idx].GetNumber()
		default:
			// TODO: oneof ...???
		}
	}
	return m
}
