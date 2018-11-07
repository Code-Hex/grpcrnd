package main

import (
	"fmt"
	"os"

	"github.com/Code-Hex/grpcrnd/call"
	"github.com/Code-Hex/grpcrnd/list"
	"github.com/spf13/cobra"
)

// GRPCRand represents grpc random.
type GRPCRand struct {
	*command
}

// Options struct for parse command line arguments
type Options struct {
	Insecure   bool
	StackTrace bool
}

type command struct {
	*cobra.Command
	*Options
}

func genCommand() *command {
	c := &command{
		Command: &cobra.Command{
			Use:   "grpcrnd",
			Short: "A handy gRPC client which generate random parameter to send to gRPC method",
			RunE: func(cmd *cobra.Command, args []string) error {
				return cmd.Help()
			},
			SilenceErrors: true,
		},
		Options: &Options{},
	}
	c.Command.PersistentFlags().BoolVarP(&c.Options.Insecure, "insecure", "i", false, "with insecure")
	c.Command.PersistentFlags().BoolVar(&c.Options.StackTrace, "trace", false, "display detail error messages")
	c.Command.AddCommand(list.New(&c.Options.Insecure).Command())
	c.Command.AddCommand(call.New(&c.Options.Insecure).Command())
	return c
}

// New returns GRPCRand struct pointer
func New() *GRPCRand {
	return &GRPCRand{
		command: genCommand(),
	}
}

// Run method will create a project and returns exit code
func (g *GRPCRand) Run() int {
	if err := g.command.Execute(); err != nil {
		if g.command.StackTrace {
			fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", UnwrapErrors(err))
		}
		return 1
	}
	return 0
}
