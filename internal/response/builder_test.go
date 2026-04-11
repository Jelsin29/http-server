package response

import (
	"strings"
	"testing"
)

func TestBuild(t *testing.T) {
	tests := []struct {
		name string
		msg  Message
		want string
	}{
		{
			name: "builds response with headers and body",
			msg: Message{
				StatusCode: 200,
				Reason:     "OK",
				Headers: map[string]string{
					"Content-Length": "5",
					"Content-Type":   "text/plain; charset=utf-8",
				},
				Body: []byte("hello"),
			},
			want: strings.Join([]string{
				"HTTP/1.1 200 OK",
				"Content-Length: 5",
				"Content-Type: text/plain; charset=utf-8",
				"",
				"hello",
			}, "\r\n"),
		},
		{
			name: "builds headerless response",
			msg: Message{
				StatusCode: 404,
				Reason:     "Not Found",
			},
			want: "HTTP/1.1 404 Not Found\r\n\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(Build(tt.msg))
			if got != tt.want {
				t.Fatalf("Build() = %q, want %q", got, tt.want)
			}
		})
	}
}
