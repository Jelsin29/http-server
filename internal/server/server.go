package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/jelsin/http-server/internal/request"
	"github.com/jelsin/http-server/internal/response"
	"github.com/jelsin/http-server/internal/static"
)

var documentRoot = locateDocumentRoot()
var loadAsset = static.Load
var renderMessage = func(msg response.Message) ([]byte, error) {
	return response.Build(msg), nil
}
var handleAcceptedConnection = handleConnection

func Start(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("starting server on %s: %w", addr, err)
	}
	defer ln.Close()

	return serve(ln)
}

func serve(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}

			log.Printf("accept error: %v", err)
			continue
		}

		go handleAcceptedConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	req, err := request.Parse(conn)
	if err != nil {
		log.Printf("parsing request from %s: %v", conn.RemoteAddr(), err)
		writeResponse(conn, response.PlainText(400, "Bad Request", "bad request"))
		return
	}

	responseBytes, err := buildResponse(req)
	if err != nil {
		log.Printf("building response for %s %s: %v", req.Method, req.Target, err)
		writeResponse(conn, response.PlainText(500, "Internal Server Error", "internal server error"))
		return
	}

	writeResponse(conn, responseBytes)
}

func buildResponse(req *request.Request) ([]byte, error) {
	asset, err := loadAsset(documentRoot, req.Target)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, static.ErrTraversal) {
			return response.PlainText(404, "Not Found", "not found"), nil
		}

		return nil, err
	}

	responseBytes, err := renderMessage(response.Message{
		StatusCode: 200,
		Reason:     "OK",
		Headers: map[string]string{
			"Content-Length": strconv.Itoa(len(asset.Body)),
			"Content-Type":   asset.ContentType,
		},
		Body: asset.Body,
	})
	if err != nil {
		return nil, fmt.Errorf("rendering response: %w", err)
	}

	return responseBytes, nil
}

func writeResponse(conn net.Conn, responseBytes []byte) {
	if _, err := conn.Write(responseBytes); err != nil {
		log.Printf("writing response to %s: %v", conn.RemoteAddr(), err)
	}
}

func locateDocumentRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "public"
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "public"))
}
