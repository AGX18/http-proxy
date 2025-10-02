package utils

import (
	"reflect"
	"testing"
)

// TestParser is the test function for the Parser.
func TestParser(t *testing.T) {
	p := NewParser()
    got, err := p.IngestString("GET / HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n")
    if err != nil {
        t.Fatalf("IngestString() error = %v", err)
    }
    want := &Request{
        Method:    "GET",
        Path:     "/",
        Version:  "HTTP/1.1",
        Headers:  map[string]string{"Host": "example.com", "Connection": "close"},
    }

    if !reflect.DeepEqual(got, want) {
        t.Errorf("Parse() = %+v; want %+v", got, want)
    }
}

func TestParserFragmented(t *testing.T) {
	testCases := []struct {
        name    string // A name for the test case
        str     []string
        want    *Request
    }{
        {"two parts request", []string{"GET / HTTP/1.1\r\n", "Host: example.com\r\nConnection: close\r\n\r\n"}, &Request{Method: "GET", Path: "/", Version: "HTTP/1.1", Headers: map[string]string{"Host": "example.com", "Connection": "close"}}},
        {"three parts request", []string{"GET / HTTP/1.1\r\n", "Host: example.com\r", "\nConnection: close\r\n\r\n"}, &Request{Method: "GET", Path: "/", Version: "HTTP/1.1", Headers: map[string]string{"Host": "example.com", "Connection": "close"}}},
    }

    // Iterate over the test cases
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
			var got *Request
			var err error
			p := NewParser()
			// Feed each fragment to the parser
			for _, fragment := range tc.str {
				got, err = p.IngestString(fragment)
				if err != nil && tc.want != nil {
					t.Fatalf("IngestString() error = %v", err)
				}
			}
            if !reflect.DeepEqual(got, tc.want) {
                t.Errorf("IngestString() = %+v; want %+v", got, tc.want)
            }
        })
    }
}

     