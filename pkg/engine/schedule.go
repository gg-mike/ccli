package engine

import (
	"fmt"
	"strings"
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
			Str("duration", ctx.Build.Steps[0].Duration).Msg("build context creation succeeded")
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
		Str("duration", ctx.Build.Steps[0].Duration).Msg("build context creation failed")
	if err := db.Get().Create(&ctx.Build.Steps[0]).Error; err != nil {
		e.logger.Error().Str("build_id", buildID).Str("step", "context-create").Err(err).Msg("could not write build steps")
		return ctx, ErrBuildSave
	}
	return ctx, ErrBuildInitFailed
}

func newQueueContext(buildID string, start time.Time) (model.QueueContext, error) {
	ctx := model.QueueContext{}
	ctx.Build = model.BuildFromID(buildID)
	steps := []model.BuildStep{
		{
			Name:         "Queue context creation",
			BuildNumber:  ctx.Build.Number,
			PipelineName: ctx.Build.PipelineName,
			ProjectName:  ctx.Build.ProjectName,
			Start:        start,
			Logs:         []model.BuildLog{},
		},
	}

	if db.Get().Preload(clause.Associations).First(&ctx.Build).Error != nil {
		ctx.Build.Status = model.BuildFailed
		ctx.Build.Steps = steps
		ctx.Build.End()
		return ctx, ErrInvalidBuild
	}
	ctx.Build.Steps = steps
	ctx.Build.AppendLog(model.BuildLog{Command: "[build init]", Output: "success"})

	pipeline := model.Pipeline{Name: ctx.Build.PipelineName, ProjectName: ctx.Build.ProjectName}
	project := model.Project{Name: ctx.Build.ProjectName}
	if !initSingle(&ctx, &pipeline, "pipeline") {
		return ctx, ErrInvalidPipeline
	}
	ctx.Branch = pipeline.Branch
	ctx.Config = pipeline.Config

	if !initSingle(&ctx, &project, "project") {
		return ctx, ErrInvalidProject
	}
	ctx.Repo = project.Repo

	if !initMultiple(&ctx, &ctx.Secrets, "secrets", "project_name", "pipeline_name", "path") {
		return ctx, ErrInvalidSecrets
	}
	if !initMultiple(&ctx, &ctx.Variables, "variables", "project_name", "pipeline_name", "path", "value") {
		return ctx, ErrInvalidVariables
	}

	return ctx, nil
}

func initSingle[T any](ctx *model.QueueContext, m *T, elem string) bool {
	if db.Get().First(m).Error != nil {
		ctx.Build.Status = model.BuildFailed
		ctx.Build.End()
		return false
	}
	ctx.Build.AppendLog(model.BuildLog{Command: fmt.Sprintf("[%s init]", elem), Output: "success"})
	return true
}

func initMultiple[T any](ctx *model.QueueContext, multiple *[]T, elem string, fields ...string) bool {
	selector := fmt.Sprintf("key, %s", strings.Join(fields, ", "))
	agg := fmt.Sprintf("key, %s", strings.Join(getAgg(fields), ", "))

	subQuery := db.Get().Table(elem).
		Select(selector).
		Order("key, project_name, pipeline_name")

	output, ok := getOutput(db.Get().Table("(?) as sq", subQuery).
		Select(agg).
		Group("key").
		Scan(multiple).Error)

	ctx.Build.AppendLog(model.BuildLog{Command: fmt.Sprintf("[%s init]", elem), Output: output})
	if !ok {
		ctx.Build.Status = model.BuildFailed
		ctx.Build.End()
		return false
	}
	return true
}

func getAgg(values []string) []string {
	out := []string{}
	for _, value := range values {
		out = append(out, "(array_agg("+value+")::TEXT[])[1] as "+value)
	}
	return out
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
