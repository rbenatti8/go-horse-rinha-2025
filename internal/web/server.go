package web

import (
	"bytes"
	"fmt"
	"github.com/panjf2000/gnet/v2"
	"log"
	"net"
)

var (
	statusOk                  = []byte("HTTP/1.1 200 Ok\r\nContent-Length: 0\r\n\r\n")
	statusNotFound            = []byte("HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\n\r\n")
	statusNotAllowed          = []byte("HTTP/1.1 405 Method Not Allowed\r\nContent-Length: 0\r\n\r\n")
	statusInternalServerError = []byte("HTTP/1.1 500 Internal Server Error\r\nContent-Length: 0\r\n\r\n")
	statusAccepted            = []byte("HTTP/1.1 202 Accepted\r\nContent-Length: 0\r\n\r\n")

	bOne   = []byte("\r\n")
	bTwo   = []byte(" ")
	bThree = []byte("?")
	bFour  = []byte("\r\n\r\n")
)

type Server struct {
	gnet.BuiltinEventEngine
	port int
}

func NewServer(port int) *Server {
	return &Server{
		port: port,
	}
}

func (s *Server) Start() error {
	err := gnet.Run(s, fmt.Sprintf("tcp://:%d", s.port),
		gnet.WithMulticore(true),
		gnet.WithReusePort(true),
		gnet.WithTCPNoDelay(gnet.TCPNoDelay),
		gnet.WithNumEventLoop(4),
		gnet.WithSocketRecvBuffer(4<<20),
		gnet.WithSocketSendBuffer(4<<20),
	)

	return err
}
func (s *Server) OnBoot(_ gnet.Engine) (action gnet.Action) {
	return
}

func (s *Server) OnShutdown(_ gnet.Engine) {}

func (s *Server) OnTraffic(c gnet.Conn) (action gnet.Action) {
	buf, err := c.Next(-1)
	if err != nil {
		return gnet.Close
	}

	requestLineEnd := bytes.Index(buf, bOne)
	if requestLineEnd == -1 {
		return gnet.Close
	}
	requestLineParts := bytes.Split(buf[:requestLineEnd], bTwo)
	if len(requestLineParts) != 3 {
		return gnet.Close
	}
	path, _, _ := bytes.Cut(requestLineParts[1], bThree)

	switch string(path) {
	case "/payments":
		handlePayments(c, buf)
		return gnet.None
	default:
		_, _ = c.Write(statusNotFound)
	}

	return
}

func handlePayments(c gnet.Conn, buf []byte) {
	_, _ = c.Write(statusAccepted)
	idx := bytes.Index(buf, bFour)
	if idx == -1 {
		_ = c.Close()
		return
	}

	msg := make([]byte, len(buf[idx+4:]))
	copy(msg, buf[idx+4:])

	sendTOWorker(msg)
	return
}

var addr = net.UnixAddr{Name: "/socket/worker.sock", Net: "unix"}

func sendTOWorker(msg []byte) {
	conn, err := net.DialUnix("unix", nil, &addr)
	if err != nil {
		log.Println("Error connecting to socket:", err)
		return
	}

	defer func(conn *net.UnixConn) {
		_ = conn.Close()
	}(conn)

	_, _ = conn.Write(msg)
}
