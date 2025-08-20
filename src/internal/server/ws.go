package server

import (
	"fmt"
	"net/http"
	"strconv"

	"messenger/src/internal/domain/service"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (s *Server) HandleWS(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Error(w, "invalid session", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}

	client := &service.Client{
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: userID,
	}
	s.hub.Register <- client

	lastMessages, err := s.chatService.GetLastMessages(100)
	if err == nil {
		for _, msg := range lastMessages {
			client.Send <- []byte(fmt.Sprintf("%s: %s", msg.Nick, msg.Content))
		}
	}

	go s.WritePump(client)
	go s.ReadPump(client)
}

func (s *Server) ReadPump(c *service.Client) {
	defer func() {
		s.hub.UnRegister <- c
		c.Conn.Close()
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		if err := s.chatService.SaveMessage(c.UserID, string(msg)); err != nil {
			s.log.Error("failed to save message", "err", err)
		}

		// достаём ник для рассылки
		lastMsgs, _ := s.chatService.GetLastMessages(1)
		if len(lastMsgs) > 0 {
			formattedMsg := fmt.Sprintf("%s: %s", lastMsgs[0].Nick, string(msg))
			s.hub.Broadcast <- []byte(formattedMsg)
		}
	}
}

func (s *Server) WritePump(c *service.Client) {
	for msg := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
}
