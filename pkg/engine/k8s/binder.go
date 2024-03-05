package k8s

import (
	"strings"

	"github.com/gg-mike/ccli/pkg/db"
	"github.com/gg-mike/ccli/pkg/engine/common"
	"github.com/gg-mike/ccli/pkg/kubernetes"
	"github.com/gg-mike/ccli/pkg/log"
	"github.com/gg-mike/ccli/pkg/model"
	"github.com/gg-mike/ccli/pkg/runner"
	"gorm.io/gorm"
)

type Binder struct {
	onBind func(model.QueueContext, *runner.Runner)

	client    *kubernetes.Client
	namespace string
	logger    log.Logger
}

type Config struct {
	Mode      string
	Config    string
	Namespace string
}

func NewBinder(logger log.Logger, config Config) (*Binder, error) {
	var client *kubernetes.Client
	var err error
	if config.Mode == "outer" {
		if config.Config == "" {
			client, err = kubernetes.NewDefaultOuterClient()
		} else {
			client, err = kubernetes.NewOuterClient(config.Config)
		}
	} else {
		client, err = kubernetes.NewInnerClient()
	}
	if err != nil {
		return &Binder{}, err
	}

	return &Binder{
		client:    client,
		namespace: config.Namespace,
		logger:    logger.NewComponentLogger("binder"),
	}, nil
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
			podName := strings.ReplaceAll(elem.ID, "/", "-")
			_runner, err := b.client.NewRunner(b.namespace, podName, elem.Context.Config.Image, elem.Context.Config.Shell)
			if err != nil {
				if err := db.Get().Model(&elem.Context.Build).UpdateColumn("status", model.BuildFailed).Error; err != nil {
					return err
				}
				return err
			}
			b.logger.Debug().Str("step", "bind").Str("build", podName).Msg("worker pod created")

			elem.Context.Build.AppendLog(model.BuildLog{Command: "[bind]", Output: "worker pod created"})
			elem.Context.Build.Status = model.BuildRunning
			elem.Context.Build.End()

			if db.Get().UpdateColumns(elem.Context.Build).Error != nil {
				return common.ErrUpdatingBuild
			}

			if err := db.Get().Delete(&model.QueueElem{ID: elem.ID}).Error; err != nil {
				return err
			}

			go b.onBind(elem.Context, _runner)
		}
		return nil
	})
}

func (b Binder) Unbind(workerName string) error {
	return nil
}

func (b *Binder) SetOnBind(callback func(model.QueueContext, *runner.Runner)) {
	b.onBind = callback
}
