package handler

import (
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

type State struct {
	healthy *atomic.Bool
	ready   *atomic.Bool
}

func NewState() *State {
	return &State{&atomic.Bool{}, &atomic.Bool{}}
}

// @Summary  Check liveness
// @ID       healthy
// @Tags     probes
// @Success  200 {string} OK
// @Failure  503 {string} NOT OK
// @Router   /-/healthy [get]
func (s *State) HealthyHandler() gin.HandlerFunc {
	return s.handler(s.isHealthy)
}

// @Summary  Check readiness
// @ID       ready
// @Tags     probes
// @Success  200 {string} OK
// @Failure  503 {string} NOT OK
// @Router   /-/ready [get]
func (s *State) ReadyHandler() gin.HandlerFunc {
	return s.handler(s.isReady)
}

type check func() bool

func (s *State) handler(c check) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !c() {
			ctx.String(http.StatusServiceUnavailable, "NOT OK")
			return
		}
		ctx.String(http.StatusOK, "OK")
	}
}

func (s *State) isReady() bool {
	ready := s.ready.Load()
	return ready
}

func (s *State) isHealthy() bool {
	healthy := s.healthy.Load()
	return healthy
}

func (s *State) Ready() {
	s.ready.Swap(true)
}

func (s *State) NotReady() {
	s.ready.Swap(false)
}

func (s *State) Healthy() {
	s.healthy.Swap(true)
}

func (s *State) NotHealthy() {
	s.healthy.Swap(false)
}
