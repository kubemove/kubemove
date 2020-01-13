package client

import (
	"context"
	"time"

	pb "github.com/kubemove/kubemove/pkg/plugin/framework/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func register(name string, cAddr, sAddr *addr) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(sAddr.addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to make connection")
	}

	cl := pb.NewRegisterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second /*TODO */)
	defer cancel()

	_, err = cl.RegisterPlugin(ctx,
		&pb.Request{
			Name:    name,
			Address: cAddr.addr,
		})

	if err != nil {
		conn.Close()
		return nil, errors.Wrapf(err, "Error registering plugin")
	}

	return conn, nil
}
