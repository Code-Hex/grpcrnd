package grpcrnd

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func NewGRPCConnection(ctx context.Context, addr string, insecure bool) (*grpc.ClientConn, error) {
	var dialOpts []grpc.DialOption
	if insecure {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	} else {
		certFile := ""
		creds, err := credentials.NewClientTLSFromFile(certFile, "")
		if err != nil {
			return nil, errors.Wrap(err, "failed to load credential files")
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	}
	return grpc.DialContext(ctx, addr, dialOpts...)
}
