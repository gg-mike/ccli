package migrate

import (
	"github.com/gg-mike/ccli/pkg/db"
	"github.com/gg-mike/ccli/pkg/log"
	"github.com/gg-mike/ccli/pkg/model"
)

type Flags struct {
	DbUrl     string
	Scheduler string
}

type Handler struct {
	flags  *Flags
	logger log.Logger
}

func NewHandler(logger log.Logger, f *Flags) *Handler {
	h := &Handler{
		flags:  f,
		logger: logger,
	}

	h.initDb()

	return h
}

func (h *Handler) Run() error {
	if h.flags.Scheduler == "standalone" {
		return db.Get().AutoMigrate(
			&model.Worker{},
			&model.Project{},
			&model.Pipeline{},
			&model.Build{},
			&model.BuildStep{},
			&model.Secret{},
			&model.Variable{},
			&model.QueueElem{},
		)
	} else {
		return db.Get().AutoMigrate(
			&model.Project{},
			&model.Pipeline{},
			&model.Build{},
			&model.BuildStep{},
			&model.Secret{},
			&model.Variable{},
			&model.QueueElem{},
		)
	}
}

func (h *Handler) initDb() {
	if err := db.Init(h.flags.DbUrl, log.Gorm(h.logger)); err != nil {
		h.logger.Fatal().Err(err).Msg("error while connecting to the db")
	}
	h.logger.Info().Msg("successfully connected to the db")
}
