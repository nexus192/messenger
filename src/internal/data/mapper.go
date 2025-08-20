package data

import (
	dataModel "messenger/src/internal/data/model"
	domainModel "messenger/src/internal/domain/model"
)

func ToDataMess(m domainModel.Message) dataModel.Message {
	return dataModel.Message{
		ID:        m.ID,
		UserID:    m.UserID,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}

func ToDataUser(m domainModel.User) dataModel.UserData {
	return dataModel.UserData{
		ID:       m.ID,
		Nick:     m.Nick,
		Password: m.Password,
	}
}

func ToDomainMess(m dataModel.Message) domainModel.Message {
	return domainModel.Message{
		ID:        m.ID,
		UserID:    m.UserID,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}
