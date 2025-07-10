package services

import (
"gorm.io/gorm"
)

type ConversationService struct {
db *gorm.DB
}

func NewConversationService(db *gorm.DB) *ConversationService {
return &ConversationService{db: db}
}

func (s *ConversationService) GetDB() *gorm.DB {
return s.db
}
