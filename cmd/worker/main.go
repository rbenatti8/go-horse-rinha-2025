package main

import (
	"bytes"
	"encoding/binary"
	"github.com/buger/jsonparser"
	"github.com/rbenatti8/go-horse-rinha-2025/internal/types"
	"github.com/vmihailenco/msgpack/v5"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	socketPath := "/socket/worker.sock"
	_ = os.Remove(socketPath)

	addr := net.UnixAddr{Name: socketPath, Net: "unix"}
	listener, err := net.ListenUnix("unix", &addr)
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

		go func(c *net.UnixConn) {
			defer func(c *net.UnixConn) {
				_ = c.Close()
			}(c)

			buf := make([]byte, 83)
			n, err := c.Read(buf)
			if err != nil {
				log.Println("Failed to read:", err)
				return
			}

			if n > 0 {
				cid, _ := jsonparser.GetString(buf, "correlationId")
				amount, err := jsonparser.GetFloat(buf, "amount")
				if err != nil {
					log.Println("Failed to parse amount:", err)
					return
				}

				p := types.Payment{
					Amount:      amount,
					CID:         cid,
					RequestedAt: time.Now().UTC().Format(time.RFC3339Nano),
				}

				sendTODB(p)
			}
		}(conn)
	}
}

var addr = net.UnixAddr{Name: "/socket/db.sock", Net: "unix"}

func sendTODB(p types.Payment) {
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
