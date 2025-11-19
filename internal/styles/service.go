package styles

import (
	"maps"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/nsr888/lifecalendar/internal/config"
	"github.com/nsr888/lifecalendar/internal/entity"
)

func NewService(
	cfg *config.Config,
	storage Storage,
	dayStyles map[time.Time]entity.DayInfo,
) *Service {
	return &Service{
		config:     cfg,
		categories: GenerateCategoryStyles(cfg.Categories),
		dayStyles:  dayStyles,
		storage:    storage,
	}
}

func (s *Service) GetCategoryStyle(category string) lipgloss.Style {
	if style, exists := s.categories[category]; exists {
		return style
	}
	return lipgloss.NewStyle()
}

func (s *Service) GetAllCategoryStyles() map[string]lipgloss.Style {
	result := make(map[string]lipgloss.Style)
	maps.Copy(result, s.categories)

	return result
}

func (s *Service) GetDayStyle(date time.Time) (entity.DayInfo, bool) {
	dayInfo, exists := s.dayStyles[date]
	return dayInfo, exists
}

func (s *Service) GetAllDayStyles() map[time.Time]entity.DayInfo {
	result := make(map[time.Time]entity.DayInfo)
	maps.Copy(result, s.dayStyles)

	return result
}

func ComputeYearStyles(
	config *config.Config,
	year int,
	data *entity.CategoryName,
) (map[time.Time]entity.DayInfo, error) {
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(year, 12, 31, 0, 0, 0, 0, time.Local)

	result := make(map[time.Time]entity.DayInfo)

	for currentDate := startDate; !currentDate.After(endDate); currentDate = currentDate.AddDate(0, 0, 1) {
		var winningCategory string
		winningPriority := 999 // Start with high number (low priority)

		// Check all categories for this date
		for categoryName, category := range data.Categories {
			if _, exists := category.Dates[currentDate]; exists {
				categoryConfig := config.GetCategoryConfig(categoryName)
				if categoryConfig.Priority < winningPriority {
					winningCategory = categoryName
					winningPriority = categoryConfig.Priority
				}
			}
		}

		// Store the winning category for this date
		if winningCategory != "" {
			result[currentDate] = entity.DayInfo{
				Category: winningCategory,
				Priority: winningPriority,
			}
		}
	}

	return result, nil
}

func (s *Service) GetPriority(category entity.CategoryType) int {
	categoryName := string(category)
	if config, exists := s.config.Categories[categoryName]; exists {
		return config.Priority
	}
	return 999 // High number = low priority
}

func GenerateCategoryStyles(categories map[string]config.CategoryConfig) map[string]lipgloss.Style {
	result := make(map[string]lipgloss.Style)

	for categoryName, categoryConfig := range categories {
		style := lipgloss.NewStyle()
		if categoryConfig.Fg != "" {
			style = style.Foreground(lipgloss.Color(categoryConfig.Fg))
		}
		if categoryConfig.Bg != "" {
			style = style.Background(lipgloss.Color(categoryConfig.Bg))
		}
		if categoryConfig.Bold {
			style = style.Bold(true)
		}
		if categoryConfig.Italic {
			style = style.Italic(true)
		}

		result[categoryName] = style
	}

	return result
}
