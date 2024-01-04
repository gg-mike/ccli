package standalone

import "github.com/gg-mike/ccli/pkg/model"

type Binder struct {
	onBind func(model.QueueContext)
}

func (b Binder) Bind() {
}

func (b Binder) Unbind(worker model.Worker) {

}

func (b *Binder) SetOnBind(callback func(model.QueueContext)) {
	b.onBind = callback
}
