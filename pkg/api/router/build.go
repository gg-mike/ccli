package router

import (
	"errors"
	"strconv"

	"github.com/gg-mike/ccli/pkg/model"
	"github.com/gin-gonic/gin"
)

type BuildRouter = IRouter[model.Build, model.BuildShort, struct{}]

func InitBuildRouter(pipeline *gin.RouterGroup) {
	r := NewRouter[model.Build, model.BuildShort, struct{}](
		// FILTER
		func(ctx *gin.Context) map[string]any {
			filters := map[string]any{}
			for key := range ctx.Request.URL.Query() {
				switch key {
				case "status":
					filters["status in ?"] = ctx.QueryArray(key)
				case "worker_name":
					filters["worker_name = ?"] = ctx.Query(key)
				}
			}

			return filters
		},
		// GET SELECTOR
		func(params gin.Params) (model.Build, error) {
			projectName, ok := params.Get("project_name")
			if !ok {
				return model.Build{}, errors.New("missing param 'project_name'")
			}
			pipelineName, ok := params.Get("pipeline_name")
			if !ok {
				return model.Build{}, errors.New("missing param 'pipeline_name'")
			}
			buildNumber, ok := params.Get("build_number")
			if !ok {
				return model.Build{}, errors.New("missing param 'build_number'")
			}
			_buildNumber, err := strconv.Atoi(buildNumber)
			if err != nil {
				return model.Build{}, errors.New("error parsing 'build_number'")
			}
			return model.Build{Number: uint(_buildNumber), PipelineName: pipelineName, ProjectName: projectName}, nil
		},
		// GET PARENT
		func(params gin.Params) (model.Build, error) {
			projectName, ok := params.Get("project_name")
			if !ok {
				return model.Build{}, errors.New("missing param 'project_name'")
			}
			pipelineName, ok := params.Get("pipeline_name")
			if !ok {
				return model.Build{}, errors.New("missing param 'pipeline_name'")
			}
			return model.Build{PipelineName: pipelineName, ProjectName: projectName}, nil
		},
		// MERGE
		func(left model.Build, right struct{}) model.Build {
			return left
		},
	)

	_rg := pipeline.Group(":pipeline_name/builds")

	_rg.GET("", getManyBuilds(r))
	_rg.POST("", createBuild(r))
	_rg.GET(":build_number", getOneBuild(r))
	_rg.PUT(":build_number", updateBuild(r))
}

// @Summary  Get builds
// @ID       many-builds
// @Tags     builds
// @Produce  json
// @Param    project_name  path  string  true  "Project name"
// @Param    pipeline_name path  string  true  "Pipeline name"
// @Param    page          query int     false "Page number"
// @Param    size          query int     false "Page size"
// @Param    order         query string  false "Order by field"
// @Param    status        query []int   false "Build status (possible values)"
// @Param    worker_name   query string  false "Build worker name (exact)"
// @Success  200 {object} []model.BuildShort "List of builds"
// @Failure  400 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /projects/{project_name}/pipelines/{pipeline_name}/builds [get]
func getManyBuilds(r BuildRouter) gin.HandlerFunc {
	return r.GetMany
}

// @Summary  Schedule build
// @ID       create-build
// @Tags     builds
// @Accept   json
// @Param    project_name  path string true "Project name"
// @Param    pipeline_name path string true "Pipeline name"
// @Success  202 {string} Success message
// @Failure  400 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /projects/{project_name}/pipelines/{pipeline_name}/builds [post]
func createBuild(r BuildRouter) gin.HandlerFunc {
	return r.Create
}

// @Summary  Get the single build
// @ID       single-build
// @Tags     builds
// @Produce  json
// @Param    project_name  path string true "Project name"
// @Param    pipeline_name path string true "Pipeline name"
// @Param    build_number  path int    true "Build number"
// @Success  201 {object} model.Build "Requested build"
// @Failure  400 {string} Error in request
// @Failure  404 {string} No record found
// @Failure  500 {string} Database error
// @Router   /projects/{project_name}/pipelines/{pipeline_name}/builds/{build_number} [get]
func getOneBuild(r BuildRouter) gin.HandlerFunc {
	return r.GetOne
}

// @Summary  Cancel build
// @ID       update-build
// @Tags     builds
// @Accept   json
// @Param    project_name  path string true "Project name"
// @Param    pipeline_name path string true "Pipeline name"
// @Param    build_number  path int    true "Build number"
// @Success  200 {object} model.Build "Updated build"
// @Failure  400 {string} Error in request
// @Failure  404 {string} No record found
// @Failure  500 {string} Database error
// @Router   /projects/{project_name}/pipelines/{pipeline_name}/builds/{build_number} [put]
func updateBuild(r BuildRouter) gin.HandlerFunc {
	return r.Update
}
