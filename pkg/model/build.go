package model

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type BuildStatus uint8

const (
	BuildScheduled BuildStatus = iota
	BuildRunning
	BuildSuccessful
	BuildFailed
	BuildCanceled
)

func (s BuildStatus) String() string {
	switch s {
	case BuildScheduled:
		return "scheduled"
	case BuildRunning:
		return "running"
	case BuildSuccessful:
		return "successful"
	case BuildFailed:
		return "failed"
	case BuildCanceled:
		return "canceled"
	default:
		return "unknown"
	}
}

type Build struct {
	Number       uint           `json:"number"          gorm:"primaryKey;uniqueIndex:idx_builds"`
	PipelineName string         `json:"-"               gorm:"primaryKey;uniqueIndex:idx_builds"`
	ProjectName  string         `json:"-"               gorm:"primaryKey;uniqueIndex:idx_builds"`
	Status       BuildStatus    `json:"status"          gorm:"default:0"`
	Steps        []BuildStep    `json:"steps,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:BuildNumber,PipelineName,ProjectName"`
	WorkerName   sql.NullString `json:"worker_name"`
	CreatedAt    time.Time      `json:"created_at"      gorm:"default:now()"`
	UpdatedAt    time.Time      `json:"updated_at"      gorm:"default:now()"`
}

type BuildShort struct {
	Number     uint           `json:"number"`
	Status     BuildStatus    `json:"status"`
	WorkerName sql.NullString `json:"worker_name,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func (m *Build) BeforeCreate(tx *gorm.DB) error {
	var result uint
	if err := tx.Model(&Build{}).Where(&Build{PipelineName: m.PipelineName, ProjectName: m.ProjectName}).Select("max(number)").Row().Scan(&result); err != nil {
		if err.Error() == `sql: Scan error on column index 0, name "max": converting NULL to uint is unsupported` {
			result = 0
		} else {
			return err
		}
	}
	m.Number = result + 1
	return nil
}

func (m *Build) AfterCreate(tx *gorm.DB) error {
	// TODO: schedule build
	return nil
}

func (m *Build) BeforeUpdate(tx *gorm.DB) error {
	prev, ok := tx.InstanceGet("prev")
	if !ok {
		return errors.New("prev build not given")
	}
	switch prev.(Build).Status {
	case BuildScheduled, BuildRunning:
		tx.Statement.SetColumn("status", BuildCanceled)
		// TODO: cancel build
		return nil
	default:
		return fmt.Errorf("cannot change status of build from [%s] to [%s]",
			prev.(Build).Status.String(), BuildCanceled.String())
	}
}
