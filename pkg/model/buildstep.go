package model

import (
	"time"
)

type BuildStep struct {
	Name         string        `json:"name"           gorm:"primaryKey;uniqueIndex:idx_build_steps"`
	BuildNumber  uint          `json:"-"              gorm:"primaryKey;uniqueIndex:idx_build_steps"`
	PipelineName string        `json:"-"              gorm:"primaryKey;uniqueIndex:idx_build_steps"`
	ProjectName  string        `json:"-"              gorm:"primaryKey;uniqueIndex:idx_build_steps"`
	Start        time.Time     `json:"start"`
	Duration     time.Duration `json:"duration"       gorm:"not null"`
	Logs         []BuildLog    `json:"logs,omitempty" gorm:"serializer:json"`
	CreatedAt    time.Time     `json:"created_at"     gorm:"default:now()"`
	UpdatedAt    time.Time     `json:"updated_at"     gorm:"default:now()"`
}

type BuildLog struct {
	Command string `json:"command"`
	Idx     int    `json:"idx,omitempty"`
	Total   int    `json:"total,omitempty"`
	Output  string `json:"output"`
}

func (step *BuildStep) AppendLog(log BuildLog) {
	step.Logs = append(step.Logs, log)
}

func (step *BuildStep) AppendOutput(output string) {
	if step.Logs[len(step.Logs)-1].Output != "" {
		step.Logs[len(step.Logs)-1].Output += "\n"
	}
	step.Logs[len(step.Logs)-1].Output += output
}

func (step *BuildStep) End() {
	step.Duration = time.Since(step.Start)
}
