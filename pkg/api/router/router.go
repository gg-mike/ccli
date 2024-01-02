package router

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gg-mike/ccli/pkg/api/handler"
	"github.com/gin-gonic/gin"
)

type Router[T, TShort, TInput any] struct {
	filter      func(ctx *gin.Context) map[string]any
	getSelector func(params gin.Params) (T, error)
	getParent   func(params gin.Params) (T, error)

	handler handler.IHandler[T, TShort, TInput]
}

type IRouter[T, TShort, TInput any] interface {
	GetMany(ctx *gin.Context)
	Create(ctx *gin.Context)
	GetOne(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

func NewRouter[T, TShort, TInput any](
	filter func(ctx *gin.Context) map[string]any,
	getSelector func(params gin.Params) (T, error),
	getParent func(params gin.Params) (T, error),
	merge func(T, TInput) T,
) IRouter[T, TShort, TInput] {
	return Router[T, TShort, TInput]{
		filter:      filter,
		getSelector: getSelector,
		getParent:   getParent,
		handler:     handler.NewHandler[T, TShort, TInput](merge),
	}
}

func (r Router[T, TShort, TInput]) GetMany(ctx *gin.Context) {
	page, size, order, err := extractPagination(ctx)
	if err != nil {
		ctx.String(http.StatusBadRequest, "error in query [%v]", err)
		return
	}
	projects, err := r.handler.GetMany(page, size, order, r.filter(ctx))
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, projects)
	case handler.ErrDatabase:
		ctx.String(http.StatusInternalServerError, "error during database operations")
	default:
		ctx.String(http.StatusBadRequest, err.Error())
	}
}

func (r Router[T, TShort, TInput]) Create(ctx *gin.Context) {
	parent, err := r.getParent(ctx.Params)
	if err != nil {
		ctx.String(http.StatusBadRequest, "error in params: [%v]", err)
	}
	m := *new(TInput)
	if err := ctx.BindJSON(&m); err != nil {
		ctx.String(http.StatusBadRequest, "error in json: [%v]", err)
	}
	switch r.handler.Create(parent, m) {
	case nil:
		ctx.String(http.StatusAccepted, "added new record")
	case handler.ErrDuplicate:
		ctx.String(http.StatusConflict, "duplicated key in entered data")
	case handler.ErrConflict:
		ctx.String(http.StatusConflict, "conflict in entered data")
	case handler.ErrDatabase:
		ctx.String(http.StatusInternalServerError, "error during database operations")
	}
}

func (r Router[T, TShort, TInput]) GetOne(ctx *gin.Context) {
	selector, err := r.getSelector(ctx.Params)
	if err != nil {
		ctx.String(http.StatusBadRequest, "error in params [%v]", err)
		return
	}
	project, err := r.handler.GetOne(selector)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, project)
	case handler.ErrRecordNotFound:
		ctx.String(http.StatusNotFound, "record not found")
	case handler.ErrDatabase:
		ctx.String(http.StatusInternalServerError, "error during database operations")
	}
}

func (r Router[T, TShort, TInput]) Update(ctx *gin.Context) {
	selector, err := r.getSelector(ctx.Params)
	if err != nil {
		ctx.String(http.StatusBadRequest, "error in params [%v]", err)
		return
	}
	m := *new(TInput)
	if err := ctx.BindJSON(&m); err != nil {
		ctx.String(http.StatusBadRequest, "error in json: [%v]", err)
	}
	switch r.handler.Update(selector, m) {
	case nil:
		ctx.String(http.StatusAccepted, "updated record")
	case handler.ErrRecordNotFound:
		ctx.String(http.StatusNotFound, "record not found")
	case handler.ErrDuplicate:
		ctx.String(http.StatusConflict, "duplicated key in entered data")
	case handler.ErrConflict:
		ctx.String(http.StatusConflict, "conflict in entered data")
	case handler.ErrDatabase:
		ctx.String(http.StatusInternalServerError, "error during database operations")
	}
}

func (r Router[T, TShort, TInput]) Delete(ctx *gin.Context) {
	selector, err := r.getSelector(ctx.Params)
	if err != nil {
		ctx.String(http.StatusBadRequest, "error in params [%v]", err)
		return
	}
	_, force := ctx.GetQuery("force")
	switch r.handler.Delete(selector, force) {
	case nil:
		ctx.String(http.StatusOK, "deleted record")
	case handler.ErrRecordNotFound:
		ctx.String(http.StatusNotFound, "record not found")
	case handler.ErrDatabase:
		ctx.String(http.StatusInternalServerError, "error during database operations")
	}
}

func extractPagination(ctx *gin.Context) (int, int, string, error) {
	var page, size int
	var err error
	page_, ok := ctx.GetQuery("page")
	if !ok {
		page = -1
	} else {
		page, err = strconv.Atoi(page_)
		if err != nil {
			return -1, -1, "", errors.New(`error parsing query param "page"`)
		}
	}

	size_, ok := ctx.GetQuery("size")
	if !ok {
		size = -1
	} else {
		size, err = strconv.Atoi(size_)
		if err != nil {
			return -1, -1, "", errors.New(`error parsing query param "size"`)
		}
	}

	order, _ := ctx.GetQuery("order")

	return page, size, order, nil
}
