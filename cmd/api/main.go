package main

import (
	"github.com/rbenatti8/go-horse-rinha-2025/internal/env"
	"github.com/rbenatti8/go-horse-rinha-2025/internal/web"
	"log"
	"os"
	"time"
)

func main() {
	socketPath := env.GetEnvAsString("SOCKET_PATH", "/socket/pod.sock")

	_ = os.Remove(socketPath)
	s, err := web.NewServer(socketPath)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	go func() {
		if err := s.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	for {
		if _, err := os.Stat(socketPath); err == nil {
			err := os.Chmod(socketPath, 0666)
			if err != nil {
				log.Fatalf("Failed to change socket permissions: %v", err)
			}

			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	select {}
}
