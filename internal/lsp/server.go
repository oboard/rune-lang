package lsp

import (
	"bufio"
	"io"
)

func Serve(in io.Reader, out io.Writer) error {
	s := &server{
		reader: bufio.NewReader(in),
		out:    out,
		docs:   map[string]string{},
	}
	for {
		msg, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if err := s.handle(msg); err != nil {
			return err
		}
	}
}

type server struct {
	reader *bufio.Reader
	out    io.Writer
	docs   map[string]string
}
