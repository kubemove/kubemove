package client

import (
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type Initializer func(*grpc.Server)

var mux sync.Mutex
var conn *grpc.ClientConn

// TODO args fn to get user input
func NewClient(name string, opt Initializer) error {
	var wg sync.WaitGroup
	sAddrList, err := getServerAddr()
	if err != nil {
		return errors.Wrapf(err, "Insufficient server details")
	}

	cAddr, err := getClientAddr()
	if err != nil {
		return errors.Wrapf(err, "Insufficient client details")
	}

	mux.Lock()
	defer mux.Unlock()
	wg.Add(1)
	go func() {
		defer wg.Done()
		createServer(cAddr, opt)
	}()

	if err != nil {
		return errors.Wrapf(err, "Failed to create server")
	}

	for _, sAddr := range sAddrList {
		conn, err = register(name, cAddr, sAddr)
		if err != nil {
			return errors.Wrapf(err, "Failed to register plugin")
		}
	}
	wg.Wait()
	return nil
}

func createServer(cAddr *addr, opt Initializer) {
	defer shutdown()

	lis, err := net.Listen("tcp", cAddr.addr)
	if err != nil {
		panic(err)
		return
	}

	s := grpc.NewServer()
	opt(s)

	if err := s.Serve(lis); err != nil {
		panic(err)
		return
	}
	return
}

func shutdown() {
	mux.Lock()
	fmt.Printf("shutting down the process\n")
	mux.Unlock()
	if conn == nil {
		fmt.Printf("This is an error.. connection is nil!!\n")
	} else {
		conn.Close()
	}
	os.Exit(1)
}
