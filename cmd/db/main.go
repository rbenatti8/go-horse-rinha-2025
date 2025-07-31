package main

import (
	"encoding/binary"
	"fmt"
	"github.com/rbenatti8/go-horse-rinha-2025/internal/messages"
	"github.com/rbenatti8/go-horse-rinha-2025/internal/types"
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

			version := header[0]
			msgType := header[1]
			payloadSize := binary.BigEndian.Uint32(header[2:])

			if version != 1 || msgType != messages.TypePushPayment {
				log.Println("Unsupported version/type")
				return
			}

			payload := make([]byte, payloadSize)
			if _, err = conn.Read(payload); err != nil {
				log.Println("Failed to read payload:", err)
				return
			}

			var payment types.Payment
			if err = msgpack.Unmarshal(payload, &payment); err != nil {
				log.Println("Failed to decode msgpack:", err)
				return
			}

			fmt.Println("Received payment:", payment)
		}(conn)
	}
}
