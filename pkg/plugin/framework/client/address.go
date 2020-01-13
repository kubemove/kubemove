package client

import (
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	DEFAULT_CLIENT_PORT = "8000"
)

type addr struct {
	addr string
	cert string
}

const (
	ENV_SERVER_ADDR = "SERVER"
	ENV_SERVER_PORT = "SERVER_PORT"
	ENV_CLIENT_ADDR = "CLIENT"
	ENV_CLIENT_PORT = "CLIENT_PORT"
	ENV_SERVER_CERT = "CERT"
)

func getClientAddr() (*addr, error) {
	var err error

	caddr := os.Getenv(ENV_CLIENT_ADDR)
	cport := os.Getenv(ENV_CLIENT_PORT)

	if len(cport) == 0 {
		cport = DEFAULT_CLIENT_PORT
	}

	_, err = strconv.ParseUint(cport, 10, 16)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to parse port")
	}

	return &addr{
		addr: caddr + ":" + cport,
	}, nil
}

func getServerAddr() ([]*addr, error) {
	var saddr []*addr
	caddr := os.Getenv(ENV_SERVER_ADDR)
	cport := os.Getenv(ENV_SERVER_PORT)

	if len(cport) == 0 {
		cport = "9000"
	}

	cert := os.Getenv(ENV_SERVER_CERT)

	aip := strings.Split(caddr, ",")
	aport := strings.Split(cport, ",")

	if len(aip) != len(aport) {
		return nil, errors.Errorf("Invalid server address/port")
	}

	for _, k := range aport {
		_, err := strconv.ParseUint(k, 10, 16)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to parse port")
		}
	}

	for k := range aip {
		laddr := &addr{
			addr: aip[k] + ":" + aport[k],
			cert: cert,
		}
		saddr = append(saddr, laddr)
	}

	return saddr, nil
}
