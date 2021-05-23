package proxy

import (
	"encoding/gob"
	"fmt"
	"io"
)

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

	// Send args
	args := Args(c.args)
	encoder := gob.NewEncoder(conn)
	encoder.Encode(args)

	// Stdin
	stdin := NewStdinWriter(conn)
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
		case Stdout:
			write(c.Stdout, m)
		case Stderr:
			write(c.Stderr, m)
		case Exit:
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
