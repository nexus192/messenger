package main

import (
	"messenger/src/internal/hub"
	"messenger/src/internal/server"
)

func main() {
	h := hub.NewHub()
	s := server.NewServer(h)

	go h.Run()
	s.Start()
}
