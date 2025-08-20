package adapter

import (
	"context"
	"messenger/src/internal/data"
	dataRepo "messenger/src/internal/data/repository"
	"messenger/src/internal/domain/model"
)

type PostgresAdapter struct {
	Repo *dataRepo.PostgresRepo
}

func (r *PostgresAdapter) CreateUser(ctx context.Context, nick, password string) error {
	return r.Repo.CreateUser(ctx, nick, password)
}

func (r *PostgresAdapter) CheckUser(ctx context.Context, nick, password string) (int, error) {
	return r.Repo.CheckUser(ctx, nick, password)
}

func (r *PostgresAdapter) SaveMessage(userID int, message string) error {
	return r.Repo.SaveMessage(userID, message)
}

func (r *PostgresAdapter) GetHistory(userID int) ([]model.Message, error) {
	dataMessages, err := r.Repo.GetHistory(userID)
	if err != nil {
		return nil, err
	}

	domainMessages := make([]model.Message, len(dataMessages))
	for i, dm := range dataMessages {
		domainMessages[i] = data.ToDomainMess(dm)
	}

	return domainMessages, nil
}

func (r *PostgresAdapter) GetLastMessages(limit int) ([]model.Message, error) {
	dataMessages, err := r.Repo.GetLastMessages(limit)
	if err != nil {
		return nil, err
	}

	domainMessages := make([]model.Message, len(dataMessages))
	for i, dm := range dataMessages {
		domainMessages[i] = data.ToDomainMess(dm)
	}

	return domainMessages, nil
}

func (r *PostgresAdapter) GetNickByID(userID int) (string, error) {
	return r.Repo.GetNickByID(userID)
}
