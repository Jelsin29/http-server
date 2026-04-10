package request

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func Parse(r io.Reader) (*Request, error) {
	reader := bufio.NewReader(r)

	line, err := readLine(reader)
	if err != nil {
		return nil, fmt.Errorf("reading request line: %w", err)
	}

	method, target, version, err := parseRequestLine(line)
	if err != nil {
		return nil, err
	}

	headers, err := parseHeaders(reader)
	if err != nil {
		return nil, err
	}

	return &Request{
		Method:  method,
		Target:  target,
		Version: version,
		Headers: headers,
	}, nil
}

func parseRequestLine(line string) (string, string, string, error) {
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("parsing request line %q: expected method target version", line)
	}

	method := parts[0]
	target := parts[1]
	version := parts[2]

	if !strings.HasPrefix(version, "HTTP/") {
		return "", "", "", fmt.Errorf("parsing request line %q: invalid version %q", line, version)
	}

	return method, target, version, nil
}

func parseHeaders(reader *bufio.Reader) (map[string]string, error) {
	headers := make(map[string]string)

	for {
		line, err := readLine(reader)
		if err != nil {
			return nil, fmt.Errorf("reading headers: %w", err)
		}

		if line == "" {
			return headers, nil
		}

		name, value, found := strings.Cut(line, ":")
		if !found {
			return nil, fmt.Errorf("parsing header %q: missing colon", line)
		}

		if name == "" {
			return nil, fmt.Errorf("parsing header %q: empty name", line)
		}

		if strings.TrimSpace(name) != name {
			return nil, fmt.Errorf("parsing header %q: whitespace before colon is invalid", line)
		}

		headers[name] = strings.TrimSpace(value)
	}
}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")

	return line, nil
}
