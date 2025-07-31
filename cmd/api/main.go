package main

import (
	"github.com/rbenatti8/go-horse-rinha-2025/internal/web"
	"log"
)

func main() {
	s := web.NewServer(5000)
	err := s.Start()
	if err != nil {
		log.Fatal(err)
	}
}
