package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

func main() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	defer conn.Close()

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			fmt.Println("Received:", string(msg))
		}
	}()

	conn.WriteMessage(websocket.TextMessage, []byte("Hello from Go client"))
	select {}
}
