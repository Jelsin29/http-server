package static

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	root := t.TempDir()

	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("<h1>home</h1>"), 0o644); err != nil {
		t.Fatalf("writing index file: %v", err)
	}

	assetDir := filepath.Join(root, "assets")
	if err := os.MkdirAll(assetDir, 0o755); err != nil {
		t.Fatalf("creating assets dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(assetDir, "app.css"), []byte("body { color: red; }"), 0o644); err != nil {
		t.Fatalf("writing css file: %v", err)
	}

	tests := []struct {
		name            string
		target          string
		wantBody        string
		wantContentType string
		wantErr         error
	}{
		{
			name:            "maps root path to index file",
			target:          "/",
			wantBody:        "<h1>home</h1>",
			wantContentType: "text/html; charset=utf-8",
		},
		{
			name:            "loads nested asset",
			target:          "/assets/app.css",
			wantBody:        "body { color: red; }",
			wantContentType: "text/css; charset=utf-8",
		},
		{
			name:    "rejects traversal outside root",
			target:  "/../secret.txt",
			wantErr: ErrTraversal,
		},
		{
			name:    "returns missing file error",
			target:  "/missing.txt",
			wantErr: os.ErrNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset, err := Load(root, tt.target)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Load() error = %v, want %v", err, tt.wantErr)
				}

				if asset != nil {
					t.Fatalf("Load() asset = %#v, want nil", asset)
				}

				return
			}

			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if string(asset.Body) != tt.wantBody {
				t.Fatalf("Load() body = %q, want %q", string(asset.Body), tt.wantBody)
			}

			if asset.ContentType != tt.wantContentType {
				t.Fatalf("Load() content type = %q, want %q", asset.ContentType, tt.wantContentType)
			}
		})
	}
}

func TestContentTypeFor(t *testing.T) {
	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "html files use text html",
			file: "index.html",
			want: "text/html; charset=utf-8",
		},
		{
			name: "css files use text css",
			file: "assets/app.css",
			want: "text/css; charset=utf-8",
		},
		{
			name: "js files use javascript type",
			file: "assets/app.js",
			want: "application/javascript; charset=utf-8",
		},
		{
			name: "txt files use plain text",
			file: "notes.txt",
			want: "text/plain; charset=utf-8",
		},
		{
			name: "unknown files fall back to binary",
			file: "blob.bin",
			want: "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContentTypeFor(tt.file)
			if got != tt.want {
				t.Fatalf("ContentTypeFor() = %q, want %q", got, tt.want)
			}
		})
	}
}
