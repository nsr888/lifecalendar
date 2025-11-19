package storage

import (
	"github.com/nsr888/lifecalendar/internal/entity"
)

type Storage interface {
	IsYearDataExists(year int) bool
	GetCategoryNames(year int) ([]string, error)
	LoadCategoryByYear(year int) (*entity.CategoryName, error)
	LoadLabeledCategories(year int) ([]LabeledCategory, error)
}
