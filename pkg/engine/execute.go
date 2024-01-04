package engine

import (
	"errors"
	"time"

	"github.com/gg-mike/ccli/pkg/db"
	"github.com/gg-mike/ccli/pkg/model"
	"github.com/gg-mike/ccli/pkg/ssh"
)

var ErrBuildCancelled = errors.New("build cancelled")

func (e *Engine) execute(ctx model.QueueContext) {
	id := ctx.Build.ID()

	e.logger.Debug().Str("build_id", id).Str("step", "execute").Msg("build execution started")
	var err error
	if err = e._execute(&ctx); err != nil && err != ErrBuildCancelled {
		go e.Finished(id)
		e.logger.Warn().Str("build_id", id).Str("step", "execute").Err(err).Msg("build execution ended with error")
		if err := db.Get().Model(&ctx.Build).UpdateColumn("status", model.BuildFailed).Error; err != nil {
			e.logger.Error().Str("build_id", id).Str("step", "execute").Err(err).Msg("could not update build")
			return
		}
		return
	}

	go e.Finished(id)

	e.logger.Debug().Str("build_id", id).Str("step", "execute").Str("status", ctx.Build.Status.String()).Msg("build execution ended")

	if err == ErrBuildCancelled {
		return
	}
	if err := db.Get().Model(&ctx.Build).UpdateColumn("status", model.BuildSuccessful).Error; err != nil {
		e.logger.Error().Str("build_id", id).Str("step", "execute").Err(err).Msg("could not update build")
		return
	}
	e.logger.Debug().Str("build_id", id).Str("step", "execute").Msg("saved update to build")
}

func (e *Engine) _execute(ctx *model.QueueContext) error {
	pk, err := ctx.Worker.PK()
	if err != nil {
		e.logger.Error().Str("build_id", ctx.Build.ID()).Err(err).Msg("error during worker PK extraction")
		return err
	}

	if err := createEnvSteps(ctx); err != nil {
		e.logger.Error().Str("build_id", ctx.Build.ID()).Err(err).Msg("error during env steps creation")
		return err
	}

	connCtx, err := ssh.Init[model.BuildLog](ctx.Worker.Username, ctx.Worker.Address, pk)
	if err != nil {
		e.logger.Error().Str("build_id", ctx.Build.ID()).Err(err).Msg("error during SSH init")
		return err
	}
	connCtx.ParseCmd = parseCmd

	for _, step := range ctx.Pipeline.Config.Steps {
		if err := e.runStep(ctx, &connCtx, step); err != nil {
			if err != ErrBuildCancelled {
				e.logger.Error().Str("build_id", ctx.Build.ID()).Err(err).Msgf("error during step [%s]", step.Name)
			}
			break
		}
	}

	if e.runStep(ctx, &connCtx, model.PipelineConfigStep{Name: "Cleanup", Commands: ctx.Pipeline.Config.Cleanup}) != nil {
		e.logger.Error().Str("build_id", ctx.Build.ID()).Err(err).Msgf("error during cleanup")
	}

	if err := connCtx.Close(); err != nil {
		e.logger.Error().Str("build_id", ctx.Build.ID()).Err(err).Msg("error during SSH close")
		return err
	}

	return nil
}

func (e Engine) runStep(ctx *model.QueueContext, connCtx *ssh.Context[model.BuildLog], step model.PipelineConfigStep) error {
	e.logger.Debug().Str("build_id", ctx.Build.ID()).Str("event", "begin-step").Str("step", step.Name).Send()
	start := time.Now()

	running := true
	if err := db.Get().First(&ctx.Build).Error; err != nil {
		return err
	}
	if ctx.Build.Status == model.BuildCanceled {
		return ErrBuildCancelled
	}

	ctx.Build.Steps = append(ctx.Build.Steps, model.BuildStep{
		Name:         step.Name,
		BuildNumber:  ctx.Build.Number,
		PipelineName: ctx.Build.PipelineName,
		ProjectName:  ctx.Build.ProjectName,
		Start:        start,
		Logs:         []model.BuildLog{},
	})

	go connCtx.Run(step.Commands)

	var err error
	for running {
		select {
		case cmd := <-connCtx.CmdChan:
			e.logger.Debug().Str("build_id", ctx.Build.ID()).Str("event", "cmd").
				Int("idx", cmd.Idx).Int("total", cmd.Total).Str("cmd", cmd.Command).Send()

			ctx.Build.AppendLog(cmd)
		case out := <-connCtx.OutChan:
			e.logger.Debug().Str("build_id", ctx.Build.ID()).Str("event", "out").
				Str("out", out).Send()

			ctx.Build.AppendOutput(out)
		case err = <-connCtx.ErrChan:
			e.logger.Debug().Str("build_id", ctx.Build.ID()).Str("event", "err").
				Err(err).Send()

			ctx.Build.AppendOutput(err.Error())
			running = false
		}
	}

	ctx.Build.End()
	if err := db.Get().Create(&ctx.Build.Steps[len(ctx.Build.Steps)-1]).Error; err != nil {
		return err
	}
	if err != nil {
		if err := db.Get().Model(&ctx.Build).UpdateColumn("status", model.BuildFailed).Error; err != nil {
			return err
		}
		return err
	}
	return nil
}

func parseCmd(cmd string, idx, total int) model.BuildLog {
	return model.BuildLog{Command: cmd, Idx: idx, Total: total, Output: ""}
}
