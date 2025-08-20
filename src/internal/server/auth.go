package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (s *Server) HandleSignUp(w http.ResponseWriter, r *http.Request) {
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

	err := s.authService.SignUp(r.Context(), req.Nick, req.Password)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "signup failed"})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{Success: true, Message: "signup success"})
}

func (s *Server) HandleSignIn(w http.ResponseWriter, r *http.Request) {
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

	userID, err := s.authService.SignIn(r.Context(), req.Nick, req.Password)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "invalid credentials"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    fmt.Sprintf("%d", userID),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600,
	})

	json.NewEncoder(w).Encode(APIResponse{Success: true, Message: "login success"})
}

// func (s *Server) HandleLogout(w http.ResponseWriter, r *http.Request) {
// 	// Затираем куку с нулевым временем жизни
// 	http.SetCookie(w, &http.Cookie{
// 		Name:     "session_id",
// 		Value:    "",
// 		Path:     "/",
// 		HttpOnly: true,
// 		MaxAge:   -1, // удалить
// 	})

// 	http.Redirect(w, r, "/signin.html", http.StatusSeeOther)
// }

// func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		cookie, err := r.Cookie("session_id")
// 		if err != nil || cookie.Value == "" {
// 			http.Redirect(w, r, "/signin.html", http.StatusSeeOther)
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }
