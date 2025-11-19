package styles

import (
	"time"

	"github.com/nsr888/lifecalendar/internal/config"
	"github.com/nsr888/lifecalendar/internal/entity"
	"github.com/charmbracelet/lipgloss"
)

// StyleService provides unified style management for categories and day-level styling
type StyleService interface {
	// Visual style management
	GetCategoryStyle(category string) lipgloss.Style
	GetAllCategoryStyles() map[string]lipgloss.Style

	// Day style access
	GetDayStyle(date time.Time) (entity.DayInfo, bool)
	GetAllDayStyles() map[time.Time]entity.DayInfo

	// Priority management for backward compatibility
	GetPriority(category entity.CategoryType) int
}

// Service implements StyleService
type Service struct {
	config     *config.Config
	categories map[string]lipgloss.Style
	dayStyles  map[time.Time]entity.DayInfo
	storage    Storage
}

// Storage interface for data access
type Storage interface {
	LoadCategoryByYear(year int) (*entity.CategoryName, error)
	IsYearDataExists(year int) bool
	GetCategoryNames(year int) ([]string, error)
}
