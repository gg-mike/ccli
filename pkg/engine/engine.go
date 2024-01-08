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
			e.eventLog(EventSchedule, EventProcessed, buildID)

			e.schedule(buildID)

			e.eventLog(EventSchedule, EventComplete, buildID)
		case buildID := <-e.finishedBuild:
			e.eventLog(EventFinished, EventProcessed, buildID)

			if err := e.finished(buildID); err != nil {
				e.eventErr(EventFinished, EventComplete, err, buildID)
			} else {
				e.eventLog(EventFinished, EventComplete, buildID)
			}
		case ctx := <-e.addToQueue:
			e.eventLog(EventAddToQueue, EventProcessed, ctx.Build.ID())

			ctx.Build.Steps = append(ctx.Build.Steps, model.BuildStep{
				Name:         "Worker binding",
				BuildNumber:  ctx.Build.Number,
				PipelineName: ctx.Build.PipelineName,
				ProjectName:  ctx.Build.ProjectName,
				Start:        time.Now(),
				Logs:         []model.BuildLog{},
			})

			if err := db.Get().Create(&model.QueueElem{ID: ctx.Build.ID(), Context: ctx}).Error; err != nil {
				e.eventErr(EventAddToQueue, EventFailed, err, ctx.Build.ID())
				continue
			}

			if err := e.binder.Bind(); err != nil {
				e.eventErr(EventAddToQueue, EventFailed, err, ctx.Build.ID())
			}

			e.eventLog(EventAddToQueue, EventComplete, ctx.Build.ID())
		case <-e.changeInWorkers:
			e.eventLog(EventChangeInWorkers, EventProcessed)

			if err := e.binder.Bind(); err != nil {
				e.eventErr(EventChangeInWorkers, EventFailed, err)
			} else {
				e.eventLog(EventChangeInWorkers, EventComplete)
			}
		case <-e.shutdown:
			e.eventLog(EventShutdown, EventProcessed)
			run = false
			e.done <- true
			e.eventLog(EventShutdown, EventComplete)
		}
	}

	e.logger.Info().Msg("engine shutdown")
}

func (e *Engine) Schedule(buildID string) {
	e.eventLog(EventSchedule, EventReceived, buildID)
	e.newBuild <- buildID
}

func (e *Engine) Finished(buildID string) {
	e.eventLog(EventFinished, EventReceived, buildID)
	e.finishedBuild <- buildID
}

func (e *Engine) AddToQueue(ctx model.QueueContext) {
	e.eventLog(EventAddToQueue, EventReceived, ctx.Build.ID())
	e.addToQueue <- ctx
}

func (e *Engine) ChangeInWorkers() {
	e.eventLog(EventChangeInWorkers, EventReceived)
	e.changeInWorkers <- true
}

func (e *Engine) Shutdown() chan any {
	e.eventLog(EventShutdown, EventReceived)
	go func() { e.shutdown <- true }()
	return e.done
}

func (e Engine) eventLog(event EngineEvent, status EngineEventStatus, buildID ...string) {
	if len(buildID) > 0 {
		e.logger.Debug().Str("event", event.String()).Str("status", status.String()).Str("build_id", buildID[0]).Send()
	} else {
		e.logger.Debug().Str("event", event.String()).Str("status", status.String()).Send()
	}

}

func (e Engine) eventErr(event EngineEvent, status EngineEventStatus, err error, buildID ...string) {
	if len(buildID) > 0 {
		e.logger.Error().Str("event", event.String()).Str("status", status.String()).Str("build_id", buildID[0]).Err(err).Send()
	} else {
		e.logger.Error().Str("event", event.String()).Str("status", status.String()).Err(err).Send()
	}
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
