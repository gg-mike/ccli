package standalone

import (
	"github.com/gg-mike/ccli/pkg/db"
	"github.com/gg-mike/ccli/pkg/model"
	"gorm.io/gorm"
)

type Binder struct {
	onBind func(model.QueueContext)
}

func (b Binder) Bind() error {
	return db.Get().Transaction(func(tx *gorm.DB) error {
		var q []model.QueueElem
		if err := db.Get().Find(&q).Error; err != nil {
			return err
		}

		if len(q) == 0 {
			return nil
		}

		for _, elem := range q {
			var workers []model.Worker
			if err := tx.Model(&model.Worker{}).Find(&workers, &model.Worker{Status: model.WorkerIdle}).Error; err != nil {
				return err
			}

			worker, err := SelectWorker(elem.Context.Config, workers)

			if err == ErrNoAvailableWorker {
				return nil
			} else if err == ErrNoAvailableWorkerForConfiguration {
				continue
			}

			if err := bind(&elem, worker); err != nil {
				if err := db.Get().Model(&elem.Context.Build).UpdateColumn("status", model.BuildFailed).Error; err != nil {
					return err
				}
				return err
			} else {
				if err := db.Get().Delete(&model.QueueElem{ID: elem.ID}).Error; err != nil {
					return err
				}
				go b.onBind(elem.Context)
			}

		}
		return nil
	})
}

func (b Binder) Unbind(worker model.Worker) {

}

func (b *Binder) SetOnBind(callback func(model.QueueContext)) {
	b.onBind = callback
}
