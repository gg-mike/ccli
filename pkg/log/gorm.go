// Based on https://github.com/wei840222/gorm-zerolog/blob/main/logger.go
package log

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type GormLogger struct {
	SlowThreshold         time.Duration
	SourceField           string
	SkipErrRecordNotFound bool
	Logger                zerolog.Logger
}

func Gorm(logger Logger) *GormLogger {
	return &GormLogger{
		Logger:                logger.impl,
		SkipErrRecordNotFound: false,
	}
}

func (logger *GormLogger) LogMode(gormLogger.LogLevel) gormLogger.Interface {
	return logger
}

func (logger *GormLogger) Info(ctx context.Context, s string, args ...interface{}) {
	logger.Logger.Info().Str("module", "gorm").Msgf(s, args...)
}

func (logger *GormLogger) Warn(ctx context.Context, s string, args ...interface{}) {
	logger.Logger.Warn().Str("module", "gorm").Msgf(s, args...)
}

func (logger *GormLogger) Error(ctx context.Context, s string, args ...interface{}) {
	logger.Logger.Error().Str("module", "gorm").Msgf(s, args...)
}

func (logger *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, _ := fc()
	fields := map[string]interface{}{
		"sql":      strings.ReplaceAll(sql, "\"", ""),
		"duration": elapsed,
	}
	if logger.SourceField != "" {
		fields[logger.SourceField] = utils.FileWithLineNum()
	}
	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && logger.SkipErrRecordNotFound) {
		logger.Logger.Error().Str("module", "gorm").Err(err).Fields(fields).Msg("query error")
		return
	}

	if logger.SlowThreshold != 0 && elapsed > logger.SlowThreshold {
		logger.Logger.Warn().Str("module", "gorm").Fields(fields).Msg("slow query")
		return
	}

	logger.Logger.Debug().Str("module", "gorm").Fields(fields).Msg("query")
}
