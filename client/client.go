package client

import (
	"net"
)

type Config struct {
	Network string
	Address string
}

type Client struct {
	conf Config
}

func New(conf Config) Client {
	return Client{
		conf: conf,
	}
}

func (c Client) connect() (net.Conn, error) {
	return net.Dial(c.conf.Network, c.conf.Address)
}

func (c Client) Cmd(args ...string) Cmd {
	return Cmd{
		client: c,
		args:   args,
	}
}
