package engine

import (
	"errors"

	"github.com/gg-mike/ccli/pkg/db"
	"github.com/gg-mike/ccli/pkg/model"
	"github.com/gg-mike/ccli/pkg/runner"
)

var ErrBuildCancelled = errors.New("build cancelled")

func (e *Engine) execute(ctx model.QueueContext, _runner *runner.Runner) {
	build := model.BuildFromID(ctx.Build.ID())
	e.logger.Debug().Str("build_id", build.ID()).Str("step", "execute").Msg("build execution started")

	var err error
	if err = e.run(&ctx, _runner); err != nil && err != ErrBuildCancelled {
		go e.Finished(build.ID())
		e.logger.Warn().Str("build_id", build.ID()).Str("step", "execute").Err(err).Msg("build execution ended with error")
		if err := db.Get().Model(&build).UpdateColumn("status", model.BuildFailed).Error; err != nil {
			e.logger.Error().Str("build_id", build.ID()).Str("step", "execute").Err(err).Msg("could not update build")
			return
		}
		return
	}

	go e.Finished(build.ID())

	e.logger.Debug().Str("build_id", build.ID()).Str("step", "execute").Str("status", ctx.Build.Status).Msg("build execution ended")

	if err == ErrBuildCancelled {
		return
	}

	if err := db.Get().Model(&build).UpdateColumn("status", model.BuildSuccessful).Error; err != nil {
		e.logger.Error().Str("build_id", build.ID()).Str("step", "execute").Err(err).Msg("could not update build")
		return
	}
}
