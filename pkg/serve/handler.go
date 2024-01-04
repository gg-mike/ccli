package serve

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	docs "github.com/gg-mike/ccli/docs"
	"github.com/gg-mike/ccli/pkg/api/handler"
	"github.com/gg-mike/ccli/pkg/api/router"
	"github.com/gg-mike/ccli/pkg/db"
	"github.com/gg-mike/ccli/pkg/engine"
	"github.com/gg-mike/ccli/pkg/engine/standalone"
	"github.com/gg-mike/ccli/pkg/log"
	"github.com/gg-mike/ccli/pkg/scheduler"
	"github.com/gg-mike/ccli/pkg/vault"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	state  *handler.State
	engine *engine.Engine
}

func NewHandler(logger log.Logger, f *Flags) *Handler {
	h := &Handler{
		flags:  f,
		logger: logger,
		state:  handler.NewState(),
	}

	if h.flags.Scheduler == "standalone" {
		h.engine = engine.NewEngine(logger, &standalone.Binder{})
	} else {
		panic("not implemented")
	}

	h.state.Healthy()
	h.state.NotReady()

	h.initServer()
	h.initDb()
	h.initVault()
	h.initScheduler()

	return h
}

func (h *Handler) Run() {
	h.logger.Info().Msg("starting server")
	go func() {
		if err := h.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.logger.Fatal().Err(err).Msg("server start ended with error")
		}
	}()
	go h.engine.Run()

	h.state.Ready()

	h.logger.Info().
		Msgf("API documentation available at http://%s/api/docs/index.html", h.flags.Address)
}

func (h *Handler) Shutdown() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	stop()

	println()

	<-h.engine.Shutdown()

	h.logger.Info().Msg("shutting down gracefully, press Ctrl+C again to force")

	h.state.NotReady()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.srv.Shutdown(ctx); err != nil {
		h.logger.Fatal().Err(err).Msg("server shutdown with error")
	}

	h.logger.Info().Msg("server shutdown")
}

func (h *Handler) initServer() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(log.Gin(h.logger))
	r.Use(gin.Recovery())

	rg := r.Group("/api")
	router.InitProbeRouter(rg, h.state)
	router.InitWorkerRouter(rg)
	projectRg := router.InitProjectRouter(rg)
	pipelineRg := router.InitPipelineRouter(projectRg)
	router.InitBuildRouter(pipelineRg)
	router.InitSecretRouter(rg, projectRg, pipelineRg)
	router.InitVariableRouter(rg, projectRg, pipelineRg)
	router.InitQueueRouter(rg)

	docs.SwaggerInfo.Title = "ccli - CI/CD CLI Application"
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Description = "HTTP server for ccli developed using Go (Gin, Gorm)."
	rg.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	h.srv = &http.Server{
		Addr:    h.flags.Address,
		Handler: r,
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

func (h *Handler) initScheduler() {
	scheduler.Init(h.engine)
}
