package engine

import (
	"time"

	"github.com/gg-mike/ccli/pkg/db"
	"github.com/gg-mike/ccli/pkg/engine/common"
	"github.com/gg-mike/ccli/pkg/log"
	"github.com/gg-mike/ccli/pkg/model"
)

type EngineEvent int

const (
	EventSchedule EngineEvent = iota
	EventFinished
	EventAddToQueue
	EventChangeInWorkers
	EventShutdown
)

type EngineEventStatus int

const (
	EventReceived EngineEventStatus = iota
	EventProcessed
	EventComplete
	EventFailed
)

type Engine struct {
	newBuild        chan string
	finishedBuild   chan string
	addToQueue      chan model.QueueContext
	changeInWorkers chan any
	shutdown        chan any
	done            chan any

	logger log.Logger
	binder common.IBinder
}

func NewEngine(logger log.Logger, binder common.IBinder) *Engine {
	return &Engine{
		newBuild:        make(chan string),
		finishedBuild:   make(chan string),
		addToQueue:      make(chan model.QueueContext),
		changeInWorkers: make(chan any),
		shutdown:        make(chan any),
		done:            make(chan any),

		logger: logger.NewComponentLogger("engine"),
		binder: binder,
	}
}

func (e *Engine) Run() {
	e.logger.Info().Msg("starting engine")

	e.logger.Debug().Msg("binding any builds scheduled in previous run")

	e.binder.SetOnBind(e.execute)
	if err := e.binder.Bind(); err != nil {
		e.logger.Error().Err(err).Msg("bind ended with error")
	}

	run := true

	for run {
		select {
		case buildID := <-e.newBuild:
			e.logger.Debug().Str("event", EventSchedule.String()).Str("status", EventProcessed.String()).Str("build_id", buildID).Send()

			e.schedule(buildID)

			e.logger.Debug().Str("event", EventSchedule.String()).Str("status", EventComplete.String()).Str("build_id", buildID).Send()
		case buildID := <-e.finishedBuild:
			e.logger.Debug().Str("event", EventFinished.String()).Str("status", EventProcessed.String()).Str("build_id", buildID).Send()

			if err := e.finished(buildID); err != nil {
				e.logger.Error().Str("event", EventFinished.String()).Str("status", EventComplete.String()).Str("build_id", buildID).Err(err).Send()
			} else {
				e.logger.Debug().Str("event", EventFinished.String()).Str("status", EventComplete.String()).Str("build_id", buildID).Send()
			}
		case ctx := <-e.addToQueue:
			e.logger.Debug().Str("event", EventAddToQueue.String()).Str("status", EventProcessed.String()).Str("build_id", ctx.Build.ID()).Send()

			ctx.Build.Steps = append(ctx.Build.Steps, model.BuildStep{
				Name:         "Worker binding",
				BuildNumber:  ctx.Build.Number,
				PipelineName: ctx.Build.PipelineName,
				ProjectName:  ctx.Build.ProjectName,
				Start:        time.Now(),
				Logs:         []model.BuildLog{},
			})

			if err := db.Get().Create(&model.QueueElem{ID: ctx.Build.ID(), Context: ctx}).Error; err != nil {
				e.logger.Error().Str("event", EventAddToQueue.String()).Str("status", EventFailed.String()).Str("build_id", ctx.Build.ID()).Err(err).Send()
				continue
			}

			if err := e.binder.Bind(); err != nil {
				e.logger.Error().Str("event", EventAddToQueue.String()).Str("status", EventFailed.String()).Str("build_id", ctx.Build.ID()).Err(err).Send()
			}

			e.logger.Debug().Str("event", EventAddToQueue.String()).Str("status", EventComplete.String()).Str("build_id", ctx.Build.ID()).Send()
		case <-e.changeInWorkers:
			e.logger.Debug().Str("event", EventChangeInWorkers.String()).Str("status", EventProcessed.String()).Send()

			if err := e.binder.Bind(); err != nil {
				e.logger.Error().Str("event", EventChangeInWorkers.String()).Str("status", EventFailed.String()).Err(err).Send()
			} else {
				e.logger.Debug().Str("event", EventChangeInWorkers.String()).Str("status", EventComplete.String()).Send()
			}
		case <-e.shutdown:
			e.logger.Debug().Str("event", EventShutdown.String()).Str("status", EventProcessed.String()).Send()

			run = false
			e.done <- true

			e.logger.Debug().Str("event", EventShutdown.String()).Str("status", EventComplete.String()).Send()
		}
	}

	e.logger.Info().Msg("engine shutdown")
}

func (e *Engine) Schedule(buildID string) {
	e.logger.Debug().Str("event", EventSchedule.String()).Str("status", EventReceived.String()).Str("build_id", buildID).Send()
	e.newBuild <- buildID
}

func (e *Engine) Finished(buildID string) {
	e.logger.Debug().Str("event", EventFinished.String()).Str("status", EventReceived.String()).Str("build_id", buildID).Send()
	e.finishedBuild <- buildID
}

func (e *Engine) AddToQueue(ctx model.QueueContext) {
	e.logger.Debug().Str("event", EventAddToQueue.String()).Str("status", EventReceived.String()).Str("build_id", ctx.Build.ID()).Send()
	e.addToQueue <- ctx
}

func (e *Engine) ChangeInWorkers() {
	e.logger.Debug().Str("event", EventChangeInWorkers.String()).Str("status", EventReceived.String()).Send()
	e.changeInWorkers <- true
}

func (e *Engine) Shutdown() chan any {
	e.logger.Debug().Str("event", EventShutdown.String()).Str("status", EventReceived.String()).Send()
	go func() { e.shutdown <- true }()
	return e.done
}

func (e EngineEvent) String() string {
	switch e {
	case EventSchedule:
		return "schedule"
	case EventFinished:
		return "finished"
	case EventAddToQueue:
		return "add-to-queue"
	case EventChangeInWorkers:
		return "change-in-workers"
	case EventShutdown:
		return "shutdown"
	default:
		return "unknown"
	}
}

func (s EngineEventStatus) String() string {
	switch s {
	case EventReceived:
		return "received"
	case EventProcessed:
		return "processed"
	case EventComplete:
		return "complete"
	case EventFailed:
		return "failed"
	default:
		return "unknown"
	}
}
