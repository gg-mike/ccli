package engine

import (
	"github.com/gg-mike/ccli/pkg/db"
	"github.com/gg-mike/ccli/pkg/model"
)

func (e *Engine) finished(buildID string) error {
	build := model.BuildFromID(buildID)

	if err := db.Get().First(&build).Error; err != nil {
		return err
	}
	if !build.WorkerName.Valid {
		e.logger.Debug().Str("build_id", buildID).Msg("build wasn't bound to any worker")
		return nil
	}

	if err := e.binder.Unbind(build.WorkerName.String); err != nil {
		return err
	}

	go e.ChangeInWorkers()
	return nil
}
