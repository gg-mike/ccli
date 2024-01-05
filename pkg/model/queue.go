package model

import (
	"time"
)

type QueueElem struct {
	ID        string       `json:"id"         gorm:"primaryKey"`
	Context   QueueContext `json:"context"    gorm:"serializer:json;not null"`
	CreatedAt time.Time    `json:"created_at" gorm:"default:now()"`
}

type QueueElemShort struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type QueueContext struct {
	Build     Build
	Repo      string
	Branch    string
	Config    PipelineConfig
	Secrets   []Secret
	Variables []Variable
	Worker    Worker
}

func (QueueElem) TableName() string {
	return "queue"
}
