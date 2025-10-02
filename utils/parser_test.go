package utils

import (
	"reflect"
	"testing"
)

// TestParser is the test function for the Parser.
func TestParser(t *testing.T) {
    got, err := IngestString("GET / HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n")
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