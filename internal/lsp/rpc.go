package lsp

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func (s *server) readMessage() (request, error) {
	contentLength := -1
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return request{}, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(parts[0]), "Content-Length") {
			n, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return request{}, err
			}
			contentLength = n
		}
	}
	if contentLength < 0 {
		return request{}, fmt.Errorf("missing Content-Length header")
	}
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(s.reader, body); err != nil {
		return request{}, err
	}
	var req request
	if err := json.Unmarshal(body, &req); err != nil {
		return request{}, err
	}
	return req, nil
}

func (s *server) respond(id json.RawMessage, result any) error {
	if len(id) == 0 {
		return nil
	}
	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      json.RawMessage(id),
		"result":  result,
	}
	return s.write(payload)
}

func (s *server) notify(method string, params any) error {
	return s.write(map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	})
}

func (s *server) write(payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(s.out, "Content-Length: %d\r\n\r\n%s", len(data), data)
	return err
}
