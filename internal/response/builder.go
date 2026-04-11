package response

import (
	"bytes"
	"sort"
	"strconv"
)

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

func HardcodedOK() string {
	return string(Build(Message{
		StatusCode: 200,
		Reason:     "OK",
	}))
}

func sortedHeaderKeys(headers map[string]string) []string {
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}
