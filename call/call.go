package call

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	_grpc "github.com/Code-Hex/grpcrnd/grpc"

	"github.com/Code-Hex/grpcrnd/reflect"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/k0kubun/pp"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type CommandRunner interface {
	Run() func(cmd *cobra.Command, args []string) error
	Command() *cobra.Command
}

type command struct {
	cmd         *cobra.Command
	insecure    *bool
	headers     []string
	uselog      bool
	marshaler   *jsonpb.Marshaler
	unmarshaler *jsonpb.Unmarshaler
}

func New(insecure *bool) CommandRunner {
	c := &command{
		cmd: &cobra.Command{
			Use:   "call <addr> <method>",
			Short: "call gRPC method using generated random parameter",
			Example: `
* call
grpcrnd call localhost:8888 test.Test.Echo

* call with header
grpcrnd call localhost:8888 test.Test.Echo -H 'UserAgent: grpcrand'
`,
			Args:         cobra.ExactArgs(2),
			SilenceUsage: true,
		},
		insecure: insecure,
		marshaler: &jsonpb.Marshaler{
			OrigName:     true,
			EmitDefaults: true,
		},
		unmarshaler: &jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		},
	}
	c.cmd.RunE = c.Run()
	c.cmd.Flags().StringArrayVarP(&c.headers, "header", "H", nil, "send with header")
	c.cmd.Flags().BoolVarP(&c.uselog, "log", "l", false, "specify if you want to output to logs")
	return c
}

func (c *command) Command() *cobra.Command { return c.cmd }

func (c *command) Run() func(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	return func(cmd *cobra.Command, args []string) error {
		conn, err := _grpc.NewClientConnection(ctx, args[0], *c.insecure)
		if err != nil {
			return errors.Wrap(err, "failed to make a gRPC connection")
		}
		defer conn.Close()
		client := reflect.NewGRPCClient(ctx, conn)
		if err := c.Call(client, args[1]); err != nil {
			return errors.Wrap(err, "failed to call gRPC method")
		}
		return nil
	}
}

func detectServiceMethod(reflectionMethod string) (string, string, error) {
	n := strings.LastIndex(reflectionMethod, ".")
	if n < 0 {
		return "", "", errors.Errorf("invalid reflection method name: %s", reflectionMethod)
	}
	service := reflectionMethod[0:n]
	method := reflectionMethod[n+1:]
	return service, method, nil
}

func (c *command) Call(client *reflect.Client, reflectionMethod string) error {
	service, method, err := detectServiceMethod(reflectionMethod)
	if err != nil {
		return errors.Wrap(err, "unexpected format")
	}
	svc, err := client.ResolveService(service)
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

	ctx := metadata.NewOutgoingContext(context.Background(), buildOutgoingMetadata(c.headers))

	var headerMD metadata.MD
	var trailerMD metadata.MD
	resp, err := client.InvokeRPC(ctx, mdesc, msg, grpc.Header(&headerMD), grpc.Trailer(&trailerMD))
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

	if err := c.output(respJSON); err != nil {
		return errors.Wrap(err, "failed to write log")
	}
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

func (c *command) createMessage(mdesc *desc.MethodDescriptor) (*dynamic.Message, error) {
	msg := dynamic.NewMessage(mdesc.GetInputType())
	m := retriveFields(msg.GetKnownFields())
	b, err := json.MarshalIndent(&m, "", "    ")
	if err != nil {
		return nil, err
	}
	f, err := os.Create("param-rc.json")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create rc file")
	}
	defer f.Close()
	f.Write(b)
	param, err := json.Marshal(&m)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create param json")
	}
	if err := msg.UnmarshalJSONPB(c.unmarshaler, param); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal to protobuf json")
	}
	return msg, nil
}

const times = 5

func retriveFields(fields []*desc.FieldDescriptor) map[string]interface{} {
	r := NewRand()
	m := make(map[string]interface{}, 0)
	for _, field := range fields {
		key := field.GetJSONName()
		isRepeated := field.IsRepeated()
		// https://github.com/golang/protobuf/blob/157d9c53be5810dd5a0fac4a467f7d5f400042ea/protoc-gen-go/descriptor/descriptor.pb.go#L51-L81
		switch *field.GetType().Enum() {
		case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
			if isRepeated {
				n := r.pickupNum(times)
				s := make([]float64, n)
				for i := 0; i < n; i++ {
					s[i] = r.double()
				}
				m[key] = s
			} else {
				m[key] = r.double()
			}
		case descriptor.FieldDescriptorProto_TYPE_FLOAT:
			if isRepeated {
				n := r.pickupNum(times)
				s := make([]float32, n)
				for i := 0; i < n; i++ {
					s[i] = r.float()
				}
				m[key] = s
			} else {
				m[key] = r.float()
			}
		case descriptor.FieldDescriptorProto_TYPE_UINT32:
			if isRepeated {
				n := r.pickupNum(times)
				s := make([]uint32, n)
				for i := 0; i < n; i++ {
					s[i] = r.uint32()
				}
				m[key] = s
			} else {
				m[key] = r.uint32()
			}
		case descriptor.FieldDescriptorProto_TYPE_UINT64:
			if isRepeated {
				n := r.pickupNum(times)
				s := make([]uint64, n)
				for i := 0; i < n; i++ {
					s[i] = r.uint64()
				}
				m[key] = s
			} else {
				m[key] = r.uint64()
			}
		case descriptor.FieldDescriptorProto_TYPE_INT32,
			descriptor.FieldDescriptorProto_TYPE_FIXED32,
			descriptor.FieldDescriptorProto_TYPE_SFIXED32,
			descriptor.FieldDescriptorProto_TYPE_SINT32:
			if isRepeated {
				n := r.pickupNum(times)
				s := make([]int32, n)
				for i := 0; i < n; i++ {
					s[i] = r.int32()
				}
				m[key] = s
			} else {
				m[key] = r.int32()
			}
		case descriptor.FieldDescriptorProto_TYPE_INT64,
			descriptor.FieldDescriptorProto_TYPE_FIXED64,
			descriptor.FieldDescriptorProto_TYPE_SFIXED64,
			descriptor.FieldDescriptorProto_TYPE_SINT64:
			if isRepeated {
				n := r.pickupNum(times)
				s := make([]int64, n)
				for i := 0; i < n; i++ {
					s[i] = r.int64()
				}
				m[key] = s
			} else {
				m[key] = r.int64()
			}
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			if isRepeated {
				n := r.pickupNum(times)
				s := make([]bool, n)
				for i := 0; i < n; i++ {
					s[i] = r.bool()
				}
				m[key] = s
			} else {
				m[key] = r.bool()
			}
		case descriptor.FieldDescriptorProto_TYPE_BYTES:
			if isRepeated {
				n := r.pickupNum(times)
				s := make([][]byte, n)
				for i := 0; i < n; i++ {
					s[i] = r.bytes()
				}
				m[key] = s
			} else {
				m[key] = r.bytes()
			}
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			if isRepeated {
				n := r.pickupNum(times)
				s := make([]string, n)
				for i := 0; i < n; i++ {
					s[i] = r.string()
				}
				m[key] = s
			} else {
				m[key] = r.string()
			}
		// Group is deprecated in proto3.
		// case descriptor.FieldDescriptorProto_TYPE_GROUP:
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			msg := field.GetMessageType()
			if isRepeated {
				n := r.pickupNum(times)
				s := make([]map[string]interface{}, n)
				for i := 0; i < n; i++ {
					s[i] = retriveFields(msg.GetFields())
				}
				m[key] = s
			} else {
				m[key] = retriveFields(msg.GetFields())
			}
		case descriptor.FieldDescriptorProto_TYPE_ENUM:
			enum := field.GetEnumType().GetValues()
			num := len(enum)
			if isRepeated {
				n := r.pickupNum(times)
				s := make([]int32, n)
				for i := 0; i < n; i++ {
					idx := r.pickupNum(num)
					s[i] = enum[idx].GetNumber()
				}
				m[key] = s
			} else {
				idx := r.pickupNum(num)
				m[key] = enum[idx].GetNumber()
			}
		default:
			// TODO: oneof ...???
		}
	}
	return m
}
