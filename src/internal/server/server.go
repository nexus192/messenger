package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"messenger/src/internal/data/repository"
	"messenger/src/internal/hub"

	"log/slog"

	"github.com/gorilla/websocket"
)

type Server struct {
	hub  *hub.Hub
	repo *repository.PostgresRepo
	log  *slog.Logger
}

func NewServer(h *hub.Hub, repo *repository.PostgresRepo, log *slog.Logger) *Server {
	return &Server{hub: h, repo: repo, log: log}
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

// ============ NEW API HANDLERS ============

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (s *Server) handleSignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Nick     string `json:"nick"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	err := s.repo.CreateUser(r.Context(), req.Nick, req.Password)
	if err != nil {
		s.log.Error("signup failed", "err", err)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "User already exists or DB error"})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{Success: true, Message: "User created successfully"})
}

func (s *Server) handleSignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Nick     string `json:"nick"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	userID, err := s.repo.CheckUser(r.Context(), req.Nick, req.Password)
	if err != nil {
		s.log.Warn("signin failed", "err", err)
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Invalid credentials"})
		return
	}

	s.log.Info("user signed in", "userID", userID, "nick", req.Nick)
	json.NewEncoder(w).Encode(APIResponse{Success: true, Message: "Login successful"})
}

// ==========================================

// utils
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "unknown"
}

func (s *Server) Start() {
	// WebSocket
	http.HandleFunc("/ws", s.handleWS)

	// API
	http.HandleFunc("/api/signup", s.handleSignUp)
	http.HandleFunc("/api/signin", s.handleSignIn)

	// web (mess.html, signin.html, signup.html)
	http.Handle("/", http.FileServer(http.Dir("web")))

	port := 8080
	fmt.Printf("Server started at %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
