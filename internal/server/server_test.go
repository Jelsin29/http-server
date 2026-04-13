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
	"sync"
	"testing"
	"time"

	"github.com/jelsin/http-server/internal/request"
	"github.com/jelsin/http-server/internal/response"
	"github.com/jelsin/http-server/internal/static"
)

func TestStart(t *testing.T) {
	tests := []struct {
		name       string
		addr       string
		request    string
		wantStatus string
		wantBody   string
	}{
		{
			name:       "serves index file from root path",
			addr:       ":18081",
			request:    "GET / HTTP/1.1\r\nHost: localhost:18081\r\n\r\n",
			wantStatus: "HTTP/1.1 200 OK",
			wantBody:   mustReadFile(t, "../../public/index.html"),
		},
		{
			name:       "serves css asset with body",
			addr:       ":18082",
			request:    "GET /assets/app.css HTTP/1.1\r\nHost: localhost:18082\r\nUser-Agent: test\r\n\r\n",
			wantStatus: "HTTP/1.1 200 OK",
			wantBody:   mustReadFile(t, "../../public/assets/app.css"),
		},
		{
			name:       "returns bad request for malformed request",
			addr:       ":18083",
			request:    "GET / HTTP/1.1\r\nHost localhost\r\n\r\n",
			wantStatus: "HTTP/1.1 400 Bad Request",
			wantBody:   "bad request",
		},
		{
			name:       "returns not found for missing file",
			addr:       ":18084",
			request:    "GET /missing.txt HTTP/1.1\r\nHost: localhost:18084\r\n\r\n",
			wantStatus: "HTTP/1.1 404 Not Found",
			wantBody:   "not found",
		},
		{
			name:       "returns not found for traversal attempt",
			addr:       ":18085",
			request:    "GET /../secret.txt HTTP/1.1\r\nHost: localhost:18085\r\n\r\n",
			wantStatus: "HTTP/1.1 404 Not Found",
			wantBody:   "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startServer(t, tt.addr)

			responseText := exchangeRequest(t, tt.addr, tt.request)

			if !strings.Contains(responseText, tt.wantStatus) {
				t.Fatalf("response = %q, want status %q", responseText, tt.wantStatus)
			}

			if !strings.Contains(responseText, tt.wantBody) {
				t.Fatalf("response = %q, want body containing %q", responseText, tt.wantBody)
			}

			contentLength, found := responseHeader(responseText, "Content-Length")
			if !found {
				t.Fatalf("response = %q, want Content-Length header", responseText)
			}

			if contentLength != strconv.Itoa(bodyLength(responseText)) {
				t.Fatalf("Content-Length = %q, want %d", contentLength, bodyLength(responseText))
			}

			if tt.name == "serves css asset with body" && !strings.Contains(responseText, "Content-Type: text/css; charset=utf-8") {
				t.Fatalf("response = %q, want css content type", responseText)
			}

			if tt.name == "serves index file from root path" && !strings.Contains(responseText, "Content-Type: text/html; charset=utf-8") {
				t.Fatalf("response = %q, want html content type", responseText)
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
		t.Fatal("expected error when port in use, got nil")
	}

	if !strings.Contains(err.Error(), "address already in use") && !strings.Contains(err.Error(), "bind") {
		t.Fatalf("expected error containing 'address already in use' or 'bind', got %v", err)
	}
}

func TestServeContinuesAfterAcceptError(t *testing.T) {
	conn := &stubConn{readBuffer: bytes.NewBuffer(nil), remoteAddr: stubAddr("client:55")}
	listener := &scriptedListener{
		steps: []acceptStep{
			{err: errors.New("temporary accept boom")},
			{conn: conn},
			{err: net.ErrClosed},
		},
	}

	handled := make(chan net.Conn, 1)
	stubConnectionHandler(t, func(conn net.Conn) {
		handled <- conn
	})

	logs := captureLogs(t)

	if err := serve(listener); err != nil {
		t.Fatalf("serve() error = %v, want nil", err)
	}

	select {
	case got := <-handled:
		if got != conn {
			t.Fatalf("handled conn = %v, want scripted conn %v", got, conn)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("expected connection handler to run after accept error")
	}

	if !strings.Contains(logs.String(), "accept error: temporary accept boom") {
		t.Fatalf("logs = %q, want accept error entry", logs.String())
	}
}

func TestStartReturnsInternalServerError(t *testing.T) {
	stubAssetLoader(t, func(root string, target string) (*static.Asset, error) {
		if target == "/boom" {
			return nil, errors.New("disk melted")
		}

		return static.Load(root, target)
	})

	startServer(t, ":18086")

	responseText := exchangeRequest(t, ":18086", "GET /boom HTTP/1.1\r\nHost: localhost:18086\r\n\r\n")

	if !strings.Contains(responseText, "HTTP/1.1 500 Internal Server Error") {
		t.Fatalf("response = %q, want 500 status", responseText)
	}

	if !strings.Contains(responseText, "internal server error") {
		t.Fatalf("response = %q, want generic 500 body", responseText)
	}

	if strings.Contains(responseText, "disk melted") {
		t.Fatalf("response = %q, should not leak internal error details", responseText)
	}
}

func TestStartHandlesRequestsConcurrently(t *testing.T) {
	var blockOnce sync.Once
	blocked := make(chan struct{})
	release := make(chan struct{})

	stubAssetLoader(t, func(root string, target string) (*static.Asset, error) {
		if target == "/block" {
			blockOnce.Do(func() {
				close(blocked)
			})
			<-release
			return static.Load(root, "/")
		}

		return static.Load(root, target)
	})

	startServer(t, ":18087")

	firstDone := make(chan string, 1)
	firstErr := make(chan error, 1)

	go func() {
		responseText, err := exchangeRequestWithError(":18087", "GET /block HTTP/1.1\r\nHost: localhost:18087\r\n\r\n")
		if err != nil {
			firstErr <- err
			return
		}

		firstDone <- responseText
	}()

	select {
	case <-blocked:
	case err := <-firstErr:
		t.Fatalf("first request failed early: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for blocked request")
	}

	secondResponse := exchangeRequest(t, ":18087", "GET /assets/app.css HTTP/1.1\r\nHost: localhost:18087\r\n\r\n")
	if !strings.Contains(secondResponse, "HTTP/1.1 200 OK") {
		t.Fatalf("second response = %q, want 200 status", secondResponse)
	}

	if !strings.Contains(secondResponse, mustReadFile(t, "../../public/assets/app.css")) {
		t.Fatalf("second response = %q, want css body", secondResponse)
	}

	select {
	case responseText := <-firstDone:
		t.Fatalf("first response finished before release: %q", responseText)
	case err := <-firstErr:
		t.Fatalf("first request failed before release: %v", err)
	default:
	}

	close(release)

	select {
	case responseText := <-firstDone:
		if !strings.Contains(responseText, "HTTP/1.1 200 OK") {
			t.Fatalf("first response = %q, want 200 status", responseText)
		}
		if !strings.Contains(responseText, mustReadFile(t, "../../public/index.html")) {
			t.Fatalf("first response = %q, want index body", responseText)
		}
	case err := <-firstErr:
		t.Fatalf("first request failed: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for first response after release")
	}
}

func TestHandleConnection(t *testing.T) {
	tests := []struct {
		name             string
		request          string
		writeErr         error
		setup            func(t *testing.T)
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
			name:             "writes bad request response for malformed request",
			request:          "GET / HTTP/1.1\r\nHost localhost\r\n\r\n",
			wantResponse:     string(response.PlainText(400, "Bad Request", "bad request")),
			wantClosed:       true,
			wantWriteCalls:   1,
			wantLogSubstring: "parsing request",
		},
		{
			name: "writes internal error response when rendering fails",
			setup: func(t *testing.T) {
				stubRenderer(t, func(msg response.Message) ([]byte, error) {
					return nil, errors.New("render blew up")
				})
			},
			request:          "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n",
			wantResponse:     string(response.PlainText(500, "Internal Server Error", "internal server error")),
			wantClosed:       true,
			wantWriteCalls:   1,
			wantLogSubstring: "building response",
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
			if tt.setup != nil {
				tt.setup(t)
			}

			conn := &stubConn{
				readBuffer: bytes.NewBufferString(tt.request),
				writeErr:   tt.writeErr,
				remoteAddr: stubAddr("client:1234"),
			}

			logs := captureLogs(t)

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

func startServer(t *testing.T, addr string) {
	t.Helper()

	go func() {
		_ = Start(addr)
	}()

	waitForServer(t, addr)
}

func waitForServer(t *testing.T, addr string) {
	t.Helper()

	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 20*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("server on %s did not start in time", addr)
}

func exchangeRequest(t *testing.T, addr string, rawRequest string) string {
	t.Helper()

	responseText, err := exchangeRequestWithError(addr, rawRequest)
	if err != nil {
		t.Fatalf("exchange request %q on %s: %v", rawRequest, addr, err)
	}

	return responseText
}

func exchangeRequestWithError(addr string, rawRequest string) (string, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(1 * time.Second)); err != nil {
		return "", err
	}

	if _, err := conn.Write([]byte(rawRequest)); err != nil {
		return "", err
	}

	body, err := io.ReadAll(conn)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	var logs bytes.Buffer
	previousWriter := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() {
		log.SetOutput(previousWriter)
	})

	return &logs
}

func stubAssetLoader(t *testing.T, loader func(string, string) (*static.Asset, error)) {
	t.Helper()

	previous := loadAsset
	loadAsset = loader
	t.Cleanup(func() {
		loadAsset = previous
	})
}

func stubRenderer(t *testing.T, renderer func(response.Message) ([]byte, error)) {
	t.Helper()

	previous := renderMessage
	renderMessage = renderer
	t.Cleanup(func() {
		renderMessage = previous
	})
}

func stubConnectionHandler(t *testing.T, handler func(net.Conn)) {
	t.Helper()

	previous := handleAcceptedConnection
	handleAcceptedConnection = handler
	t.Cleanup(func() {
		handleAcceptedConnection = previous
	})
}

type acceptStep struct {
	conn net.Conn
	err  error
}

type scriptedListener struct {
	steps []acceptStep
	index int
	mu    sync.Mutex
}

func (l *scriptedListener) Accept() (net.Conn, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.index >= len(l.steps) {
		return nil, net.ErrClosed
	}

	step := l.steps[l.index]
	l.index++
	return step.conn, step.err
}

func (l *scriptedListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.index = len(l.steps)
	return nil
}

func (l *scriptedListener) Addr() net.Addr {
	return stubAddr("listener:8080")
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
