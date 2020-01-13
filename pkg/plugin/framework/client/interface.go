package client

type Client interface {
	RegisterPlugin(name string) interface{}
}
