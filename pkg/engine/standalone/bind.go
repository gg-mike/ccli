package standalone

import (
	"database/sql"
	"errors"

	"github.com/gg-mike/ccli/pkg/db"
	"github.com/gg-mike/ccli/pkg/model"
)

var (
	ErrUpdatingWorker = errors.New("error while updating worker")
	ErrUpdatingBuild  = errors.New("error while updating build")
)

func bind(queueElem *model.QueueElem, worker model.Worker) error {
	queueElem.Context.Worker = worker

	if db.Get().Model(&queueElem.Context.Worker).UpdateColumns(map[string]any{
		"active_builds": queueElem.Context.Worker.ActiveBuilds + 1,
		"status":        model.WorkerUsed}).Error != nil {
		return ErrUpdatingWorker
	}

	queueElem.Context.Build.AppendLog(model.BuildLog{Command: "[bind]", Output: "worker [" + queueElem.Context.Worker.Name + "] bound"})
	queueElem.Context.Build.WorkerName = sql.NullString{String: worker.Name, Valid: true}
	queueElem.Context.Build.Status = model.BuildRunning
	queueElem.Context.Build.End()

	if db.Get().UpdateColumns(queueElem.Context.Build).Error != nil {
		return ErrUpdatingBuild
	}

	return nil
}
