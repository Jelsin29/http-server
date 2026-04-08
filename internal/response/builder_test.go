package response

import "testing"

func TestHardcodedOK(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "returns exact HTTP/1.1 200 OK response",
			want: "HTTP/1.1 200 OK\r\n\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HardcodedOK()
			if got != tt.want {
				t.Errorf("HardcodedOK() = %q, want %q", got, tt.want)
			}
		})
	}
}
