package utils

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// a parser that can extract at least the HTTP version
// and Connection header from an HTTP Request.

// Define constants for our parser's State machine.
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
	State   int           // The current State (e.g., StateRequestLine)
	buffer  bytes.Buffer  // Accumulates incoming data chunks
	Request *Request      // The Request object we are building
	bodyLength int       // Expected length of the body, if any
}

// NewParser creates and returns a new Parser.
func NewParser() *Parser {
	return &Parser{
		State:   StateRequestLine,
		Request: &Request{Headers: make(map[string]string)},
		bodyLength: -1, // -1 indicates that we haven't seen a Content-Length header yet
	}
}

// Parse consumes a chunk of bytes and attempts to advance the parsing State.
// It returns an error if the input is malformed.
func (p *Parser) Parse(chunk []byte) error {
	p.buffer.Write(chunk)
	
	for (p.State != StateDone) && (p.buffer.Len() > 0) {
		switch p.State {
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
				return fmt.Errorf("malformed Request line")
			}
			p.Request.Method = parts[0]
			p.Request.Path = parts[1]
			p.Request.Version = strings.TrimSpace(parts[2])
			p.State = StateHeaders

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
					p.State = StateBody
					if p.Request.Method == "GET" || p.Request.Method == "HEAD" {
						// No body expected for GET or HEAD Requests
						p.State = StateDone
					}
					break
				}
				// Parse header line (e.g., "Host: example.com\r\n")
				colonIndex := strings.Index(line, ":")
				if colonIndex == -1 {
					return fmt.Errorf("malformed header line")
				}
				key := strings.TrimSpace(line[:colonIndex])
				value := strings.TrimSpace(line[colonIndex+1:])
				p.Request.Headers[key] = value
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
					p.Request.Body = make([]byte, p.bodyLength)
					p.buffer.Read(p.Request.Body)
					p.State = StateDone
				} else {
					// Not enough data yet, wait for the next chunk.
					return nil
				}
			// Scenario 2: Content-Length is explicitly zero.
			} else if p.bodyLength == 0 {
				p.State = StateDone
			// Scenario 3: No Content-Length header was found (bodyLength is -1).
			// This is a special case. We read the rest of the packet as the body.
			// In a real TCP stream, we would read until EOF.
			} else {
				p.Request.Body = p.buffer.Bytes()
				p.buffer.Reset() // We've consumed the rest of the buffer.
				p.State = StateDone
			}

			return nil
		case StateDone:
			// Parsing is complete.
			return nil
		default:
			return fmt.Errorf("unknown parser State")
		}
	}
	return nil


}

func (p *Parser) IngestString(data string) (*Request, error) {
	err := p.Parse([]byte(data))
	if err != nil {
		return nil, err
	}
	return p.Request, nil
}

func (p *Parser) IngestBytes(data []byte) (*Request, error) {
	err := p.Parse(data)
	if err != nil {
		return nil, err
	}
	return p.Request, nil
}
