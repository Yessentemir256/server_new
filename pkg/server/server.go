package server

import (
	"log"
	"net"
	"strings"
	"sync"
)

type HandlerFunc func(conn net.Conn)

type Server struct {
	addr     string
	mu       sync.RWMutex
	handlers map[string]HandlerFunc
}

func NewServer(addr string) *Server {
	return &Server{addr: addr, handlers: make(map[string]HandlerFunc)}
}

func (s *Server) Register(path string, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[path] = handler
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		go func() {
			defer conn.Close()

			path := "/" // Default path
			buf := make([]byte, 4096)
			n, err := conn.Read(buf)
			if err != nil {
				log.Print(err)
				return
			}
			reqLine := string(buf[:n])
			parts := strings.Split(reqLine, " ")
			if len(parts) > 1 {
				path = parts[1]
			}

			s.mu.RLock()
			handler, ok := s.handlers[path]
			s.mu.RUnlock()

			if !ok {
				return
			}

			handler(conn)
		}()
	}
}
