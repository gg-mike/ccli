package model

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gg-mike/ccli/pkg/scheduler"
	"gorm.io/gorm"
)

const (
	BuildScheduled  = "scheduled"
	BuildRunning    = "running"
	BuildSuccessful = "successful"
	BuildFailed     = "failed"
	BuildCanceled   = "canceled"
)

type Build struct {
	Number       uint           `json:"number"          gorm:"primaryKey;uniqueIndex:idx_builds"`
	PipelineName string         `json:"pipeline_name"   gorm:"primaryKey;uniqueIndex:idx_builds"`
	ProjectName  string         `json:"project_name"    gorm:"primaryKey;uniqueIndex:idx_builds"`
	Status       string         `json:"status"          gorm:"default:scheduled"`
	Steps        []BuildStep    `json:"steps,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:BuildNumber,PipelineName,ProjectName"`
	WorkerName   sql.NullString `json:"worker_name"`
	CreatedAt    time.Time      `json:"created_at"      gorm:"default:now()"`
	UpdatedAt    time.Time      `json:"updated_at"      gorm:"default:now()"`
}

type BuildShort struct {
	Number     uint           `json:"number"`
	Status     string         `json:"status"`
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
	go scheduler.Get().Schedule(m.ID())
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
		go scheduler.Get().Finished(m.ID())
		return nil
	default:
		return fmt.Errorf("cannot change status of build from [%s] to [%s]",
			prev.(Build).Status, BuildCanceled)
	}
}

func (m Build) ID() string {
	return fmt.Sprintf("%s/%s/%d",
		m.ProjectName,
		m.PipelineName,
		m.Number,
	)
}

func BuildFromID(buildID string) Build {
	parts := strings.Split(buildID, "/")
	m := Build{}
	m.ProjectName = parts[0]
	m.PipelineName = parts[1]
	num, _ := strconv.Atoi(parts[2])
	m.Number = uint(num)
	return m
}

func (m *Build) AppendLog(log BuildLog) {
	m.Steps[len(m.Steps)-1].AppendLog(log)
}

func (m *Build) AppendOutput(output string) {
	m.Steps[len(m.Steps)-1].AppendOutput(output)
}

func (m *Build) End() {
	m.Steps[len(m.Steps)-1].End()
}
