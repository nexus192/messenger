package service

import (
	"messenger/src/internal/data/repository"
)

type ChatService struct {
	Repo *repository.PostgresRepo
	Hub  *Hub
}

func NewChatService(repo *repository.PostgresRepo, hub *Hub) *ChatService {
	return &ChatService{Repo: repo, Hub: hub}
}

func (s *ChatService) SaveMessage(userID int, content string) error {
	return s.Repo.SaveMessage(userID, content)
}

func (s *ChatService) GetLastMessages(limit int) ([]MessageDTO, error) {
	messages, err := s.Repo.GetLastMessages(limit)
	if err != nil {
		return nil, err
	}

	var result []MessageDTO
	for _, m := range messages {
		nick, err := s.Repo.GetNickByID(m.UserID)
		if err != nil {
			nick = "unknown"
		}
		result = append(result, MessageDTO{
			Nick:    nick,
			Content: m.Content,
		})
	}
	return result, nil
}

// DTO для транспорта
type MessageDTO struct {
	Nick    string
	Content string
}
