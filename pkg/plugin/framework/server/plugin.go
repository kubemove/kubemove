package server

import (
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func newPlugin(name, addr string) (*plugin, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to make connection")
	}

	return &plugin{
		name: name,
		addr: addr,
		conn: conn,
	}, nil
}
