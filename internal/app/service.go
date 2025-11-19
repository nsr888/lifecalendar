package app

import (
	"fmt"
	"maps"
	"time"

	"github.com/nsr888/lifecalendar/internal/config"
	"github.com/nsr888/lifecalendar/internal/entity"
	"github.com/nsr888/lifecalendar/internal/render"
	"github.com/nsr888/lifecalendar/internal/storage"
	"github.com/nsr888/lifecalendar/internal/styles"
)

type Service struct {
	storage storage.Storage
	logger  Logger
}

type Logger interface {
	Printf(format string, v ...any)
	Fatal(v ...any)
}

func NewService(storage storage.Storage, logger Logger) *Service {
	return &Service{
		storage: storage,
		logger:  logger,
	}
}

func (s *Service) Run(initialConfig *config.Config) error {
	allDayStyles, err := s.computeAllDayStyles(initialConfig)
	if err != nil {
		return fmt.Errorf("failed to compute day styles: %w", err)
	}

	styleService := styles.NewService(initialConfig, s.storage, allDayStyles)

	return s.renderAllYears(initialConfig, styleService)
}

func (s *Service) computeAllDayStyles(cfg *config.Config) (map[time.Time]entity.DayInfo, error) {
	allDayStyles := make(map[time.Time]entity.DayInfo)

	for _, year := range cfg.Years {
		dayStyles, err := s.computeDayStylesForYear(year, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to compute day styles for year %d: %w", year, err)
		}
		maps.Copy(allDayStyles, dayStyles)
	}

	return allDayStyles, nil
}

func (s *Service) LoadCategoryByYearWithGenerated(year int) (*entity.CategoryName, error) {
	if !s.storage.IsYearDataExists(year) {
		return nil, fmt.Errorf("data for year does not exist: %d", year)
	}

	dataConfig, err := s.storage.LoadCategoryByYear(year)
	if err != nil {
		return nil, fmt.Errorf("failed to load data config for year %d: %w", year, err)
	}

	weekendDays := generateWeekendDays(year)
	dataConfig.Categories["weekends"] = &entity.Category{
		Type:  entity.CategoryWeekends,
		Dates: weekendDays,
	}

	oddWeeks := generateOddWeeks(year)
	dataConfig.Categories["odd_week"] = &entity.Category{
		Type:  entity.CategoryType("odd_week"),
		Dates: oddWeeks,
	}

	evenWeeks := generateEvenWeeks(year)
	dataConfig.Categories["even_week"] = &entity.Category{
		Type:  entity.CategoryType("even_week"),
		Dates: evenWeeks,
	}

	currentDays := generateCurrentDay(year)
	if len(currentDays) > 0 {
		dataConfig.Categories["current_day"] = &entity.Category{
			Type:  entity.CategoryType("current_day"),
			Dates: currentDays,
		}
	}

	return dataConfig, nil
}

func (s *Service) computeDayStylesForYear(year int, cfg *config.Config) (map[time.Time]entity.DayInfo, error) {
	dataConfig, err := s.LoadCategoryByYearWithGenerated(year)
	if err != nil {
		return nil, err
	}

	return styles.ComputeYearStyles(cfg, year, dataConfig)
}

func (s *Service) renderAllYears(cfg *config.Config, styleService styles.StyleService) error {
	for _, year := range cfg.Years {
		dataConfig, err := s.LoadCategoryByYearWithGenerated(year)
		if err != nil {
			return fmt.Errorf("failed to load data config for year %d: %w", year, err)
		}

		renderService := render.NewRenderService(year, dataConfig, cfg, styleService)
		renderService.SetMaxWidth(cfg.Rendering.MaxWidthInChars)

		labeledCategories, err := s.storage.LoadLabeledCategories(year)
		if err != nil {
			return fmt.Errorf("failed to load labeled categories for year %d: %w", year, err)
		}

		renderService.RenderYearTitle(year)
		renderService.RenderCompactYearViewWithSidePanel(labeledCategories)
	}

	return nil
}

func generateCurrentDay(year int) map[time.Time]struct{} {
	currentDay := make(map[time.Time]struct{})

	now := time.Now()
	if now.Year() == year {
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		currentDay[today] = struct{}{}
	}

	return currentDay
}

func generateWeekendDays(year int) map[time.Time]struct{} {
	weekendDays := make(map[time.Time]struct{})

	for month := 1; month <= 12; month++ {
		daysInMonth := daysIn(month, year)
		for day := 1; day <= daysInMonth; day++ {
			date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
			if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
				weekendDays[date] = struct{}{}
			}
		}
	}

	return weekendDays
}

func daysIn(month, year int) int {
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.Local).Day()
}

func generateOddWeeks(year int) map[time.Time]struct{} {
	oddWeeks := make(map[time.Time]struct{})

	for month := 1; month <= 12; month++ {
		daysInMonth := daysIn(month, year)
		for day := 1; day <= daysInMonth; day++ {
			date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
			_, week := date.ISOWeek()
			if week%2 == 1 {
				oddWeeks[date] = struct{}{}
			}
		}
	}

	return oddWeeks
}

func generateEvenWeeks(year int) map[time.Time]struct{} {
	evenWeeks := make(map[time.Time]struct{})

	for month := 1; month <= 12; month++ {
		daysInMonth := daysIn(month, year)
		for day := 1; day <= daysInMonth; day++ {
			date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
			_, week := date.ISOWeek()
			if week%2 == 0 {
				evenWeeks[date] = struct{}{}
			}
		}
	}

	return evenWeeks
}
