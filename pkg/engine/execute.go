package engine

import (
	"errors"
	"fmt"
	"time"

	"github.com/gg-mike/ccli/pkg/db"
	"github.com/gg-mike/ccli/pkg/model"
	"github.com/gg-mike/ccli/pkg/ssh"
)

var ErrBuildCancelled = errors.New("build cancelled")
var ErrBuildFailed = errors.New("build failed")

func (e *Engine) execute(ctx model.QueueContext) {
	id := ctx.Build.ID()

	e.logger.Debug().Str("build_id", id).Str("step", "execute").Msg("build execution started")
	var err error
	build := model.BuildFromID(ctx.Build.ID())
	if err = e._execute(&ctx); err != nil && err != ErrBuildCancelled {
		go e.Finished(id)
		e.logger.Warn().Str("build_id", id).Str("step", "execute").Err(err).Msg("build execution ended with error")
		if err := db.Get().Model(&build).UpdateColumn("status", model.BuildFailed).Error; err != nil {
			e.logger.Error().Str("build_id", id).Str("step", "execute").Err(err).Msg("could not update build")
			return
		}
		return
	}

	go e.Finished(id)

	e.logger.Debug().Str("build_id", id).Str("step", "execute").Str("status", ctx.Build.Status).Msg("build execution ended")

	if err == ErrBuildCancelled {
		return
	}
	if err := db.Get().Model(&build).UpdateColumn("status", model.BuildSuccessful).Error; err != nil {
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

	failed := false

	for _, step := range ctx.Config.Steps {
		if err := e.runStep(ctx, &connCtx, step); err != nil {
			if err != ErrBuildCancelled {
				e.logger.Warn().Str("build_id", ctx.Build.ID()).Err(err).Msgf("error during step [%s]", step.Name)
				failed = true
			}
			break
		}
	}

	if e.runStep(ctx, &connCtx, model.PipelineConfigStep{Name: "Cleanup", Commands: ctx.Config.Cleanup}) != nil {
		e.logger.Warn().Str("build_id", ctx.Build.ID()).Err(err).Msgf("error during cleanup")
	}

	if err := connCtx.Close(); err != nil {
		e.logger.Error().Str("build_id", ctx.Build.ID()).Err(err).Msg("error during SSH close")
		return err
	}

	if failed {
		return ErrBuildFailed
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

	buildStep := model.BuildStep{
		Name:         step.Name,
		BuildNumber:  ctx.Build.Number,
		PipelineName: ctx.Build.PipelineName,
		ProjectName:  ctx.Build.ProjectName,
		Start:        start,
		Logs:         []model.BuildLog{},
	}

	go connCtx.Run(step.Commands)

	<-connCtx.OutChan

	fmt.Printf("\n### %s ###\n\n", step.Name)

	var err error
	for running {
		select {
		case cmd := <-connCtx.CmdChan:
			fmt.Printf("\033[32m[%d/%d] $ %s\033[0m\n", cmd.Idx+1, cmd.Total, cmd.Command)
			buildStep.AppendLog(cmd)
		case out := <-connCtx.OutChan:
			fmt.Println(out)
			buildStep.AppendOutput(out)
		case err = <-connCtx.ErrChan:
			if err != nil {
				fmt.Printf("\033[1;31m%s\033[00m\n", err.Error())
				buildStep.AppendOutput(err.Error())
			}
			running = false
		}
	}

	buildStep.End()

	ctx.Build.Steps = append(ctx.Build.Steps, buildStep)

	if err := db.Get().Create(&buildStep).Error; err != nil {
		return err
	}
	return err
}

func parseCmd(cmd string, idx, total int) model.BuildLog {
	return model.BuildLog{Command: cmd, Idx: idx, Total: total, Output: ""}
}
