package common

import (
	"github.com/gg-mike/ccli/pkg/model"
	"github.com/gg-mike/ccli/pkg/runner"
)

type IBinder interface {
	Bind() error
	Unbind(workerName string) error

	SetOnBind(callback func(model.QueueContext, *runner.Runner))
}
