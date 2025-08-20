package service

import (
	"context"
	"messenger/src/internal/domain/repository"
)

type AuthService struct {
	Repo repository.Repository
}

func NewAuthService(repo repository.Repository) *AuthService {
	return &AuthService{Repo: repo}
}

func (s *AuthService) SignUp(ctx context.Context, nick, password string) error {
	return s.Repo.CreateUser(ctx, nick, password)
}

func (s *AuthService) SignIn(ctx context.Context, nick, password string) (int, error) {
	return s.Repo.CheckUser(ctx, nick, password)
}
