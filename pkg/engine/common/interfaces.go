package common

import "github.com/gg-mike/ccli/pkg/model"

type IBinder interface {
	Bind()
	Unbind(worker model.Worker)

	SetOnBind(callback func(model.QueueContext))
}
