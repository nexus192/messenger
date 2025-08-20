package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"messenger/src/internal/data/repository"
	"messenger/src/internal/hub"

	"log/slog"

	"github.com/gorilla/websocket"
)

type WSMessage struct {
	Nick    string `json:"nick"`
	Content string `json:"content"`
}

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
	// проверяем куку
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

	client := &hub.Client{
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: userID, // добавляем UserID
	}
	s.hub.Register <- client

	lastMessages, err := s.repo.GetLastMessages(100)
	if err != nil {
		s.log.Error("failed to load last messages", "err", err)
	} else {
		for _, msg := range lastMessages {
			nick, err := s.repo.GetNickByID(msg.UserID)
			if err != nil {
				nick = "unknown"
			}
			client.Send <- []byte(fmt.Sprintf("%s: %s", nick, msg.Content))
		}
	}

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

		nick, err := s.repo.GetNickByID(c.UserID)
		if err != nil {
			s.log.Error("failed to get user nick", "err", err)
			nick = "unknown"
		}

		if err := s.repo.SaveMessage(c.UserID, string(msg)); err != nil {
			s.log.Error("failed to save message", "err", err)
		}

		// ✅ Ник подставляем только для рассылки
		formattedMsg := fmt.Sprintf("%s: %s", nick, string(msg))
		s.hub.Broadcast <- []byte(formattedMsg)
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

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    fmt.Sprintf("%d", userID),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600,
	})

	s.log.Info("user signed in", "userID", userID, "nick", req.Nick)
	json.NewEncoder(w).Encode(APIResponse{Success: true, Message: "Login successful"})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Затираем куку с нулевым временем жизни
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // удалить
	})

	http.Redirect(w, r, "/signin.html", http.StatusSeeOther)
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/signin.html", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) Start() {
	// WebSocket
	http.HandleFunc("/ws", s.handleWS)

	// API
	http.HandleFunc("/api/signup", s.handleSignUp)
	http.HandleFunc("/api/signin", s.handleSignIn)

	// Файлы
	fs := http.FileServer(http.Dir("web"))

	// Явно указываем разрешённые "публичные" страницы
	http.Handle("/signin.html", fs)
	http.Handle("/signup.html", fs)

	http.HandleFunc("/api/logout", s.handleLogout)

	// Всё остальное защищаем middleware
	http.Handle("/",
		s.authMiddleware(fs),
	)

	port := 8080
	fmt.Printf("Server started at %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
