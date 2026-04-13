package response

import (
	"bytes"
	"sort"
	"strconv"
)

const plainTextContentType = "text/plain; charset=utf-8"

type Message struct {
	StatusCode int
	Reason     string
	Headers    map[string]string
	Body       []byte
}

func Build(msg Message) []byte {
	var buffer bytes.Buffer

	buffer.WriteString("HTTP/1.1 ")
	buffer.WriteString(strconv.Itoa(msg.StatusCode))
	buffer.WriteByte(' ')
	buffer.WriteString(msg.Reason)
	buffer.WriteString("\r\n")

	keys := sortedHeaderKeys(msg.Headers)
	for _, key := range keys {
		buffer.WriteString(key)
		buffer.WriteString(": ")
		buffer.WriteString(msg.Headers[key])
		buffer.WriteString("\r\n")
	}

	buffer.WriteString("\r\n")
	buffer.Write(msg.Body)

	return buffer.Bytes()
}

func PlainText(statusCode int, reason string, body string) []byte {
	bodyBytes := []byte(body)

	return Build(Message{
		StatusCode: statusCode,
		Reason:     reason,
		Headers: map[string]string{
			"Content-Length": strconv.Itoa(len(bodyBytes)),
			"Content-Type":   plainTextContentType,
		},
		Body: bodyBytes,
	})
}

func sortedHeaderKeys(headers map[string]string) []string {
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}
