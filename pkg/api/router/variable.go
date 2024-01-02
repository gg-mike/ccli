package router

import (
	"database/sql"
	"errors"

	"github.com/gg-mike/ccli/pkg/model"
	"github.com/gin-gonic/gin"
)

type VariableRouter = IRouter[model.Variable, model.Variable, model.VariableInput]

func InitVariableRouter(base, project, pipeline *gin.RouterGroup) {
	r := NewRouter[model.Variable, model.Variable, model.VariableInput](
		// FILTER
		func(ctx *gin.Context) map[string]any {
			filters := map[string]any{}
			for key := range ctx.Request.URL.Query() {
				switch key {
				case "key":
					filters["key LIKE ?"] = "%" + ctx.Query(key) + "%"
				}
			}

			return filters
		},
		// GET SELECTOR
		func(params gin.Params) (model.Variable, error) {
			projectName, okProject := params.Get("project_name")
			_projectName := sql.NullString{String: projectName, Valid: okProject}
			pipelineName, okPipeline := params.Get("pipeline_name")
			_pipelineName := sql.NullString{String: pipelineName, Valid: okPipeline}

			variableKey, ok := params.Get("variable_key")
			if !ok {
				return model.Variable{}, errors.New("missing param 'variable_key'")
			}
			return model.Variable{Key: variableKey, ProjectName: _projectName, PipelineName: _pipelineName}, nil
		},
		// GET PARENT
		func(params gin.Params) (model.Variable, error) {
			projectName, okProject := params.Get("project_name")
			_projectName := sql.NullString{String: projectName, Valid: okProject}
			pipelineName, okPipeline := params.Get("pipeline_name")
			_pipelineName := sql.NullString{String: pipelineName, Valid: okPipeline}

			return model.Variable{ProjectName: _projectName, PipelineName: _pipelineName}, nil
		},
		// MERGE
		func(left model.Variable, right model.VariableInput) model.Variable {
			left.Key = right.Key
			left.Value = right.Value
			left.Path = right.Path
			return left
		},
	)

	{
		_rg := base.Group("/variables")

		_rg.GET("", getManyVariables(r))
		_rg.POST("", createVariable(r))
		_rg.PUT(":variable_key", updateVariable(r))
		_rg.DELETE(":variable_key", deleteVariable(r))
	}
	{
		_rg := project.Group(":project_name/variables")

		_rg.GET("", getManyVariables(r))
		_rg.POST("", createVariable(r))
		_rg.PUT(":variable_key", updateVariable(r))
		_rg.DELETE(":variable_key", deleteVariable(r))
	}
	{
		_rg := pipeline.Group(":pipeline_name/variables")

		_rg.GET("", getManyVariables(r))
		_rg.POST("", createVariable(r))
		_rg.PUT(":variable_key", updateVariable(r))
		_rg.DELETE(":variable_key", deleteVariable(r))
	}
}

// @Summary  Get variables
// @Tags     variables
// @Produce  json
// @Param    project_name  path  string true  "Project name"
// @Param    pipeline_name path  string true  "Pipeline name"
// @Param    page          query int    false "Page number"
// @Param    size          query int    false "Page size"
// @Param    order         query string false "Order by field"
// @Param    key           query string false "Variable key (pattern)"
// @Success  200 {object} []model.Variable "List of variables"
// @Failure  400 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /variables [get]
// @Router   /projects/{project_name}/variables [get]
// @Router   /projects/{project_name}/pipelines/{pipeline_name}/variables [get]
func getManyVariables(r VariableRouter) gin.HandlerFunc {
	return r.GetMany
}

// @Summary  Create new variable
// @Tags     variables
// @Accept   json
// @Param    project_name  path string              true "Project name"
// @Param    pipeline_name path string              true "Pipeline name"
// @Param    variable      body model.VariableInput true "New variable entry"
// @Success  202 {string} Success message
// @Failure  400 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /variables [post]
// @Router   /projects/{project_name}/variables [post]
// @Router   /projects/{project_name}/pipelines/{pipeline_name}/variables [post]
func createVariable(r VariableRouter) gin.HandlerFunc {
	return r.Create
}

// @Summary  Update variable
// @Tags     variables
// @Accept   json
// @Param    project_name  path string             true "Project name"
// @Param    pipeline_name path string             true "Pipeline name"
// @Param    variable_key  path string             true "Variable key"
// @Param    variable      body model.VariableInput true "Updated variable entry"
// @Success  200 {object} model.Variable "Updated variable"
// @Failure  400 {string} Error in request
// @Failure  404 {string} No record found
// @Failure  500 {string} Database error
// @Router   /variables/{variable_key} [put]
// @Router   /projects/{project_name}/variables/{variable_key} [put]
// @Router   /projects/{project_name}/pipelines/{pipeline_name}/variables/{variable_key} [put]
func updateVariable(r VariableRouter) gin.HandlerFunc {
	return r.Update
}

// @Summary  Delete variable
// @Tags     variables
// @Param    project_name  path  string  true  "Project name"
// @Param    pipeline_name path  string  true  "Pipeline name"
// @Param    variable_key  path  string  true  "Variable key"
// @Success  200 {string} Success message
// @Failure  404 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /variables/{variable_key} [delete]
// @Router   /projects/{project_name}/variables/{variable_key} [delete]
// @Router   /projects/{project_name}/pipelines/{pipeline_name}/variables/{variable_key} [delete]
func deleteVariable(r VariableRouter) gin.HandlerFunc {
	return r.Delete
}
