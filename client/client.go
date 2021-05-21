package client

import (
	"net"
)

type Client struct {
	Network string
	Address string
}

func New(network string, address string) Client {
	return Client{
		Network: network,
		Address: address,
	}
}

func (c Client) connect() (net.Conn, error) {
	return net.Dial(c.Network, c.Address)
}

func (c Client) Cmd(args ...string) Cmd {
	return Cmd{
		client: c,
		args:   args,
	}
}
