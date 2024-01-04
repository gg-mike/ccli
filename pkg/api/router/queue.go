package router

import (
	"errors"
	"fmt"

	"github.com/gg-mike/ccli/pkg/model"
	"github.com/gin-gonic/gin"
)

type QueueRouter = IRouter[model.QueueElem, model.QueueElemShort, struct{}]

func InitQueueRouter(base *gin.RouterGroup) {
	r := NewRouter[model.QueueElem, model.QueueElemShort, struct{}](
		// FILTER
		func(ctx *gin.Context) map[string]any {
			return map[string]any{}
		},
		// GET SELECTOR
		func(params gin.Params) (model.QueueElem, error) {
			projectName, ok := params.Get("project_name")
			if !ok {
				return model.QueueElem{}, errors.New("missing param 'project_name'")
			}
			pipelineName, ok := params.Get("pipeline_name")
			if !ok {
				return model.QueueElem{}, errors.New("missing param 'pipeline_name'")
			}
			buildNumber, ok := params.Get("build_number")
			if !ok {
				return model.QueueElem{}, errors.New("missing param 'build_number'")
			}
			return model.QueueElem{ID: fmt.Sprintf("%s/%s/%s", projectName, pipelineName, buildNumber)}, nil
		},
		// GET PARENT
		func(params gin.Params) (model.QueueElem, error) {
			return model.QueueElem{}, nil
		},
		// MERGE
		func(left model.QueueElem, right struct{}) model.QueueElem {
			return left
		},
	)

	_rg := base.Group("/queue")

	_rg.GET("", getQueue(r))
	_rg.GET(":project_name/:pipeline_name/:build_number", getQueueElem(r))
	_rg.DELETE(":project_name/:pipeline_name/:build_number", deleteQueueElem(r))
}

// @Summary  Get queue
// @ID       queue
// @Tags     queue
// @Produce  json
// @Param    page          query int     false "Page number"
// @Param    size          query int     false "Page size"
// @Param    order         query string  false "Order by field"
// @Success  200 {object} []model.QueueElemShort "Queue"
// @Failure  400 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /queue [get]
func getQueue(r QueueRouter) gin.HandlerFunc {
	return r.GetMany
}

// @Summary  Get the single queue elem
// @ID       single-queue-elem
// @Tags     queue
// @Produce  json
// @Param    project_name  path string true "Project name"
// @Param    pipeline_name path string true "Pipeline name"
// @Param    build_number  path int    true "Build number"
// @Success  201 {object} model.QueueElem "Requested queue elem"
// @Failure  400 {string} Error in request
// @Failure  404 {string} No record found
// @Failure  500 {string} Database error
// @Router   /queue/{project_name}/{pipeline_name}/{build_number} [get]
func getQueueElem(r QueueRouter) gin.HandlerFunc {
	return r.GetOne
}

// @Summary  Delete queue elem
// @ID       delete-queue-elem
// @Tags     queue
// @Param    project_name  path string true "Project name"
// @Param    pipeline_name path string true "Pipeline name"
// @Param    build_number  path int    true "Build number"
// @Success  200 {string} Success message
// @Failure  404 {string} Error in request
// @Failure  500 {string} Database error
// @Router    /queue/{project_name}/{pipeline_name}/{build_number} [delete]
func deleteQueueElem(r QueueRouter) gin.HandlerFunc {
	return r.Delete
}
