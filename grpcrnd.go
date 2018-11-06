package grpcrnd

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// GRPCRand represents grpc random.
type GRPCRand struct {
	*Options
}

// New returns GRPCRand struct pointer
func New() *GRPCRand {
	return &GRPCRand{
		Options: &Options{},
	}
}

// Run method will create a project and returns exit code
func (g *GRPCRand) Run() int {
	if e := g.run(); e != nil {
		exitCode, err := UnwrapErrors(e)
		if err != nil {
			if g.StackTrace {
				fmt.Fprintf(os.Stderr, "Error: %+v\n", e)
			} else {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			return exitCode
		}
	}
	return 0
}

func (g *GRPCRand) prepare() error {
	_, err := parseOptions(g.Options, os.Args[1:])
	if err != nil {
		return errors.Wrap(err, "failed to parse options")
	}
	return nil
}

func (g *GRPCRand) run() error {
	if err := g.prepare(); err != nil {
		return errors.Wrap(err, "failed to setup")
	}
	conn, err := NewGRPCConnection(
		context.Background(),
		"",
		true,
	)
	if err != nil {
		return err
	}
	client := NewReflectionGRPCClient(conn)
	svcs, err := client.ListServices()
	if err != nil {
		return errors.Wrap(err, "failed to invoke ListServices")
	}
	client.Call(g.Header, svcs[1])
	return nil
}

func parseOptions(opts *Options, argv []string) ([]string, error) {
	o, err := opts.parse(argv)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse arguments")
	}
	if opts.Help {
		return nil, makeUsageError(errors.New(opts.usage()))
	}
	return o, nil
}
