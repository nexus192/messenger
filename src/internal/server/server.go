package server

import (
	"fmt"
	"net/http"

	"messenger/src/internal/hub"

	"github.com/gorilla/websocket"
)

type Server struct {
	hub *hub.Hub
}

func NewServer(h *hub.Hub) *Server {
	return &Server{hub: h}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // пока разрешаем всем
	},
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}

	client := &hub.Client{Conn: conn, Send: make(chan []byte, 256)}
	s.hub.Register <- client

	go s.writePump(client)
	go s.readPump(client)
}

func (s *Server) readPump(c *hub.Client) {
	defer func() {
		s.hub.UnRegister <- c
		c.Conn.Close()
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		s.hub.Broadcast <- msg
	}
}

func (s *Server) writePump(c *hub.Client) {
	for msg := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
}

func (s *Server) Start() {
	http.HandleFunc("/ws", s.handleWS)
	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
