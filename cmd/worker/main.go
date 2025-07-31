package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	socketPath := "/socket/worker.sock"
	_ = os.Remove(socketPath)

	addr := net.UnixAddr{Name: socketPath, Net: "unix"}
	listener, err := net.ListenUnix("unix", &addr)
	if err != nil {
		log.Fatal("Erro ao escutar socket:", err)
	}
	defer listener.Close()
	fmt.Println("Worker ouvindo no socket...")

	for {
		conn, err := listener.AcceptUnix()
		if err != nil {
			log.Println("Erro ao aceitar conexão:", err)
			continue
		}

		go func(c *net.UnixConn) {
			defer func(c *net.UnixConn) {
				_ = c.Close()
			}(c)

			buf := make([]byte, 70)
			n, err := c.Read(buf)
			if err != nil {
				log.Println("Erro ao ler do socket:", err)
				return
			}

			if n > 0 {
				sendTODB(buf)
			}
		}(conn)
	}
}

var addr = net.UnixAddr{Name: "/socket/db.sock", Net: "unix"}

func sendTODB(msg []byte) {
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
