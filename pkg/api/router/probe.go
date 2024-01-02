package router

import (
	"github.com/gg-mike/ccli/pkg/api/handler"
	"github.com/gin-gonic/gin"
)

func InitProbeRouter(rg *gin.RouterGroup, state *handler.State) *gin.RouterGroup {
	_rg := rg.Group("/-")

	_rg.GET("/healthy", state.HealthyHandler())
	_rg.GET("/ready", state.ReadyHandler())

	return _rg
}
