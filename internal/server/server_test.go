package server

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want string
	}{
		{
			name: "responds with 200 OK on fixed port",
			addr: ":18081",
			want: "HTTP/1.1 200 OK\r\n\r\n",
		},
		{
			name: "responds with 200 OK on different port",
			addr: ":18082",
			want: "HTTP/1.1 200 OK\r\n\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errChan := make(chan error, 1)
			go func() {
				errChan <- Start(tt.addr)
			}()

			// Give server time to start
			time.Sleep(10 * time.Millisecond)

			conn, err := net.Dial("tcp", tt.addr)
			if err != nil {
				t.Fatalf("failed to connect: %v", err)
			}
			defer conn.Close()

			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			if err != nil {
				t.Fatalf("failed to read: %v", err)
			}

			got := string(buffer[:n])
			if got != tt.want {
				t.Errorf("received %q, want %q", got, tt.want)
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
