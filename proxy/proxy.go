package proxy

import (
	"encoding/gob"
	"fmt"
	"io"
	"net"

	"github.com/ozbe/aws-vault-proxy/protocol"
)

type Config struct {
	Network string
	Address string
}

type Client struct {
	conf Config
}

func NewClient(conf Config) Client {
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

type Cmd struct {
	client   Client
	args     []string
	Stdout   io.Writer
	Stderr   io.Writer
	Stdin    io.Reader
	ExitCode *int
}

func (c *Cmd) Run() error {
	// create connection
	conn, err := c.client.connect()
	if err != nil {
		return err
	}

	// Send command
	cmd := protocol.Cmd{
		Args: c.args,
	}
	encoder := gob.NewEncoder(conn)
	encoder.Encode(cmd)

	// Stdin
	stdin := protocol.NewStdinWriter(conn)
	go io.Copy(stdin, c.Stdin)

	// Stdout, stderr, exit
	decoder := gob.NewDecoder(conn)
	var msg interface{}

	for {
		err = decoder.Decode(&msg)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch m := msg.(type) {
		case protocol.Stdout:
			write(c.Stdout, m.Data)
		case protocol.Stderr:
			write(c.Stderr, m.Data)
		case protocol.Exit:
			c.ExitCode = &m.ExitCode
			if m.Error != nil {
				return m.Error
			}
		default:
			return fmt.Errorf("unexpected msg: %v", m)
		}
	}

	return nil
}

func write(w io.Writer, p []byte) error {
	if w == nil {
		return nil
	}
	_, err := w.Write(p)
	return err
}
