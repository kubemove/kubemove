package server

import (
	"net"
	"os"
	"strconv"

	pb "github.com/kubemove/kubemove/pkg/plugin/framework/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	SERVER_PORT = "9000"
)

const (
	ENV_SERVER_ADDR = "SERVER"
	ENV_SERVER_PORT = "SERVER_PORT"
)

// TODO args fn
func NewServer(cb func(*grpc.ClientConn) interface{}) (Server, error) {
	addr, cert, err := getServerAddr()
	if err != nil {
		return nil, errors.New("Failed to fetch server details")
	}

	//TODO
	_ = cert
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := &server{
		lis: lis,
		cb:  cb,
	}
	s.gs = grpc.NewServer()
	pb.RegisterRegisterServer(s.gs, s)

	s.Plugin = make(map[string]plugin)
	return s, nil
}

func getServerAddr() (string, string, error) {
	caddr := os.Getenv(ENV_SERVER_ADDR)
	cport := os.Getenv(ENV_SERVER_PORT)

	if len(cport) == 0 {
		cport = SERVER_PORT
	}

	_, err := strconv.ParseUint(cport, 10, 16)
	if err != nil {
		return "", "", errors.Wrapf(err, "Unable to parse port")
	}

	return caddr + ":" + cport, "", nil
}

func (s *server) Serve() {
	s.active = true
	if err := s.gs.Serve(s.lis); err != nil {
		panic(err)
	}
}
