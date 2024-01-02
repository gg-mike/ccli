package router

import (
	"errors"

	"github.com/gg-mike/ccli/pkg/model"
	"github.com/gin-gonic/gin"
)

type PipelineRouter = IRouter[model.Pipeline, model.PipelineShort, model.PipelineInput]

func InitPipelineRouter(project *gin.RouterGroup) *gin.RouterGroup {
	r := NewRouter[model.Pipeline, model.PipelineShort, model.PipelineInput](
		// FILTER
		func(ctx *gin.Context) map[string]any {
			filters := map[string]any{}
			for key := range ctx.Request.URL.Query() {
				switch key {
				case "name":
					filters["name LIKE ?"] = "%" + ctx.Query(key) + "%"
				case "branch":
					filters["branch LIKE ?"] = "%" + ctx.Query(key) + "%"
				}
			}

			return filters
		},
		// GET SELECTOR
		func(params gin.Params) (model.Pipeline, error) {
			projectName, ok := params.Get("project_name")
			if !ok {
				return model.Pipeline{}, errors.New("missing param 'project_name'")
			}
			pipelineName, ok := params.Get("pipeline_name")
			if !ok {
				return model.Pipeline{}, errors.New("missing param 'pipeline_name'")
			}
			return model.Pipeline{Name: pipelineName, ProjectName: projectName}, nil
		},
		// GET PARENT
		func(params gin.Params) (model.Pipeline, error) {
			projectName, ok := params.Get("project_name")
			if !ok {
				return model.Pipeline{}, errors.New("missing param 'project_name'")
			}
			return model.Pipeline{ProjectName: projectName}, nil
		},
		// MERGE
		func(left model.Pipeline, right model.PipelineInput) model.Pipeline {
			left.Name = right.Name
			left.Branch = right.Branch
			left.Config = right.Config
			return left
		},
	)

	_rg := project.Group(":project_name/pipelines")

	_rg.GET("", getManyPipelines(r))
	_rg.POST("", createPipeline(r))
	_rg.GET(":pipeline_name", getOnePipeline(r))
	_rg.PUT(":pipeline_name", updatePipeline(r))
	_rg.DELETE(":pipeline_name", deletePipeline(r))

	return _rg
}

// @Summary  Get pipelines
// @ID       many-pipelines
// @Tags     pipelines
// @Produce  json
// @Param    project_name path  string true  "Project name"
// @Param    page         query int    false "Page number"
// @Param    size         query int    false "Page size"
// @Param    order        query string false "Order by field"
// @Param    name         query string false "Pipeline name (pattern)"
// @Param    branch       query string false "Pipeline branch (pattern)"
// @Success  200 {object} []model.PipelineShort "List of pipelines"
// @Failure  400 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /projects/{project_name}/pipelines [get]
func getManyPipelines(r PipelineRouter) gin.HandlerFunc {
	return r.GetMany
}

// @Summary  Create new pipeline
// @ID       create-pipeline
// @Tags     pipelines
// @Accept   json
// @Param    project_name path string              true "Project name"
// @Param    pipeline     body model.PipelineInput true "New pipeline entry"
// @Success  202 {string} Success message
// @Failure  400 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /projects/{project_name}/pipelines [post]
func createPipeline(r PipelineRouter) gin.HandlerFunc {
	return r.Create
}

// @Summary  Get the single pipeline
// @ID       single-pipeline
// @Tags     pipelines
// @Produce  json
// @Param    project_name  path string true "Project name"
// @Param    pipeline_name path string true "Pipeline name"
// @Success  201 {object} model.Pipeline "Requested pipeline"
// @Failure  400 {string} Error in request
// @Failure  404 {string} No record found
// @Failure  500 {string} Database error
// @Router   /projects/{project_name}/pipelines/{pipeline_name} [get]
func getOnePipeline(r PipelineRouter) gin.HandlerFunc {
	return r.GetOne
}

// @Summary  Update pipeline
// @ID       update-pipeline
// @Tags     pipelines
// @Accept   json
// @Param    project_name  path string             true "Project name"
// @Param    pipeline_name path string             true "Pipeline name"
// @Param    pipeline      body model.PipelineInput true "Updated pipeline entry"
// @Success  200 {object} model.Pipeline "Updated pipeline"
// @Failure  400 {string} Error in request
// @Failure  404 {string} No record found
// @Failure  500 {string} Database error
// @Router   /projects/{project_name}/pipelines/{pipeline_name} [put]
func updatePipeline(r PipelineRouter) gin.HandlerFunc {
	return r.Update
}

// @Summary  Delete pipeline
// @ID       delete-pipeline
// @Tags     pipelines
// @Param    project_name  path  string  true  "Project name"
// @Param    pipeline_name path  string  true  "Pipeline name"
// @Param    force         query boolean false "Force deletion"
// @Success  200 {string} Success message
// @Failure  404 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /projects/{project_name}/pipelines/{pipeline_name} [delete]
func deletePipeline(r PipelineRouter) gin.HandlerFunc {
	return r.Delete
}
