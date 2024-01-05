package common

import "github.com/gg-mike/ccli/pkg/model"

type IBinder interface {
	Bind() error
	Unbind(worker model.Worker)

	SetOnBind(callback func(model.QueueContext))
}
