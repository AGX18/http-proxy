package utils

import (
	"bytes"
	"fmt"
	"strings"
)

// a parser that can extract at least the HTTP version
// and Connection header from an HTTP request.

// Define constants for our parser's state machine.
const (
	StateRequestLine = iota
	StateHeaders
	StateBody
	StateDone
)

type Request struct {
	Method  string
	Path    string
	Version string
	Headers map[string]string
	Body    []byte
}

type Parser struct {
	state   int           // The current state (e.g., StateRequestLine)
	buffer  bytes.Buffer  // Accumulates incoming data chunks
	request *Request      // The Request object we are building
}

// NewParser creates and returns a new Parser.
func NewParser() *Parser {
	return &Parser{
		state:   StateRequestLine,
		request: &Request{Headers: make(map[string]string)},
	}
}

// Parse consumes a chunk of bytes and attempts to advance the parsing state.
// It returns an error if the input is malformed.
func (p *Parser) Parse(chunk []byte) error {
	p.buffer.Write(chunk)
	
	for (p.state != StateDone) && (p.buffer.Len() > 0) {
		switch p.state {
		case StateRequestLine:
			line, err := p.buffer.ReadString('\n')
			if err != nil {
				// Not enough data yet, wait for the next chunk.
				p.buffer.WriteString(line) // Put the partial line back
				return nil
			}

			// Parse the line (e.g., "GET / HTTP/1.1\r\n")
			parts := strings.Fields(line)
			if len(parts) < 3 {
				return fmt.Errorf("malformed request line")
			}
			p.request.Method = parts[0]
			p.request.Path = parts[1]
			p.request.Version = strings.TrimSpace(parts[2])
			p.state = StateHeaders

		case StateHeaders:
			for {
				line, err := p.buffer.ReadString('\n')
				if err != nil {
					// Not enough data yet, wait for the next chunk.
					p.buffer.WriteString(line) // Put the partial line back
					return nil
				}
				line = strings.TrimSpace(line)
				if line == "" || line == "\r" {
					// End of headers
					p.state = StateBody
					break
				}
				// Parse header line (e.g., "Host: example.com\r\n")
				colonIndex := strings.Index(line, ":")
				if colonIndex == -1 {
					return fmt.Errorf("malformed header line")
				}
				key := strings.TrimSpace(line[:colonIndex])
				value := strings.TrimSpace(line[colonIndex+1:])
				p.request.Headers[key] = value
			}
		case StateBody:
			// For simplicity, we assume no body for now.
			p.state = StateDone
			return nil
		case StateDone:
			// Parsing is complete.
			return nil
		default:
			return fmt.Errorf("unknown parser state")
		}
	}
	return nil


}

func IngestString(data string) (*Request, error) {
	parser := NewParser()
	err := parser.Parse([]byte(data))
	if err != nil {
		return nil, err
	}
	return parser.request, nil
}

func IngestBytes(data []byte) (*Request, error) {
	parser := NewParser()
	err := parser.Parse(data)
	if err != nil {
		return nil, err
	}
	return parser.request, nil
}
