package web

import (
	"bytes"
	"encoding/binary"
	"fmt"
	goJson "github.com/goccy/go-json"
	"github.com/panjf2000/gnet/v2"
	"github.com/rbenatti8/go-horse-rinha-2025/internal/messages"
	"github.com/vmihailenco/msgpack/v5"
	"io"
	"log"
	"net"
	"net/url"
	"time"
)

var (
	statusOk                  = []byte("HTTP/1.1 200 Ok\r\nContent-Length: %d\r\n\r\n")
	statusNotFound            = []byte("HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\n\r\n")
	statusNotAllowed          = []byte("HTTP/1.1 405 Method Not Allowed\r\nContent-Length: 0\r\n\r\n")
	statusInternalServerError = []byte("HTTP/1.1 500 Internal Server Error\r\nContent-Length: 0\r\n\r\n")
	statusAccepted            = []byte("HTTP/1.1 202 Accepted\r\nContent-Length: 0\r\n\r\n")

	bOne   = []byte("\r\n")
	bTwo   = []byte(" ")
	bThree = []byte("?")
	bFour  = []byte("\r\n\r\n")

	workerAddr = net.UnixAddr{Name: "/socket/worker.sock", Net: "unix"}
	dbAddr     = net.UnixAddr{Name: "/socket/db.sock", Net: "unix"}
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
		gnet.WithSocketRecvBuffer(1<<20), // 1MB
		gnet.WithSocketSendBuffer(1<<20), // 1MB
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
	path, query, _ := bytes.Cut(requestLineParts[1], bThree)

	switch string(path) {
	case "/payments":
		handlePayments(c, buf)
		return gnet.None
	case "/payments-summary":
		handlePaymentsSummary(c, query)
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

func handlePaymentsSummary(c gnet.Conn, query []byte) {
	params, _ := url.ParseQuery(string(query))

	from := params.Get("from")
	to := params.Get("to")

	m := messages.SummarizePayments{
		From: from,
		To:   to,
	}

	payload, err := msgpack.Marshal(m)
	if err != nil {
		return
	}

	var buf bytes.Buffer

	buf.WriteByte(1)                                               // version
	buf.WriteByte(messages.TypeSummarizePayments)                  // type = insert
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(payload))) // payload size
	buf.Write(payload)

	conn, err := net.DialUnix("unix", nil, &dbAddr)
	if err != nil {
		log.Println("Error connecting to socket:", err)
		return
	}

	defer func(conn *net.UnixConn) {
		_ = conn.Close()
	}(conn)

	_, _ = conn.Write(buf.Bytes())

	header := make([]byte, 6) // 1 + 1 + 4
	if _, err := conn.Read(header); err != nil {
		log.Println("Failed to read header:", err)
		return
	}

	//version := header[0]
	// msgType := header[1]
	payloadSize := binary.BigEndian.Uint32(header[2:])

	resp := make([]byte, payloadSize)
	n, err := conn.Read(resp)
	if err != nil && err != io.EOF {
		log.Println("Error reading from socket:", err)
		_, _ = c.Write(statusInternalServerError)
		return
	}

	var summary messages.SummarizedPayments
	if err = msgpack.Unmarshal(resp[:n], &summary); err != nil {
		log.Println("Failed to decode msgpack:", err)
		_, _ = c.Write(statusInternalServerError)
		return
	}

	b, _ := goJson.Marshal(summary)
	h := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Length: %d\r\nContent-Type: application/json\r\n\r\n", len(b))
	c.Write([]byte(h))
	c.Write(b)
}

func sendTOWorker(msg []byte) {
	t := time.Now()
	conn, err := net.DialUnix("unix", nil, &workerAddr)
	if err != nil {
		log.Println("Error connecting to socket:", err)
		return
	}

	defer func(conn *net.UnixConn) {
		_ = conn.Close()
	}(conn)

	_, _ = conn.Write(msg)

	log.Printf("Time taken: %s", time.Since(t))
}
