package handler

import (
	"errors"

	"github.com/gg-mike/ccli/pkg/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Handler[T, TShort, TInput any] struct {
	merge func(T, TInput) T
}

type IHandler[T, TShort, TInput any] interface {
	GetMany(page, size int, order string, filters map[string]any) ([]TShort, error)
	Create(parent T, m TInput) error
	GetOne(selector T) (T, error)
	Update(selector T, m TInput) error
	Delete(selector T, force bool) error
}

func NewHandler[T, TShort, TInput any](merge func(T, TInput) T) IHandler[T, TShort, TInput] {
	return Handler[T, TShort, TInput]{
		merge: merge,
	}
}

func (h Handler[T, TShort, TInput]) GetMany(page, size int, order string, filters map[string]any) ([]TShort, error) {
	offset, limit, order, err := paginate(page, size, order)
	if err != nil {
		return *new([]TShort), err
	}

	o := new([]TShort)
	if err = filter[T, TShort](o, filters).Offset(offset).Limit(limit).Order(order).Find(o).Error; err != nil && err != gorm.ErrRecordNotFound {
		return *new([]TShort), ErrDatabase
	}
	return *o, nil
}

func (h Handler[T, TShort, TInput]) Create(parent T, m TInput) error {
	_m := h.merge(parent, m)
	err := db.Get().InstanceSet("input", m).Create(&_m).Error
	switch err {
	case nil:
		return nil
	case gorm.ErrDuplicatedKey:
		return ErrDuplicate
	case gorm.ErrForeignKeyViolated:
		return ErrConflict
	default:
		return ErrDatabase
	}
}

func (h Handler[T, TShort, TInput]) GetOne(query T) (T, error) {
	m := *new(T)
	err := db.Get().Preload(clause.Associations).First(&m, &query).Error
	switch err {
	case nil:
		return m, nil
	case gorm.ErrRecordNotFound:
		return *new(T), ErrRecordNotFound
	default:
		return *new(T), ErrDatabase
	}
}

func (h Handler[T, TShort, TInput]) Delete(query T, force bool) error {
	o := new([]T)
	if err := db.Get().Clauses(clause.Returning{}).InstanceSet("force", force).Where(&query).Delete(&o).Error; err != nil {
		return ErrDatabase
	} else if len(*o) == 0 {
		return ErrRecordNotFound
	} else {
		return nil
	}
}

func (h Handler[T, TShort, TInput]) Update(selector T, m TInput) error {
	prev, err := h.GetOne(selector)
	if err != nil {
		return err
	}

	err = db.Get().Model(&selector).InstanceSet("input", m).InstanceSet("prev", prev).Updates(h.merge(prev, m)).Error
	switch err {
	case nil:
		return nil
	case gorm.ErrDuplicatedKey:
		return ErrDuplicate
	case gorm.ErrForeignKeyViolated:
		return ErrConflict
	default:
		return ErrDatabase
	}
}

func paginate(page, size int, order string) (int, int, string, error) {
	if page >= 0 && size < 0 {
		return -1, -1, "", errors.New("page size cannot be lower than 0 for nonnegative page number")
	}
	if size == 0 {
		return -1, -1, "", errors.New("page size cannot be 0")
	}
	if page < 0 && size < 0 {
		return -1, -1, order, nil
	}
	if page < 0 && size >= 0 {
		return -1, size, order, nil
	}

	return page * size, size, order, nil
}

func filter[T, TShort any](o *[]TShort, filters map[string]any) *gorm.DB {
	_db := db.Get().Model(*new(T))
	for key, value := range filters {
		_db = _db.Where(key, value)
	}
	return _db
}
