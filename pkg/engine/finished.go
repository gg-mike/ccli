package engine

import (
	"github.com/gg-mike/ccli/pkg/db"
	"github.com/gg-mike/ccli/pkg/model"
	"gorm.io/gorm"
)

func (e *Engine) finished(buildID string) {
	build := model.BuildFromID(buildID)

	if db.Get().Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&build).Error; err != nil {
			return err
		}
		if !build.WorkerName.Valid {
			e.logger.Debug().Str("build_id", buildID).Msg("build wasn't bound to any worker")
			return nil
		}
		e.logger.Debug().Str("worker_name", build.WorkerName.String).Msg("decrement active builds on bound worker")
		var worker model.Worker
		if err := tx.Where(&model.Worker{Name: build.WorkerName.String}).First(&worker).Error; err != nil {
			e.logger.Error().Str("worker_name", build.WorkerName.String).Err(err).Msg("could not decrement active builds counter")
			return err
		}
		if err := tx.Model(&worker).UpdateColumn("active_builds", worker.ActiveBuilds-1).Error; err != nil {
			e.logger.Error().Str("worker_name", build.WorkerName.String).Err(err).Msg("could not decrement active builds counter")
			return err
		}

		defer e.binder.Unbind(worker)

		return nil
	}) != nil {
		return
	}
	go e.ChangeInWorkers()
}
