package main

import (
	"bytes"
	"encoding/binary"
	"github.com/buger/jsonparser"
	"github.com/rbenatti8/go-horse-rinha-2025/internal/messages"
	"github.com/vmihailenco/msgpack/v5"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	socketPath := "/socket/worker.sock"
	_ = os.Remove(socketPath)

	dbAddr := net.UnixAddr{Name: socketPath, Net: "unix"}
	listener, err := net.ListenUnix("unix", &dbAddr)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}
	defer listener.Close()

	log.Println("Listening on", socketPath)

	for {
		conn, err := listener.AcceptUnix()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn *net.UnixConn) {
	defer func(c *net.UnixConn) {
		_ = c.Close()
	}(conn)

	buf := make([]byte, 83)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("Failed to read:", err)
		return
	}

	if n > 0 {
		cid, _ := jsonparser.GetString(buf, "correlationId")
		amount, _ := jsonparser.GetFloat(buf, "amount")

		p := messages.Payment{
			Amount:      amount,
			CID:         cid,
			RequestedAt: time.Now().UTC().Format(time.RFC3339Nano),
		}

		sendTODB(p)
	}
}

var addr = net.UnixAddr{Name: "/socket/db.sock", Net: "unix"}

func sendTODB(p messages.Payment) {
	conn, err := net.DialUnix("unix", nil, &addr)
	if err != nil {
		log.Println("Error connecting to socket:", err)
		return
	}

	defer func(conn *net.UnixConn) {
		_ = conn.Close()
	}(conn)

	payload, err := msgpack.Marshal(p)
	if err != nil {
		return
	}

	var buf bytes.Buffer

	buf.WriteByte(1)                                               // version
	buf.WriteByte(1)                                               // type = insert
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(payload))) // payload size
	buf.Write(payload)

	_, _ = conn.Write(buf.Bytes())
}
