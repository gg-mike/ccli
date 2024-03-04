package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// TODO: cascading delete (secrets, issue: https://github.com/go-gorm/gorm/issues/5001)
type Pipeline struct {
	Name        string         `json:"name"                gorm:"primaryKey;uniqueIndex:idx_pipelines"`
	ProjectName string         `json:"-"                   gorm:"primaryKey;uniqueIndex:idx_pipelines"`
	Branch      string         `json:"branch"              gorm:"not null"`
	Config      PipelineConfig `json:"config"              gorm:"serializer:json;not null"`
	Variables   []Variable     `json:"variables,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:PipelineName,ProjectName"`
	Secrets     []Secret       `json:"secrets,omitempty"   gorm:"foreignKey:PipelineName,ProjectName"`
	Builds      []Build        `json:"builds,omitempty"    gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:PipelineName,ProjectName"`
	CreatedAt   time.Time      `json:"created_at"          gorm:"default:now()"`
	UpdatedAt   time.Time      `json:"updated_at"          gorm:"default:now()"`
}

type PipelineShort struct {
	Name      string    `json:"name"`
	Branch    string    `json:"branch"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PipelineInput struct {
	Name   string         `json:"name"`
	Branch string         `json:"branch"`
	Config PipelineConfig `json:"config"`
}

type PipelineConfig struct {
	System  string               `json:"system"`
	Image   string               `json:"image"`
	Steps   []PipelineConfigStep `json:"steps"`
	Cleanup []string             `json:"cleanup"`
}

type PipelineConfigStep struct {
	Name     string   `json:"name"`
	Commands []string `json:"commands"`
}

func (m *Pipeline) BeforeDelete(tx *gorm.DB) error {
	if !isForce(tx) {
		if len(m.Builds) == 0 {
			return nil
		}
		return errors.New("cannot delete project with builds (use 'force' query param to overwrite)")
	}

	for _, build := range m.Builds {
		if build.Status == BuildRunning {
			return errors.New("cannot delete pipeline with running builds")
		}
	}
	return nil
}
