package standalone

import (
	"database/sql"

	"github.com/gg-mike/ccli/pkg/db"
	"github.com/gg-mike/ccli/pkg/docker"
	"github.com/gg-mike/ccli/pkg/log"
	"github.com/gg-mike/ccli/pkg/model"
	"github.com/gg-mike/ccli/pkg/runner"
	"github.com/gg-mike/ccli/pkg/ssh"
	"gorm.io/gorm"
)

type Binder struct {
	onBind func(model.QueueContext, *runner.Runner)

	logger log.Logger
}

func NewBinder(logger log.Logger) *Binder {
	return &Binder{
		logger: logger.NewComponentLogger("binder"),
	}
}

func (b Binder) Bind() error {
	return db.Get().Transaction(func(tx *gorm.DB) error {
		var q []model.QueueElem
		if err := db.Get().Find(&q).Error; err != nil {
			return err
		}

		if len(q) == 0 {
			b.logger.Debug().Str("step", "bind").Msg("no build in queue")
			return nil
		}

		for _, elem := range q {
			var workers []model.Worker
			if err := tx.Model(&model.Worker{}).Where("status <> 'unreachable'").Find(&workers).Error; err != nil {
				return err
			}

			worker, err := SelectWorker(elem.Context.Config, workers)

			if err == ErrNoAvailableWorker {
				return nil
			} else if err == ErrNoAvailableWorkerForConfiguration {
				continue
			}

			b.logger.Debug().Str("step", "bind").Str("build", elem.ID).Str("worker", worker.Name).Msg("worker selected")

			if db.Get().Model(&worker).UpdateColumns(map[string]any{
				"active_builds": worker.ActiveBuilds + 1,
				"status":        model.WorkerUsed}).Error != nil {
				return ErrUpdatingWorker
			}

			elem.Context.Build.AppendLog(model.BuildLog{Command: "[bind]", Output: "worker [" + worker.Name + "] bound"})
			elem.Context.Build.WorkerName = sql.NullString{String: worker.Name, Valid: true}
			elem.Context.Build.Status = model.BuildRunning
			elem.Context.Build.End()

			if db.Get().UpdateColumns(elem.Context.Build).Error != nil {
				return ErrUpdatingBuild
			}

			_runner, err := getRunner(worker.IsStatic)(&elem, worker)
			if err != nil {
				if err := db.Get().Model(&elem.Context.Build).UpdateColumn("status", model.BuildFailed).Error; err != nil {
					return err
				}
				return err
			}

			if err := db.Get().Delete(&model.QueueElem{ID: elem.ID}).Error; err != nil {
				return err
			}

			b.logger.Debug().Str("step", "bind").Str("build", elem.ID).Str("worker", worker.Name).Msg("worker bound")
			go b.onBind(elem.Context, _runner)
		}
		return nil
	})
}

func (b Binder) Unbind(workerName string) error {
	return db.Get().Transaction(func(tx *gorm.DB) error {
		b.logger.Debug().Str("name", workerName).Msg("decrement active builds on bound worker")
		var worker model.Worker
		if err := tx.Where(&model.Worker{Name: workerName}).First(&worker).Error; err != nil {
			b.logger.Error().Str("name", workerName).Err(err).Msg("could not decrement active builds counter")
			return err
		}
		if err := tx.Model(&worker).UpdateColumn("active_builds", worker.ActiveBuilds-1).Error; err != nil {
			b.logger.Error().Str("name", workerName).Err(err).Msg("could not decrement active builds counter")
			return err
		}
		if worker.ActiveBuilds == 0 {
			if err := tx.Model(&worker).UpdateColumn("status", model.WorkerIdle).Error; err != nil {
				b.logger.Error().Str("name", workerName).Err(err).Msg("could not change status")
				return err
			}
		}
		return nil
	})
}

func (b *Binder) SetOnBind(callback func(model.QueueContext, *runner.Runner)) {
	b.onBind = callback
}

func getRunner(isStatic bool) func(*model.QueueElem, model.Worker) (*runner.Runner, error) {
	if isStatic {
		return func(_ *model.QueueElem, w model.Worker) (*runner.Runner, error) {
			pk, err := w.PK()
			if err != nil {
				return &runner.Runner{}, err
			}

			return ssh.NewRunner(w.Username, w.Address, pk)
		}
	}
	return func(qe *model.QueueElem, w model.Worker) (*runner.Runner, error) {
		return docker.NewRunner(w.Address, qe.Context.Config.Image)
	}
}
