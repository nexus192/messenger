// package server

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"strconv"

// 	"messenger/src/internal/data/repository"
// 	"messenger/src/internal/hub"

// 	"log/slog"

// 	"github.com/gorilla/websocket"
// )

// type WSMessage struct {
// 	Nick    string `json:"nick"`
// 	Content string `json:"content"`
// }

// type Server struct {
// 	hub  *hub.Hub
// 	repo *repository.PostgresRepo
// 	log  *slog.Logger
// }

// func NewServer(h *hub.Hub, repo *repository.PostgresRepo, log *slog.Logger) *Server {
// 	return &Server{hub: h, repo: repo, log: log}
// }

// var upgrader = websocket.Upgrader{
// 	CheckOrigin: func(r *http.Request) bool {
// 		return true // пока разрешаем всем
// 	},
// }

// type APIResponse struct {
// 	Success bool   `json:"success"`
// 	Message string `json:"message"`
// }

package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"messenger/src/internal/domain/service"
)

type Server struct {
	hub         *service.Hub
	authService *service.AuthService
	chatService *service.ChatService
	log         *slog.Logger
}

func NewServer(h *service.Hub, auth *service.AuthService, chat *service.ChatService, log *slog.Logger) *Server {
	return &Server{hub: h, authService: auth, chatService: chat, log: log}
}

func (s *Server) Start() {
	// WebSocket
	http.HandleFunc("/ws", s.HandleWS)

	// API
	http.HandleFunc("/api/signup", s.HandleSignUp)
	http.HandleFunc("/api/signin", s.HandleSignIn)

	// Файлы
	fs := http.FileServer(http.Dir("internal/client/web"))

	// Явно указываем разрешённые "публичные" страницы
	http.Handle("/signin.html", fs)
	http.Handle("/signup.html", fs)

	http.HandleFunc("/api/logout", s.HandleLogout)

	// Всё остальное защищаем middleware
	http.Handle("/",
		s.AuthMiddleware(fs),
	)

	port := 8080
	fmt.Printf("Server started at %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func (s *Server) HandleLogout(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/signin.html", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
