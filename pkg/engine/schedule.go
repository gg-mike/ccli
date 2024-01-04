package engine

import (
	"fmt"
	"time"

	"github.com/gg-mike/ccli/pkg/db"
	"github.com/gg-mike/ccli/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (e *Engine) schedule(buildID string) {
	ctx, err := e.buildContext(buildID)
	if err != nil {
		if err != ErrInvalidBuild {
			e.logger.Fatal().Str("build_id", buildID).Str("step", "context-create").Err(err).Msg("fatal error during build context creation")
		}
		return
	}

	go e.AddToQueue(ctx)
}

func (e *Engine) buildContext(buildID string) (model.QueueContext, error) {
	start := time.Now()
	e.logger.Debug().Str("build_id", buildID).Str("step", "context-create").Msg("build context creation started")

	ctx, err := newQueueContext(buildID, start)

	if err == nil {
		ctx.Build.End()

		e.logger.Debug().Str("build_id", buildID).Str("step", "context-create").
			Str("duration", ctx.Build.Steps[len(ctx.Build.Steps)-1].Duration.String()).Msg("build context creation succeeded")
		if err := db.Get().Create(&ctx.Build.Steps[0]).Error; err != nil {
			e.logger.Error().Str("build_id", buildID).Str("step", "context-create").Err(err).Msg("could not write build steps")
			return ctx, ErrBuildSave
		}
		return ctx, nil
	}

	if err == ErrInvalidBuild {
		e.logger.Warn().Str("build_id", buildID).Str("step", "context-create").Err(err).Msg("build not available - rescheduling")
		go func() {
			time.Sleep(time.Second)
			go e.Schedule(buildID)
		}()
		return ctx, ErrInvalidBuild
	}

	ctx.Build.Status = model.BuildFailed
	ctx.Build.End()

	e.logger.Debug().Str("build_id", buildID).Str("step", "context-create").Err(err).
		Str("duration", ctx.Build.Steps[len(ctx.Build.Steps)-1].Duration.String()).Msg("build context creation failed")
	if err := db.Get().Create(&ctx.Build.Steps[0]).Error; err != nil {
		e.logger.Error().Str("build_id", buildID).Str("step", "context-create").Err(err).Msg("could not write build steps")
		return ctx, ErrBuildSave
	}
	return ctx, ErrBuildInitFailed
}

func newQueueContext(buildID string, start time.Time) (model.QueueContext, error) {
	ctx := model.QueueContext{}
	ctx.Build = model.BuildFromID(buildID)
	ctx.Pipeline.Name = ctx.Build.PipelineName
	ctx.Pipeline.ProjectName = ctx.Build.ProjectName
	ctx.Project.Name = ctx.Build.ProjectName
	ctx.Build.Steps = []model.BuildStep{
		{
			Name:         "Queue context creation",
			BuildNumber:  ctx.Build.Number,
			PipelineName: ctx.Build.PipelineName,
			ProjectName:  ctx.Build.ProjectName,
			Start:        start,
			Logs:         []model.BuildLog{},
		},
	}
	fmt.Printf("newQueueContext: %+v\n", ctx.Build.Steps)

	if !initSingle(&ctx, &ctx.Build, "build") {
		return ctx, ErrInvalidBuild
	}

	if !initSingle(&ctx, &ctx.Pipeline, "pipeline") {
		return ctx, ErrInvalidPipeline
	}

	if !initSingle(&ctx, &ctx.Project, "project") {
		return ctx, ErrInvalidProject
	}

	if !initMultiple(&ctx, &ctx.GlobalSecrets, "secrets") {
		return ctx, ErrInvalidSecrets
	}

	if !initMultiple(&ctx, &ctx.GlobalVariables, "variables") {
		return ctx, ErrInvalidVariables
	}

	return ctx, nil
}

func initSingle[T any](ctx *model.QueueContext, single *T, elem string) bool {
	steps := ctx.Build.Steps
	if db.Get().Preload(clause.Associations).First(single).Error != nil {
		ctx.Build.Status = model.BuildFailed
		ctx.Build.Steps = steps
		ctx.Build.End()
		return false
	}
	ctx.Build.Steps = steps
	ctx.Build.AppendLog(model.BuildLog{Command: fmt.Sprintf("[%s init]", elem), Output: "success"})
	return true
}

func initMultiple[T any](ctx *model.QueueContext, multiple *[]T, elem string) bool {
	output, ok := getOutput(db.Get().Find(multiple, "project_name IS NULL AND pipeline_name IS NULL").Error)
	ctx.Build.AppendLog(model.BuildLog{Command: fmt.Sprintf("[global %s init]", elem), Output: output})
	if !ok {
		ctx.Build.Status = model.BuildFailed
		ctx.Build.End()
		return false
	}
	return true
}

func getOutput(err error) (string, bool) {
	switch err {
	case nil:
		return "success", true
	case gorm.ErrRecordNotFound:
		return "non found", true
	default:
		return "failed: " + err.Error(), false
	}
}
