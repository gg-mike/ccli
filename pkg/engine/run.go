package engine

import (
	"fmt"
	"time"

	"github.com/gg-mike/ccli/pkg/db"
	"github.com/gg-mike/ccli/pkg/model"
	"github.com/gg-mike/ccli/pkg/runner"
)

func (e *Engine) run(ctx *model.QueueContext, _runner *runner.Runner) error {
	if err := createEnvSteps(ctx); err != nil {
		e.logger.Error().Str("build_id", ctx.Build.ID()).Err(err).Msg("error during env steps creation")
		return err
	}

	failed := false

	for _, step := range ctx.Config.Steps {
		if err := runStep(ctx, _runner, step); err != nil {
			if err != ErrBuildCancelled {
				e.logger.Warn().Str("build_id", ctx.Build.ID()).Err(err).Msgf("error during step [%s]", step.Name)
				failed = true
			}
			break
		}
	}

	if err := runStep(ctx, _runner, model.PipelineConfigStep{Name: "Cleanup", Commands: ctx.Config.Cleanup}); err != nil {
		e.logger.Warn().Str("build_id", ctx.Build.ID()).Err(err).Msgf("error during cleanup")
	}

	if err := _runner.Shutdown(); err != nil {
		e.logger.Error().Str("build_id", ctx.Build.ID()).Err(err).Msg("error during Runner close")
		return err
	}

	if failed {
		return runner.ErrBuildFailed
	}
	return nil
}

func runStep(ctx *model.QueueContext, _runner *runner.Runner, step model.PipelineConfigStep) error {
	start := time.Now()

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

	_runner.OnCmd = onCmd(&buildStep)
	_runner.OnOut = onOut(&buildStep)

	fmt.Printf("\n### %s ###\n\n", step.Name)

	err := _runner.Run(step.Commands)

	buildStep.End()

	ctx.Build.Steps = append(ctx.Build.Steps, buildStep)

	if err := db.Get().Create(&buildStep).Error; err != nil {
		return err
	}
	return err
}

func onCmd(buildStep *model.BuildStep) func(cmd string, idx, total int) {
	return func(cmd string, idx, total int) {
		fmt.Printf("\033[32m[%d/%d] $ %s\033[0m\n", idx+1, total, cmd)
		buildStep.AppendLog(model.BuildLog{Command: cmd, Idx: idx + 1, Total: total, Output: ""})
	}
}

func onOut(buildStep *model.BuildStep) func(out string) {
	return func(out string) {
		fmt.Println(out)
		buildStep.AppendOutput(out)
	}
}
