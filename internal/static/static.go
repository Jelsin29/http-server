package static

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var ErrTraversal = errors.New("path escapes public root")

type Asset struct {
	Body        []byte
	ContentType string
}

var contentTypes = map[string]string{
	".html": "text/html; charset=utf-8",
	".css":  "text/css; charset=utf-8",
	".js":   "application/javascript; charset=utf-8",
	".txt":  "text/plain; charset=utf-8",
}

func Load(root string, target string) (*Asset, error) {
	filePath, err := resolvePath(root, target)
	if err != nil {
		return nil, err
	}

	body, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading asset %s: %w", filePath, err)
	}

	return &Asset{
		Body:        body,
		ContentType: ContentTypeFor(filePath),
	}, nil
}

func ContentTypeFor(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if contentType, ok := contentTypes[ext]; ok {
		return contentType
	}

	return "application/octet-stream"
}

func resolvePath(root string, target string) (string, error) {
	cleanTarget, err := cleanTarget(target)
	if err != nil {
		return "", err
	}

	return filepath.Join(root, filepath.FromSlash(cleanTarget)), nil
}

func cleanTarget(target string) (string, error) {
	rawPath := stripQuery(target)
	if rawPath == "" {
		rawPath = "/"
	}

	if !strings.HasPrefix(rawPath, "/") {
		return "", fmt.Errorf("invalid target %q: must start with slash", target)
	}

	segments := strings.Split(strings.TrimPrefix(rawPath, "/"), "/")
	stack := make([]string, 0, len(segments))

	for _, segment := range segments {
		switch segment {
		case "", ".":
			continue
		case "..":
			if len(stack) == 0 {
				return "", ErrTraversal
			}

			stack = stack[:len(stack)-1]
		default:
			stack = append(stack, segment)
		}
	}

	if len(stack) == 0 {
		return "index.html", nil
	}

	return path.Join(stack...), nil
}

func stripQuery(target string) string {
	withoutQuery, _, _ := strings.Cut(target, "?")
	withoutFragment, _, _ := strings.Cut(withoutQuery, "#")
	return withoutFragment
}
