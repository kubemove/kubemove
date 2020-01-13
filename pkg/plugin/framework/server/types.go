package server

import (
	"net"
	"sync"

	"google.golang.org/grpc"
)

// TODO refactor
type Server interface {
	Serve()
	List() []string
	Get(string, bool) interface{}
	Put(string) bool
}

type plugin struct {
	//TODO lock at plugin level
	name string
	addr string
	conn *grpc.ClientConn
	path interface{}
	busy bool
}

type server struct {
	sync.RWMutex
	active bool
	Plugin map[string]plugin
	gs     *grpc.Server
	lis    net.Listener
	cb     func(*grpc.ClientConn) interface{}
}
