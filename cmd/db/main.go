package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/rbenatti8/go-horse-rinha-2025/internal/messages"
	"github.com/vmihailenco/msgpack/v5"
	"log"
	"net"
	"os"
)

func main() {
	socketPath := "/socket/db.sock"
	_ = os.Remove(socketPath)

	addr := net.UnixAddr{Name: socketPath, Net: "unix"}
	listener, err := net.ListenUnix("unix", &addr)
	if err != nil {
		log.Fatal("error listening socket:", err)
	}

	defer func(listener *net.UnixListener) {
		_ = listener.Close()
	}(listener)

	fmt.Println("Database listening on socket...")

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

			header := make([]byte, 6) // 1 + 1 + 4
			if _, err := conn.Read(header); err != nil {
				log.Println("Failed to read header:", err)
				return
			}

			//version := header[0]
			msgType := header[1]
			payloadSize := binary.BigEndian.Uint32(header[2:])

			if msgType == messages.TypePushPayment {
				payload := make([]byte, payloadSize)
				if _, err = conn.Read(payload); err != nil {
					log.Println("Failed to read payload:", err)
					return
				}

				var payment messages.Payment
				if err = msgpack.Unmarshal(payload, &payment); err != nil {
					log.Println("Failed to decode msgpack:", err)
					return
				}
			}

			if msgType == messages.TypeSummarizePayments {
				payload := make([]byte, payloadSize)
				if _, err = conn.Read(payload); err != nil {
					log.Println("Failed to read payload:", err)
					return
				}

				var summary messages.SummarizePayments
				if err = msgpack.Unmarshal(payload, &summary); err != nil {
					log.Println("Failed to decode msgpack:", err)
					return
				}

				fmt.Println("Received payment summary request:", summary)

				m := messages.SummarizedPayments{
					Default: messages.SummarizedProcessor{
						TotalAmount:   0,
						TotalRequests: 0,
					},
					Fallback: messages.SummarizedProcessor{
						TotalAmount:   0,
						TotalRequests: 0,
					},
				}

				summarizedPayments, err := msgpack.Marshal(m)
				if err != nil {
					log.Println("Failed to encode msgpack:", err)
					return
				}

				var buf bytes.Buffer

				buf.WriteByte(1)                                                          // version
				buf.WriteByte(messages.TypeSummarizedPayments)                            // type = insert
				_ = binary.Write(&buf, binary.BigEndian, uint32(len(summarizedPayments))) // payload size
				buf.Write(summarizedPayments)

				_, _ = conn.Write(buf.Bytes())
			}

		}(conn)
	}
}
