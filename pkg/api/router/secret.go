package router

import (
	"database/sql"
	"errors"

	"github.com/gg-mike/ccli/pkg/model"
	"github.com/gin-gonic/gin"
)

type SecretRouter = IRouter[model.Secret, model.Secret, model.SecretInput]

func InitSecretRouter(base, project, pipeline *gin.RouterGroup) {
	r := NewRouter[model.Secret, model.Secret, model.SecretInput](
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
		func(params gin.Params) (model.Secret, error) {
			projectName, okProject := params.Get("project_name")
			_projectName := sql.NullString{String: projectName, Valid: okProject}
			pipelineName, okPipeline := params.Get("pipeline_name")
			_pipelineName := sql.NullString{String: pipelineName, Valid: okPipeline}

			secretKey, ok := params.Get("secret_key")
			if !ok {
				return model.Secret{}, errors.New("missing param 'secret_key'")
			}
			return model.Secret{Key: secretKey, ProjectName: _projectName, PipelineName: _pipelineName}, nil
		},
		// GET PARENT
		func(params gin.Params) (model.Secret, error) {
			projectName, okProject := params.Get("project_name")
			_projectName := sql.NullString{String: projectName, Valid: okProject}
			pipelineName, okPipeline := params.Get("pipeline_name")
			_pipelineName := sql.NullString{String: pipelineName, Valid: okPipeline}

			return model.Secret{ProjectName: _projectName, PipelineName: _pipelineName}, nil
		},
		// MERGE
		func(left model.Secret, right model.SecretInput) model.Secret {
			left.Key = right.Key
			left.Path = right.Path
			return left
		},
	)

	{
		_rg := base.Group("/secrets")

		_rg.GET("", getManySecrets(r))
		_rg.POST("", createSecret(r))
		_rg.PUT(":secret_key", updateSecret(r))
		_rg.DELETE(":secret_key", deleteSecret(r))
	}
	{
		_rg := project.Group(":project_name/secrets")

		_rg.GET("", getManySecrets(r))
		_rg.POST("", createSecret(r))
		_rg.PUT(":secret_key", updateSecret(r))
		_rg.DELETE(":secret_key", deleteSecret(r))
	}
	{
		_rg := pipeline.Group(":pipeline_name/secrets")

		_rg.GET("", getManySecrets(r))
		_rg.POST("", createSecret(r))
		_rg.PUT(":secret_key", updateSecret(r))
		_rg.DELETE(":secret_key", deleteSecret(r))
	}
}

// @Summary  Get secrets
// @Tags     secrets
// @Produce  json
// @Param    project_name  path  string true  "Project name"
// @Param    pipeline_name path  string true  "Pipeline name"
// @Param    page          query int    false "Page number"
// @Param    size          query int    false "Page size"
// @Param    order         query string false "Order by field"
// @Param    key           query string false "Secret key (pattern)"
// @Success  200 {object} []model.Secret "List of secrets"
// @Failure  400 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /secrets [get]
// @Router   /projects/{project_name}/secrets [get]
// @Router   /projects/{project_name}/pipelines/{pipeline_name}/secrets [get]
func getManySecrets(r SecretRouter) gin.HandlerFunc {
	return r.GetMany
}

// @Summary  Create new secret
// @Tags     secrets
// @Accept   json
// @Param    project_name  path string            true  "Project name"
// @Param    pipeline_name path string            true  "Pipeline name"
// @Param    secret        body model.SecretInput true "New secret entry"
// @Success  202 {string} Success message
// @Failure  400 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /secrets [post]
// @Router   /projects/{project_name}/secrets [post]
// @Router   /projects/{project_name}/pipelines/{pipeline_name}/secrets [post]
func createSecret(r SecretRouter) gin.HandlerFunc {
	return r.Create
}

// @Summary  Update secret
// @Tags     secrets
// @Accept   json
// @Param    project_name  path string            true "Project name"
// @Param    pipeline_name path string            true "Pipeline name"
// @Param    secret_key    path string            true "Secret key"
// @Param    secret        body model.SecretInput true "Updated secret entry"
// @Success  200 {object} model.Secret "Updated secret"
// @Failure  400 {string} Error in request
// @Failure  404 {string} No record found
// @Failure  500 {string} Database error
// @Router   /secrets/{secret_key} [put]
// @Router   /projects/{project_name}/secrets/{secret_key} [put]
// @Router   /projects/{project_name}/pipelines/{pipeline_name}/secrets/{secret_key} [put]
func updateSecret(r SecretRouter) gin.HandlerFunc {
	return r.Update
}

// @Summary  Delete secret
// @Tags     secrets
// @Param    project_name  path string true "Project name"
// @Param    pipeline_name path string true "Pipeline name"
// @Param    secret_key    path string true "Secret key"
// @Success  200 {string} Success message
// @Failure  404 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /secrets/{secret_key} [delete]
// @Router   /projects/{project_name}/secrets/{secret_key} [delete]
// @Router   /projects/{project_name}/pipelines/{pipeline_name}/secrets/{secret_key} [delete]
func deleteSecret(r SecretRouter) gin.HandlerFunc {
	return r.Delete
}
