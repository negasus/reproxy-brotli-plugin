package brotli

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/andybalholm/brotli"
)

var testBody = []byte("AAAAABBBBB")

type next struct{}

func (h *next) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("X-Test", "foo")
	rw.Write(testBody)
}

func TestCall(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	brWriter := brotli.NewWriter(buf)
	_, errWrite := brWriter.Write(testBody)
	if errWrite != nil {
		t.Fatalf("error write, %v", errWrite)
	}
	errFlush := brWriter.Flush()
	if errFlush != nil {
		t.Fatalf("error flush, %v", errFlush)
	}
	errClose := brWriter.Close()
	if errClose != nil {
		t.Fatalf("error close, %v", errClose)
	}

	n := &next{}

	h := Call(n)

	req := httptest.NewRequest(http.MethodGet, "https://example.com", http.NoBody)
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rw.Code)
	}

	if !bytes.Equal(rw.Body.Bytes(), buf.Bytes()) {
		t.Fatalf("wrong response")
	}

	if rw.Header().Get("Content-Encoding") != "br" {
		t.Fatalf("expected %v, got %v", "br", rw.Header().Get("Content-Encoding"))
	}
	if rw.Header().Get("Content-Length") != strconv.Itoa(buf.Len()) {
		t.Fatalf("expected %v, got %v", buf.Len(), rw.Header().Get("Content-Length"))
	}
	if rw.Header().Get("X-Test") != "foo" {
		t.Fatalf("expected %v, got %v", "foo", rw.Header().Get("X-Test"))
	}

}
