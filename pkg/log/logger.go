package log

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger struct {
	impl      zerolog.Logger
	module    string
	component string
}

func NewLogger(module, component, level, dir string) Logger {
	if dir == "" {
		return setup(module, component, level)
	}
	return setupMultiOutput(module, component, level, dir)
}

func (logger Logger) NewComponentLogger(component string) Logger {
	logger.component = component
	return logger
}

func setup(module, component, level string) Logger {
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		r, _ := regexp.Compile(`[^\\/]+[\\/][^\\/]+$`)
		shortPath := r.FindString(file)
		if shortPath != "" {
			file = shortPath
		}
		file = strings.ReplaceAll(file, "\\", "/")
		return file + ":" + strconv.Itoa(line)
	}
	lvl, err := zerolog.ParseLevel(level)
	cw := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	logger := Logger{log.Logger.Level(lvl).Output(cw), module, component}
	if err != nil {
		logger.Warn().Err(err).Send()
	}
	return logger
}

func setupMultiOutput(module, component, level, dir string) Logger {
	os.MkdirAll(dir, os.ModePerm)

	logger := setup(module, component, level)

	f, err := os.Create(fmt.Sprintf("%s/%s-%s.log", dir, module, time.Now().Format("20060102-150405")))
	if err != nil {
		logger.Error().Err(fmt.Errorf("error creating log file (%w)", err)).Send()
		return logger
	}

	writers := io.MultiWriter(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}, f)

	return Logger{logger.impl.Output(writers), module, component}
}

func (logger Logger) Debug() *zerolog.Event {
	return logger.impl.Debug().Str("module", logger.module).Str("component", logger.component).Caller(1)
}

func (logger Logger) Info() *zerolog.Event {
	return logger.impl.Info().Str("module", logger.module).Str("component", logger.component).Caller(1)
}

func (logger Logger) Warn() *zerolog.Event {
	return logger.impl.Warn().Str("module", logger.module).Str("component", logger.component).Caller(1)
}

func (logger Logger) Error() *zerolog.Event {
	return logger.impl.Error().Str("module", logger.module).Str("component", logger.component).Caller(1)
}

func (logger Logger) Fatal() *zerolog.Event {
	return logger.impl.Fatal().Str("module", logger.module).Str("component", logger.component).Caller(1)
}

func (logger Logger) Panic() *zerolog.Event {
	return logger.impl.Panic().Str("module", logger.module).Str("component", logger.component).Caller(1)
}

func (logger Logger) Trace() *zerolog.Event {
	return logger.impl.Trace().Str("module", logger.module).Str("component", logger.component).Caller(1)
}
