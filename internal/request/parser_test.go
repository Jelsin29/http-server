package request

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        *Request
		wantErrText string
	}{
		{
			name: "parses request line and headers",
			input: "GET /hello HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"User-Agent: curl/8.0\r\n" +
				"X-Spaces:   keep me trimmed   \r\n" +
				"\r\n",
			want: &Request{
				Method:  "GET",
				Target:  "/hello",
				Version: "HTTP/1.1",
				Headers: map[string]string{
					"Host":       "localhost:8080",
					"User-Agent": "curl/8.0",
					"X-Spaces":   "keep me trimmed",
				},
			},
		},
		{
			name: "rejects malformed request line",
			input: "GET /only-two-parts\r\n" +
				"Host: localhost\r\n" +
				"\r\n",
			wantErrText: "expected method target version",
		},
		{
			name: "rejects malformed header",
			input: "GET / HTTP/1.1\r\n" +
				"Host localhost\r\n" +
				"\r\n",
			wantErrText: "missing colon",
		},
		{
			name: "rejects header with whitespace before colon",
			input: "GET / HTTP/1.1\r\n" +
				"Host : localhost\r\n" +
				"\r\n",
			wantErrText: "whitespace before colon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(strings.NewReader(tt.input))
			if tt.wantErrText != "" {
				if err == nil {
					t.Fatalf("Parse() error = nil, want error containing %q", tt.wantErrText)
				}

				if !strings.Contains(err.Error(), tt.wantErrText) {
					t.Fatalf("Parse() error = %v, want text %q", err, tt.wantErrText)
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if got.Method != tt.want.Method {
				t.Errorf("Method = %q, want %q", got.Method, tt.want.Method)
			}

			if got.Target != tt.want.Target {
				t.Errorf("Target = %q, want %q", got.Target, tt.want.Target)
			}

			if got.Version != tt.want.Version {
				t.Errorf("Version = %q, want %q", got.Version, tt.want.Version)
			}

			if len(got.Headers) != len(tt.want.Headers) {
				t.Fatalf("len(Headers) = %d, want %d", len(got.Headers), len(tt.want.Headers))
			}

			for name, wantValue := range tt.want.Headers {
				if got.Headers[name] != wantValue {
					t.Errorf("Headers[%q] = %q, want %q", name, got.Headers[name], wantValue)
				}
			}
		})
	}
}
