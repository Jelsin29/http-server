package server

import (
	"fmt"
	"log"
	"net"

	"github.com/jelsin/http-server/internal/request"
	"github.com/jelsin/http-server/internal/response"
)

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

	if _, err := request.Parse(conn); err != nil {
		log.Printf("parsing request from %s: %v", conn.RemoteAddr(), err)
		return
	}

	if _, err := conn.Write([]byte(response.HardcodedOK())); err != nil {
		log.Printf("writing response to %s: %v", conn.RemoteAddr(), err)
	}
}
