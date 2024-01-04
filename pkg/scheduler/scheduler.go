package scheduler

type IScheduler interface {
	Schedule(buildID string)
	Finished(buildID string)
	ChangeInWorkers()
}

var scheduler IScheduler

func Get() IScheduler {
	if scheduler == nil {
		panic("scheduler is not initialized")
	}
	return scheduler
}

func Init(s IScheduler) {
	if scheduler != nil {
		panic("scheduler is already initialized")
	}

	scheduler = s
}
