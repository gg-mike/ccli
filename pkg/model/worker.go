package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/gg-mike/ccli/pkg/docker"
	"github.com/gg-mike/ccli/pkg/scheduler"
	"github.com/gg-mike/ccli/pkg/ssh"
	"github.com/gg-mike/ccli/pkg/vault"
	"gorm.io/gorm"
)

const (
	WorkerIdle        = "idle"
	WorkerUsed        = "used"
	WorkerUnreachable = "unreachable"
)

const (
	WorkerMin      = "min"
	WorkerBalanced = "balance"
	WorkerMax      = "max"
)

type Worker struct {
	Name         string    `json:"name"             gorm:"primaryKey"`
	Address      string    `json:"address"          gorm:"not null"`
	System       string    `json:"system"           gorm:"not null"`
	Username     string    `json:"username"         gorm:"not null"`
	IsStatic     bool      `json:"is_static"        gorm:"not null"`
	Status       string    `json:"status"           gorm:"default:idle"`
	Strategy     string    `json:"strategy"         gorm:"default:balance"`
	ActiveBuilds int       `json:"active_builds"    gorm:"default:0"`
	Capacity     int       `json:"capacity"         gorm:"default:0"`
	Builds       []Build   `json:"builds,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	CreatedAt    time.Time `json:"created_at"       gorm:"default:now()"`
	UpdatedAt    time.Time `json:"updated_at"       gorm:"default:now()"`
}

type WorkerShort struct {
	Name         string    `json:"name"`
	Address      string    `json:"address"`
	System       string    `json:"system"`
	Username     string    `json:"username"`
	IsStatic     bool      `json:"is_static"`
	Status       string    `json:"status"`
	Strategy     string    `json:"strategy"`
	ActiveBuilds int       `json:"active_builds"`
	Capacity     int       `json:"capacity"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type WorkerInput struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	System     string `json:"system"`
	IsStatic   bool   `json:"is_static"`
	Strategy   string `json:"strategy"`
	Username   string `json:"username"`
	PrivateKey string `json:"private_key"`
	Capacity   int    `json:"capacity"`
}

func (m *Worker) BeforeCreate(tx *gorm.DB) error {
	if m.IsStatic {
		privateKey, ok := getPK(tx)
		if !ok {
			return errors.New("no private_key field given in instance")
		}
		if !testConnection(*m, privateKey) {
			m.Status = WorkerUnreachable
		}
	} else if err := docker.NewClient(m.Address); err != nil {
		return err
	}
	return nil
}

func (m *Worker) AfterCreate(tx *gorm.DB) error {
	privateKey, _ := getPK(tx)
	return vault.SetStr(m.Name, privateKey)
}

func (m *Worker) AfterSave(tx *gorm.DB) error {
	go scheduler.Get().ChangeInWorkers()
	return nil
}

func (m *Worker) BeforeUpdate(tx *gorm.DB) error {
	if _, ok := getPK(tx); !ok {
		return nil
	}
	return vault.Del(m.Name)
}

func (m *Worker) AfterUpdate(tx *gorm.DB) error {
	prev, ok := tx.InstanceGet("prev")
	if !ok {
		return errors.New("prev worker not given")
	}
	if m.IsStatic {
		privateKey, ok := getPK(tx)
		if !ok {
			pKey, err := vault.GetStr(m.Name)
			if err != nil {
				return fmt.Errorf("error during retrieving private key: %v", err)
			}
			privateKey = pKey
		}
		var status string
		if !testConnection(*m, privateKey) {
			status = WorkerUnreachable
		} else if prev.(Worker).Status != WorkerUnreachable {
			status = prev.(Worker).Status
		} else {
			status = WorkerIdle
		}
		if err := tx.Model(&m).UpdateColumn("status", status).Error; err != nil {
			return err
		}
		if !ok {
			return nil
		}
		return vault.SetStr(m.Name, privateKey)
	} else {
		if err := docker.DeleteClient(prev.(Worker).Address); err != nil {
			return err
		}
		if err := docker.NewClient(m.Address); err != nil {
			return err
		}
	}
	return nil
}

func (m *Worker) BeforeDelete(tx *gorm.DB) error {
	for _, build := range m.Builds {
		if build.Status == BuildRunning {
			return errors.New("cannot delete worker with running builds")
		}
	}
	return nil
}

func (m *Worker) AfterDelete(tx *gorm.DB) error {
	if err := docker.DeleteClient(m.Address); err != nil {
		return err
	}
	return vault.Del(m.Name)
}

func testConnection(worker Worker, privateKey string) bool {
	return ssh.CheckConnection(worker.Username, worker.Address, privateKey) == nil
}

func getPK(tx *gorm.DB) (string, bool) {
	input, ok := tx.InstanceGet("input")
	if !ok {
		return "", false
	}
	return input.(WorkerInput).PrivateKey, ok
}

func (m Worker) PK() (string, error) {
	if m.IsStatic {
		return vault.GetStr(m.Name)
	}
	return "", errors.New("docker host worker does not have a private key")
}
