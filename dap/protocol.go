package dap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync/atomic"
)

var seq int64

func Message(msg any) ([]byte, error) {
	b, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("BuildMessage: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Content-Length: %d%s", len(b), NL))
	buf.WriteString(NL)
	buf.Write(b)
	return buf.Bytes(), nil
}

type Request struct {
	Seq       int64  `json:"seq"`
	Type      string `json:"type"` // ???: always "request"?
	Command   string `json:"command"`
	Arguments any    `json:"arguments,omitempty"`
}

func NewRequest(command string, arguments any) Request {
	return Request{
		Seq:       atomic.AddInt64(&seq, 1),
		Type:      "request",
		Command:   command,
		Arguments: arguments,
	}
}

func ReadMessage(r io.Reader, scratch []byte, buf *bytes.Buffer) (map[string]string, string, string, error) {
	headers, rawHeaders, err := ReadHeaders(r, scratch, buf)
	if err != nil {
		return nil, "", "", fmt.Errorf("ReadMessage: error reading headers: %w", err)
	}
	rawContentLength := headers["Content-Length"]
	contentLength, err := strconv.Atoi(rawContentLength)
	if err != nil {
		return nil, "", "", fmt.Errorf("ReadMessage: invalid Content-Length header: %s", rawContentLength)
	}
	body, err := ReadBody(r, buf, contentLength)
	if err != nil {
		return nil, "", "", fmt.Errorf("ReadMessage: error reading body: %w", err)
	}
	return headers, rawHeaders, body, nil
}

func ReadHeaders(r io.Reader, scratch []byte, buf *bytes.Buffer) (map[string]string, string, error) {
	const sep = NL + NL
	buf.Reset()
	for {
		n, err := r.Read(scratch)
		if err != nil {
			return nil, "", fmt.Errorf("ReadHeaders: error reading: %w", err)
		}
		if _, err = buf.Write(scratch[:n]); err != nil {
			return nil, "", fmt.Errorf("ReadHeaders: error writing to buffer: %w", err)
		}
		s := buf.String()
		if idx := strings.Index(s, sep); idx > -1 {
			rawHeaders := s[:idx]
			headers, err := ParseHeaders(rawHeaders)
			if err != nil {
				return nil, rawHeaders, fmt.Errorf("ReadHeaders: error parsing headers: %w", err)
			}
			rest := s[idx+len(sep):]
			buf.Reset()
			buf.WriteString(rest)
			return headers, rawHeaders, nil
		}
	}
}

func ParseHeaders(s string) (map[string]string, error) {
	headers := make(map[string]string)
	for _, line := range strings.Split(s, NL) {
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("unexpected header format: %s", line)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		headers[key] = value
	}
	return headers, nil
}

func ReadBody(r io.Reader, buf *bytes.Buffer, contentLength int) (string, error) {
	if buf.Len() < contentLength {
		if _, err := io.CopyN(buf, r, int64(contentLength-buf.Len())); err != nil {
			return "", fmt.Errorf("ReadBody: %w", err)
		}
	}
	s := buf.String()
	body := s[:contentLength]
	buf.Reset()
	buf.WriteString(s[contentLength:])
	return body, nil
}

type Response struct {
	Seq        int64           `json:"seq"`
	Type       string          `json:"type"`
	RequestSeq int64           `json:"request_seq"`
	Command    string          `json:"command"`
	Success    bool            `json:"success"`
	Body       json.RawMessage `json:"body"`
}

type Event struct {
	Seq   int64           `json:"seq"`
	Type  string          `json:"type"`
	Event string          `json:"event"`
	Body  json.RawMessage `json:"body"`
}
