package protocol

import (
	"encoding/gob"
	"io"
)

func init() {
	gob.Register(Cmd{})
	gob.Register(Stdin{})
	gob.Register(Stdout{})
	gob.Register(Stderr{})
	gob.Register(Exit{})
}

type Cmd struct {
	Args []string
}

type Stdin struct {
	Data []byte
}

type Stdout struct {
	Data []byte
}

type Stderr struct {
	Data []byte
}

type Exit struct {
	ExitCode int
	Error    error
}

type Writer struct {
	encoder *gob.Encoder
	wrap    func([]byte) interface{}
}

func NewStdinWriter(w io.Writer) Writer {
	return newWriter(w, func(p []byte) interface{} {
		return Stdin{p}
	})
}

func NewStdoutWriter(w io.Writer) Writer {
	return newWriter(w, func(p []byte) interface{} {
		return Stdout{p}
	})
}

func NewStderrWriter(w io.Writer) Writer {
	return newWriter(w, func(p []byte) interface{} {
		return Stderr{p}
	})
}

func newWriter(w io.Writer, wrap func([]byte) interface{}) Writer {
	return Writer{
		encoder: gob.NewEncoder(w),
		wrap:    wrap,
	}
}

func (w Writer) Write(p []byte) (int, error) {
	msg := w.wrap(p)
	err := w.encoder.Encode(&msg)

	if err != nil {
		return 0, err
	}

	return len(p), err
}
