package utils

import (
	"bytes"
	"fmt"
	"strconv"
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
	bodyLength int       // Expected length of the body, if any
}

// NewParser creates and returns a new Parser.
func NewParser() *Parser {
	return &Parser{
		state:   StateRequestLine,
		request: &Request{Headers: make(map[string]string)},
		bodyLength: -1, // -1 indicates that we haven't seen a Content-Length header yet
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
				if key == "Content-Length" {
					length, err := strconv.Atoi(value)
					if err != nil {
						return fmt.Errorf("invalid Content-Length value: %q", value)
					}
					p.bodyLength = length
				}
		case StateBody:
			// Scenario 1: We have a Content-Length.
			if p.bodyLength > 0 {
				if p.buffer.Len() >= p.bodyLength {
					p.request.Body = make([]byte, p.bodyLength)
					p.buffer.Read(p.request.Body)
					p.state = StateDone
				} else {
					// Not enough data yet, wait for the next chunk.
					return nil
				}
			// Scenario 2: Content-Length is explicitly zero.
			} else if p.bodyLength == 0 {
				p.state = StateDone
			// Scenario 3: No Content-Length header was found (bodyLength is -1).
			// This is a special case. We read the rest of the packet as the body.
			// In a real TCP stream, we would read until EOF.
			} else {
				p.request.Body = p.buffer.Bytes()
				p.buffer.Reset() // We've consumed the rest of the buffer.
				p.state = StateDone
			}

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

func (p *Parser) IngestString(data string) (*Request, error) {
	err := p.Parse([]byte(data))
	if err != nil {
		return nil, err
	}
	return p.request, nil
}

func (p *Parser) IngestBytes(data []byte) (*Request, error) {
	err := p.Parse(data)
	if err != nil {
		return nil, err
	}
	return p.request, nil
}
