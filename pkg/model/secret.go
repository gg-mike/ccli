package model

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/gg-mike/ccli/pkg/vault"
	"gorm.io/gorm"
)

type Secret struct {
	Key          string         `json:"key"            gorm:"primaryKey;uniqueIndex:idx_secrets"`
	ProjectName  sql.NullString `json:"-"              gorm:"uniqueIndex:idx_secrets"`
	PipelineName sql.NullString `json:"-"              gorm:"uniqueIndex:idx_secrets"`
	Path         string         `json:"path,omitempty"`
	CreatedAt    time.Time      `json:"created_at"     gorm:"default:now()"`
	UpdatedAt    time.Time      `json:"updated_at"     gorm:"default:now()"`
}

type SecretInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Path  string `json:"path"`
}

func (m *Secret) BeforeCreate(tx *gorm.DB) error {
	if strings.HasPrefix(m.Key, "_") {
		return errors.New("secret cannot start with '_'")
	}
	return nil
}

func (m *Secret) AfterCreate(tx *gorm.DB) error {
	value, ok := getValue(tx)
	if !ok {
		return errors.New("no value field given in instance")
	}
	return vault.SetStr(m.getUnique(), value)
}

func (m *Secret) AfterUpdate(tx *gorm.DB) error {
	if strings.HasPrefix(m.Key, "_") {
		return errors.New("secret cannot start with '_'")
	}
	if value, ok := getValue(tx); ok {
		return vault.SetStr(m.getUnique(), value)
	}
	return nil
}

func (m *Secret) AfterDelete(tx *gorm.DB) error {
	return vault.Del(m.getUnique())
}

func (m Secret) Value() (string, error) {
	return vault.GetStr(m.getUnique())
}

func (m *Secret) getUnique() string {
	unique := ""
	if m.ProjectName.Valid {
		unique += m.ProjectName.String + "/"
	}
	if m.PipelineName.Valid {
		unique += m.PipelineName.String + "/"
	}
	return unique + m.Key
}

func getValue(tx *gorm.DB) (string, bool) {
	input, ok := tx.InstanceGet("input")
	if !ok {
		return "", false
	}
	return input.(SecretInput).Value, ok
}
