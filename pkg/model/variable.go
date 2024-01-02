package model

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Variable struct {
	Key          string         `json:"key"            gorm:"primaryKey;uniqueIndex:idx_variables;not null"`
	ProjectName  sql.NullString `json:"-"              gorm:"uniqueIndex:idx_variables"`
	PipelineName sql.NullString `json:"-"              gorm:"uniqueIndex:idx_variables"`
	Value        string         `json:"value"          gorm:"not null"`
	Path         string         `json:"path,omitempty"`
	CreatedAt    time.Time      `json:"created_at"     gorm:"default:now();not null"`
	UpdatedAt    time.Time      `json:"updated_at"     gorm:"default:now();not null"`
}

type VariableInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Path  string `json:"path"`
}

func (m *Variable) BeforeCreate(tx *gorm.DB) error {
	if strings.HasPrefix(m.Key, "_") {
		return errors.New("variable cannot start with '_'")
	}
	return nil
}

func (m *Variable) AfterUpdate(tx *gorm.DB) error {
	if strings.HasPrefix(m.Key, "_") {
		return errors.New("variable cannot start with '_'")
	}
	return nil
}
