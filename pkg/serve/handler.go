package serve

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gg-mike/ccli/pkg/db"
	"github.com/gg-mike/ccli/pkg/log"
	"github.com/gg-mike/ccli/pkg/vault"
	"github.com/gin-gonic/gin"
)

type Flags struct {
	Address    string
	DbUrl      string
	VaultUrl   string
	VaultToken string
	Scheduler  string
}

type Handler struct {
	flags  *Flags
	logger log.Logger
	srv    *http.Server
}

func NewHandler(logger log.Logger, f *Flags) *Handler {
	h := &Handler{
		flags:  f,
		logger: logger,
	}

	h.initServer()
	h.initDb()
	h.initVault()

	return h
}

func (h *Handler) Run() {
	h.logger.Info().Msg("starting server")
	go func() {
		if err := h.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.logger.Fatal().Err(err).Msg("server start ended with error")
		}
	}()
}

func (h *Handler) Shutdown() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	stop()

	println()

	h.logger.Info().Msg("shutting down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.srv.Shutdown(ctx); err != nil {
		h.logger.Fatal().Err(err).Msg("server shutdown with error")
	}

	h.logger.Info().Msg("server shutdown")
}

func (h *Handler) initServer() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(log.Gin())
	router.Use(gin.Recovery())
	h.srv = &http.Server{
		Addr:    h.flags.Address,
		Handler: router,
	}
}

func (h *Handler) initVault() {
	if err := vault.Init(h.flags.VaultUrl, h.flags.VaultToken); err != nil {
		h.logger.Fatal().Err(err).Msg("error while connecting to the vault")
	}
	h.logger.Info().Msg("successfully connected to the vault")
}

func (h *Handler) initDb() {
	if err := db.Init(h.flags.DbUrl, log.Gorm(h.logger)); err != nil {
		h.logger.Fatal().Err(err).Msg("error while connecting to the db")
	}
	h.logger.Info().Msg("successfully connected to the db")
}
