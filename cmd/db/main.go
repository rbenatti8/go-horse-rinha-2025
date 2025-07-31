package main

import (
	"fmt"
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
				fmt.Printf("Mensagem recebida: %s\n", string(buf[:n]))
			}
		}(conn)
	}
}
