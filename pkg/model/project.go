package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// TODO: cascading delete (secrets, issue: https://github.com/go-gorm/gorm/issues/5001)
type Project struct {
	Name      string     `json:"name"                gorm:"primaryKey"`
	Repo      string     `json:"repo"                gorm:"not null"`
	Variables []Variable `json:"variables,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Secrets   []Secret   `json:"secrets,omitempty"`
	Pipelines []Pipeline `json:"pipelines,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt time.Time  `json:"created_at"          gorm:"default:now()"`
	UpdatedAt time.Time  `json:"updated_at"          gorm:"default:now()"`
}

type ProjectShort struct {
	Name      string    `json:"name"`
	Repo      string    `json:"repo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProjectInput struct {
	Name string `json:"name"`
	Repo string `json:"repo"`
}

func (m *Project) BeforeDelete(tx *gorm.DB) error {
	if !isForce(tx) {
		if len(m.Pipelines) == 0 {
			return nil
		}
		return errors.New("cannot delete project with pipelines (use 'force' query param to overwrite)")
	}

	var count int64
	tx.Joins("builds", tx.Where(&Build{Status: BuildRunning})).Count(&count).
		Find(nil, &Pipeline{ProjectName: m.Name})

	if count == 0 {
		return nil
	}
	return errors.New("cannot delete project with running builds")
}
