package router

import (
	"errors"

	"github.com/gg-mike/ccli/pkg/model"
	"github.com/gin-gonic/gin"
)

type WorkerRouter = IRouter[model.Worker, model.WorkerShort, model.WorkerInput]

func InitWorkerRouter(base *gin.RouterGroup) {
	r := NewRouter[model.Worker, model.WorkerShort, model.WorkerInput](
		// FILTER
		func(ctx *gin.Context) map[string]any {
			filters := map[string]any{}
			for key := range ctx.Request.URL.Query() {
				switch key {
				case "name":
					filters["name   LIKE ?"] = "%" + ctx.Query(key) + "%"
				case "system":
					filters["system LIKE ?"] = "%" + ctx.Query(key) + "%"
				case "status":
					filters["status IN ?"] = ctx.QueryArray(key)
				case "static":
					_, ok := ctx.GetQuery(key)
					filters["is_static = ?"] = ok
				}
			}

			return filters
		},
		// GET SELECTOR
		func(params gin.Params) (model.Worker, error) {
			workerName, ok := params.Get("worker_name")
			if !ok {
				return model.Worker{}, errors.New("missing param 'worker_name'")
			}
			return model.Worker{Name: workerName}, nil
		},
		// GET PARENT
		func(params gin.Params) (model.Worker, error) {
			return model.Worker{}, nil
		},
		// MERGE
		func(left model.Worker, right model.WorkerInput) model.Worker {
			left.Name = right.Name
			left.Address = right.Address
			left.System = right.System
			left.IsStatic = right.IsStatic
			left.Strategy = right.Strategy
			left.Username = right.Username
			left.Capacity = right.Capacity
			return left
		},
	)

	_rg := base.Group("/workers")

	_rg.GET("", getManyWorkers(r))
	_rg.POST("", createWorker(r))
	_rg.GET(":worker_name", getOneWorker(r))
	_rg.PUT(":worker_name", updateWorker(r))
	_rg.DELETE(":worker_name", deleteWorker(r))
}

// @Summary  Get workers
// @ID       many-workers
// @Tags     workers
// @Produce  json
// @Param    page      query int    false "Page number"
// @Param    size      query int    false "Page size"
// @Param    order     query string false "Order by field"
// @Param    name      query string false "Worker name (pattern)"
// @Param    system    query string false "Worker system (pattern)"
// @Param    status    query []int  false "Worker status (possible values)"
// @Param    is_static query bool   false "Worker type"
// @Success  200 {object} []model.WorkerShort "List of workers"
// @Failure  400 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /workers [get]
func getManyWorkers(r WorkerRouter) gin.HandlerFunc {
	return r.GetMany
}

// @Summary  Create new worker
// @ID       create-worker
// @Tags     workers
// @Accept   json
// @Param    worker body model.WorkerInput true "New worker entry"
// @Success  202 {string} Success message
// @Failure  400 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /workers [post]
func createWorker(r WorkerRouter) gin.HandlerFunc {
	return r.Create
}

// @Summary  Get the single worker
// @ID       single-worker
// @Tags     workers
// @Produce  json
// @Param    worker_name path string true "Worker name"
// @Success  201 {object} model.Worker "Requested worker"
// @Failure  400 {string} Error in request
// @Failure  404 {string} No record found
// @Failure  500 {string} Database error
// @Router   /workers/{worker_name} [get]
func getOneWorker(r WorkerRouter) gin.HandlerFunc {
	return r.GetOne
}

// @Summary  Update worker
// @ID       update-worker
// @Tags     workers
// @Accept   json
// @Param    worker_name path string             true "Worker name"
// @Param    worker      body model.WorkerInput true "Updated worker entry"
// @Success  200 {object} model.Worker "Updated worker"
// @Failure  400 {string} Error in request
// @Failure  404 {string} No record found
// @Failure  500 {string} Database error
// @Router   /workers/{worker_name} [put]
func updateWorker(r WorkerRouter) gin.HandlerFunc {
	return r.Update
}

// @Summary  Delete worker
// @ID       delete-worker
// @Tags     workers
// @Param    worker_name path  string  true  "Worker name"
// @Param    force        query boolean false "Force deletion"
// @Success  200 {string} Success message
// @Failure  404 {string} Error in request
// @Failure  500 {string} Database error
// @Router   /workers/{worker_name} [delete]
func deleteWorker(r WorkerRouter) gin.HandlerFunc {
	return r.Delete
}
