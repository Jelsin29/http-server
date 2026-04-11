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

func Start(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("starting server on %s: %w", addr, err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}

		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	req, err := request.Parse(conn)
	if err != nil {
		log.Printf("parsing request from %s: %v", conn.RemoteAddr(), err)
		return
	}

	responseBytes, err := buildResponse(req)
	if err != nil {
		log.Printf("building response for %s %s: %v", req.Method, req.Target, err)
		return
	}

	if _, err := conn.Write(responseBytes); err != nil {
		log.Printf("writing response to %s: %v", conn.RemoteAddr(), err)
	}
}

func buildResponse(req *request.Request) ([]byte, error) {
	asset, err := static.Load(documentRoot, req.Target)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, static.ErrTraversal) {
			return []byte(response.Build(response.Message{
				StatusCode: 404,
				Reason:     "Not Found",
				Headers: map[string]string{
					"Content-Length": strconv.Itoa(len("not found")),
					"Content-Type":   "text/plain; charset=utf-8",
				},
				Body: []byte("not found"),
			})), nil
		}

		return nil, err
	}

	return response.Build(response.Message{
		StatusCode: 200,
		Reason:     "OK",
		Headers: map[string]string{
			"Content-Length": strconv.Itoa(len(asset.Body)),
			"Content-Type":   asset.ContentType,
		},
		Body: asset.Body,
	}), nil
}

func locateDocumentRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "public"
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "public"))
}
