package repository

import (
	"context"
	"messenger/src/internal/domain/model"
)

type Repository interface {
	GetLastMessages(limit int) ([]model.Message, error)
	SaveMessage(userID int, message string) error
	GetHistory(userID int) ([]model.Message, error)
	GetNickByID(userID int) (string, error)
	CreateUser(ctx context.Context, nick, password string) error
	CheckUser(ctx context.Context, nick, password string) (int, error)
}
