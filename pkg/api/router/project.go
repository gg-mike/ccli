package router

import (
	"errors"

	"github.com/gg-mike/ccli/pkg/model"
	"github.com/gin-gonic/gin"
)

type ProjectRouter = IRouter[model.Project, model.ProjectShort, model.ProjectInput]

func InitProjectRouter(base *gin.RouterGroup) *gin.RouterGroup {
	r := NewRouter[model.Project, model.ProjectShort, model.ProjectInput](
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
		func(params gin.Params) (model.Project, error) {
			projectName, ok := params.Get("project_name")
			if !ok {
				return model.Project{}, errors.New("missing param 'project_name'")
			}
			return model.Project{Name: projectName}, nil
		},
		// GET PARENT
		func(params gin.Params) (model.Project, error) {
			return model.Project{}, nil
		},
		// MERGE
		func(left model.Project, right model.ProjectInput) model.Project {
			left.Name = right.Name
			left.Repo = right.Repo
			return left
		},
	)

	_rg := base.Group("/projects")

	_rg.GET("", getManyProjects(r))
	_rg.POST("", createProject(r))
	_rg.GET(":project_name", getOneProject(r))
	_rg.PUT(":project_name", updateProject(r))
	_rg.DELETE(":project_name", deleteProject(r))

	return _rg
}

// @Summary  Get projects
// @ID       many-projects
// @Tags     projects
// @Produce  json
// @Param    page  query int    false "Page number"
// @Param    size  query int    false "Page size"
// @Param    order query string false "Order by field"
// @Param    name  query string false "Project name (pattern)"
// @Param    repo  query string false "Project repo (pattern)"
// @Success  200 {object} []model.ProjectShort "List of projects"
// @Failure  400 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /projects [get]
func getManyProjects(r ProjectRouter) gin.HandlerFunc {
	return r.GetMany
}

// @Summary  Create new project
// @ID       create-project
// @Tags     projects
// @Accept   json
// @Param    project body model.ProjectInput true "New project entry"
// @Success  202 {string} Success message
// @Failure  400 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /projects [post]
func createProject(r ProjectRouter) gin.HandlerFunc {
	return r.Create
}

// @Summary  Get the single project
// @ID       single-project
// @Tags     projects
// @Produce  json
// @Param    project_name path string true "Project name"
// @Success  201 {object} model.Project "Requested project"
// @Failure  400 {string} Error in request
// @Failure  404 {string} No record found
// @Failure  500 {string} Database error
// @Router   /projects/{project_name} [get]
func getOneProject(r ProjectRouter) gin.HandlerFunc {
	return r.GetOne
}

// @Summary  Update project
// @ID       update-project
// @Tags     projects
// @Accept   json
// @Param    project_name path string             true "Project name"
// @Param    project      body model.ProjectInput true "Updated project entry"
// @Success  200 {object} model.Project "Updated project"
// @Failure  400 {string} Error in request
// @Failure  404 {string} No record found
// @Failure  500 {string} Database error
// @Router   /projects/{project_name} [put]
func updateProject(r ProjectRouter) gin.HandlerFunc {
	return r.Update
}

// @Summary  Delete project
// @ID       delete-project
// @Tags     projects
// @Param    project_name path  string  true  "Project name"
// @Param    force        query boolean false "Force deletion"
// @Success  200 {string} Success message
// @Failure  404 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /projects/{project_name} [delete]
func deleteProject(r ProjectRouter) gin.HandlerFunc {
	return r.Delete
}
