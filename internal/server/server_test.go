package server

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jelsin/http-server/internal/request"
)

func TestStart(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		request string
		want    string
	}{
		{
			name:    "serves index file from root path",
			addr:    ":18081",
			request: "GET / HTTP/1.1\r\nHost: localhost:18081\r\n\r\n",
			want:    mustReadFile(t, "../../public/index.html"),
		},
		{
			name:    "serves css asset with body",
			addr:    ":18082",
			request: "GET /assets/app.css HTTP/1.1\r\nHost: localhost:18082\r\nUser-Agent: test\r\n\r\n",
			want:    mustReadFile(t, "../../public/assets/app.css"),
		},
		{
			name:    "returns not found for missing file",
			addr:    ":18084",
			request: "GET /missing.txt HTTP/1.1\r\nHost: localhost:18084\r\n\r\n",
			want:    "not found",
		},
		{
			name:    "returns not found for traversal attempt",
			addr:    ":18085",
			request: "GET /../secret.txt HTTP/1.1\r\nHost: localhost:18085\r\n\r\n",
			want:    "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func() {
				_ = Start(tt.addr)
			}()

			// Give server time to start.
			time.Sleep(10 * time.Millisecond)

			conn, err := net.Dial("tcp", tt.addr)
			if err != nil {
				t.Fatalf("failed to connect: %v", err)
			}
			defer conn.Close()

			if _, err := conn.Write([]byte(tt.request)); err != nil {
				t.Fatalf("failed to write request: %v", err)
			}

			got, err := io.ReadAll(conn)
			if err != nil {
				t.Fatalf("failed to read: %v", err)
			}

			responseText := string(got)

			if tt.name == "returns not found for missing file" || tt.name == "returns not found for traversal attempt" {
				if !strings.Contains(responseText, "HTTP/1.1 404 Not Found") {
					t.Fatalf("response = %q, want 404 status", responseText)
				}
			} else if !strings.Contains(responseText, "HTTP/1.1 200 OK") {
				t.Fatalf("response = %q, want 200 status", responseText)
			}

			if !strings.Contains(responseText, tt.want) {
				t.Errorf("response = %q, want body containing %q", responseText, tt.want)
			}

			contentLength, found := responseHeader(responseText, "Content-Length")
			if !found {
				t.Errorf("response = %q, want Content-Length header", responseText)
			} else if contentLength != strconv.Itoa(bodyLength(responseText)) {
				t.Errorf("Content-Length = %q, want %d", contentLength, bodyLength(responseText))
			}

			if tt.name == "serves css asset with body" && !strings.Contains(responseText, "Content-Type: text/css; charset=utf-8") {
				t.Errorf("response = %q, want css content type", responseText)
			}

			if tt.name == "serves index file from root path" && !strings.Contains(responseText, "Content-Type: text/html; charset=utf-8") {
				t.Errorf("response = %q, want html content type", responseText)
			}
		})
	}
}

func TestStartPortInUse(t *testing.T) {
	ln, err := net.Listen("tcp", ":18080")
	if err != nil {
		t.Fatalf("failed to bind test port: %v", err)
	}
	defer ln.Close()

	err = Start(":18080")
	if err == nil {
		t.Error("expected error when port in use, got nil")
	}

	if !strings.Contains(err.Error(), "address already in use") &&
		!strings.Contains(err.Error(), "bind") {
		t.Errorf("expected error containing 'address already in use' or 'bind', got %v", err)
	}
}

func TestStartMalformedRequest(t *testing.T) {
	addr := ":18083"

	go func() {
		_ = Start(addr)
	}()

	// Give server time to start.
	time.Sleep(10 * time.Millisecond)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	request := "GET / HTTP/1.1\r\nHost localhost\r\n\r\n"
	if _, err := conn.Write([]byte(request)); err != nil {
		t.Fatalf("failed to write malformed request: %v", err)
	}

	got, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("received %q, want empty response for malformed request", string(got))
	}
}

func TestHandleConnection(t *testing.T) {
	tests := []struct {
		name             string
		request          string
		writeErr         error
		wantResponse     string
		wantClosed       bool
		wantWriteCalls   int
		wantLogSubstring string
	}{
		{
			name:           "writes response and closes on valid request",
			request:        "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n",
			wantResponse:   string(mustBuildResponse(t, &request.Request{Method: "GET", Target: "/", Version: "HTTP/1.1"})),
			wantClosed:     true,
			wantWriteCalls: 1,
		},
		{
			name:           "writes not found response for missing file",
			request:        "GET /missing.txt HTTP/1.1\r\nHost: localhost\r\n\r\n",
			wantResponse:   string(mustBuildResponse(t, &request.Request{Method: "GET", Target: "/missing.txt", Version: "HTTP/1.1"})),
			wantClosed:     true,
			wantWriteCalls: 1,
		},
		{
			name:             "logs parse error and closes malformed request",
			request:          "GET / HTTP/1.1\r\nHost localhost\r\n\r\n",
			wantClosed:       true,
			wantWriteCalls:   0,
			wantLogSubstring: "parsing request",
		},
		{
			name:             "logs write error and still closes connection",
			request:          "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n",
			writeErr:         errors.New("boom"),
			wantClosed:       true,
			wantWriteCalls:   1,
			wantLogSubstring: "writing response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &stubConn{
				readBuffer: bytes.NewBufferString(tt.request),
				writeErr:   tt.writeErr,
				remoteAddr: stubAddr("client:1234"),
			}

			var logs bytes.Buffer
			previousWriter := log.Writer()
			log.SetOutput(&logs)
			defer log.SetOutput(previousWriter)

			handleConnection(conn)

			if conn.closed != tt.wantClosed {
				t.Fatalf("closed = %v, want %v", conn.closed, tt.wantClosed)
			}

			if conn.writeCalls != tt.wantWriteCalls {
				t.Fatalf("writeCalls = %d, want %d", conn.writeCalls, tt.wantWriteCalls)
			}

			if conn.writeBuffer.String() != tt.wantResponse {
				t.Fatalf("response = %q, want %q", conn.writeBuffer.String(), tt.wantResponse)
			}

			if tt.wantLogSubstring == "" {
				if logs.Len() != 0 {
					t.Fatalf("unexpected logs: %q", logs.String())
				}
				return
			}

			if !strings.Contains(logs.String(), tt.wantLogSubstring) {
				t.Fatalf("logs = %q, want substring %q", logs.String(), tt.wantLogSubstring)
			}
		})
	}
}

func mustBuildResponse(t *testing.T, req *request.Request) []byte {
	t.Helper()

	responseBytes, err := buildResponse(req)
	if err != nil {
		t.Fatalf("buildResponse() error = %v", err)
	}

	return responseBytes
}

func mustReadFile(t *testing.T, name string) string {
	t.Helper()

	body, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("reading file %s: %v", name, err)
	}

	return string(body)
}

func responseHeader(responseText string, name string) (string, bool) {
	lines := strings.Split(responseText, "\r\n")
	prefix := name + ": "

	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimPrefix(line, prefix), true
		}
	}

	return "", false
}

func bodyLength(responseText string) int {
	parts := strings.SplitN(responseText, "\r\n\r\n", 2)
	if len(parts) != 2 {
		return 0
	}

	return len(parts[1])
}

type stubConn struct {
	readBuffer  *bytes.Buffer
	writeBuffer bytes.Buffer
	writeErr    error
	closed      bool
	writeCalls  int
	remoteAddr  net.Addr
}

func (c *stubConn) Read(p []byte) (int, error) {
	return c.readBuffer.Read(p)
}

func (c *stubConn) Write(p []byte) (int, error) {
	c.writeCalls++
	if c.writeErr != nil {
		return 0, c.writeErr
	}

	return c.writeBuffer.Write(p)
}

func (c *stubConn) Close() error {
	c.closed = true
	return nil
}

func (c *stubConn) LocalAddr() net.Addr {
	return stubAddr("server:8080")
}

func (c *stubConn) RemoteAddr() net.Addr {
	if c.remoteAddr != nil {
		return c.remoteAddr
	}

	return stubAddr("client:0")
}

func (c *stubConn) SetDeadline(time.Time) error {
	return nil
}

func (c *stubConn) SetReadDeadline(time.Time) error {
	return nil
}

func (c *stubConn) SetWriteDeadline(time.Time) error {
	return nil
}

type stubAddr string

func (a stubAddr) Network() string {
	return "tcp"
}

func (a stubAddr) String() string {
	return string(a)
}
