package list

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/Code-Hex/grpcrnd/grpc"
	"github.com/Code-Hex/grpcrnd/reflect"
	"github.com/spf13/cobra"

	"github.com/pkg/errors"
)

type CommandRunner interface {
	Run() func(cmd *cobra.Command, args []string) error
	Command() *cobra.Command
}

type command struct {
	cmd      *cobra.Command
	insecure *bool
}

func New(insecure *bool) CommandRunner {
	c := &command{
		cmd: &cobra.Command{
			Use:   "ls <addr> [options]",
			Short: "list methods (included services) provided by gRPC server",
			Example: `
* List services and method
grpcurl ls localhost:8888
`,
			Aliases:      []string{"ls", "l"},
			Args:         cobra.ExactArgs(1),
			SilenceUsage: true,
		},
		insecure: insecure,
	}
	c.cmd.RunE = c.Run()
	return c
}

func (c *command) Command() *cobra.Command { return c.cmd }

func (c *command) Run() func(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	return func(cmd *cobra.Command, args []string) error {
		conn, err := grpc.NewClientConnection(ctx, args[0], *c.insecure)
		if err != nil {
			return errors.Wrap(err, "failed to make a gRPC connection")
		}
		defer conn.Close()
		client := reflect.NewGRPCClient(ctx, conn)
		if err := List(client); err != nil {
			return errors.Wrap(err, "failed to list servieces and methods")
		}
		return nil
	}
}

const reflectionServiceName = "grpc.reflection.v1alpha.ServerReflection"

func List(client *reflect.Client) error {
	ssvcs, err := client.ListServices()
	if err != nil {
		return errors.Wrap(err, "failed to get services list")
	}
	var buf bytes.Buffer
	for _, s := range ssvcs {
		if s == reflectionServiceName {
			continue
		}
		svc, err := client.ResolveService(s)
		if err != nil {
			return errors.Wrap(err, "failed to resolve service")
		}
		for _, method := range svc.GetMethods() {
			fmt.Fprintln(&buf, s+"."+method.GetName())
		}
	}
	io.Copy(os.Stdout, &buf)
	return nil
}
