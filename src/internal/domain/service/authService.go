package service

import (
	"context"
	"messenger/src/internal/data/repository"
)

type AuthService struct {
	Repo *repository.PostgresRepo
}

func NewAuthService(repo *repository.PostgresRepo) *AuthService {
	return &AuthService{Repo: repo}
}

func (s *AuthService) SignUp(ctx context.Context, nick, password string) error {
	return s.Repo.CreateUser(ctx, nick, password)
}

func (s *AuthService) SignIn(ctx context.Context, nick, password string) (int, error) {
	return s.Repo.CheckUser(ctx, nick, password)
}
