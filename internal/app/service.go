package app

import (
	"encoding/json"
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

func (s *Service) computeAllDayStyles(
	cfg *config.Config,
) (map[time.Time]entity.DayInfo, error) {
	allDayStyles := make(map[time.Time]entity.DayInfo)

	for _, year := range cfg.Years {
		dayStyles, err := s.computeDayStylesForYear(year, cfg)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to compute day styles for year %d: %w",
				year,
				err,
			)
		}
		maps.Copy(allDayStyles, dayStyles)
	}

	return allDayStyles, nil
}

func (s *Service) LoadCategoryByYearWithGenerated(
	year int,
) (*entity.CategoryName, error) {
	if !s.storage.IsYearDataExists(year) {
		return nil, fmt.Errorf("data for year does not exist: %d", year)
	}

	dataConfig, err := s.storage.LoadCategoryByYear(year)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to load data config for year %d: %w",
			year,
			err,
		)
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

func (s *Service) computeDayStylesForYear(
	year int,
	cfg *config.Config,
) (map[time.Time]entity.DayInfo, error) {
	dataConfig, err := s.LoadCategoryByYearWithGenerated(year)
	if err != nil {
		return nil, err
	}

	return styles.ComputeYearStyles(cfg, year, dataConfig)
}

func (s *Service) renderAllYears(
	cfg *config.Config,
	styleService styles.StyleService,
) error {
	for _, year := range cfg.Years {
		dataConfig, err := s.LoadCategoryByYearWithGenerated(year)
		if err != nil {
			return fmt.Errorf(
				"failed to load data config for year %d: %w",
				year,
				err,
			)
		}

		renderService := render.NewService(year, dataConfig, cfg, styleService)
		renderService.SetMaxWidth(cfg.Rendering.MaxWidthInChars)

		labeledCategories, err := s.storage.LoadLabeledCategories(year)
		if err != nil {
			return fmt.Errorf(
				"failed to load labeled categories for year %d: %w",
				year,
				err,
			)
		}

		renderService.RenderYearTitle(year)

		// Choose rendering format based on config
		switch cfg.Rendering.Format {
		case "three_column":
			renderService.RenderThreeColumnView(labeledCategories)
		default:
			renderService.RenderCompactYearViewWithSidePanel(labeledCategories)
		}
	}

	return nil
}

func generateCurrentDay(year int) map[time.Time]struct{} {
	currentDay := make(map[time.Time]struct{})

	now := time.Now()
	if now.Year() == year {
		today := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			0,
			0,
			0,
			0,
			time.Local,
		)
		currentDay[today] = struct{}{}
	}

	return currentDay
}

func generateWeekendDays(year int) map[time.Time]struct{} {
	weekendDays := make(map[time.Time]struct{})

	for month := 1; month <= 12; month++ {
		daysInMonth := daysIn(month, year)
		for day := 1; day <= daysInMonth; day++ {
			date := time.Date(
				year,
				time.Month(month),
				day,
				0,
				0,
				0,
				0,
				time.Local,
			)
			if date.Weekday() == time.Saturday ||
				date.Weekday() == time.Sunday {
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
			date := time.Date(
				year,
				time.Month(month),
				day,
				0,
				0,
				0,
				0,
				time.Local,
			)
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
			date := time.Date(
				year,
				time.Month(month),
				day,
				0,
				0,
				0,
				0,
				time.Local,
			)
			_, week := date.ISOWeek()
			if week%2 == 0 {
				evenWeeks[date] = struct{}{}
			}
		}
	}

	return evenWeeks
}

func (s *Service) countWeekendsAndHolidays(
	start, end time.Time,
	holidays map[time.Time]struct{},
) (weekendCount, holidayCount int) {
	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		if current.Weekday() == time.Saturday ||
			current.Weekday() == time.Sunday {
			weekendCount++
		}

		if _, isHoliday := holidays[current]; isHoliday {
			holidayCount++
		}
	}
	return
}

func (s *Service) RunJSONPlan(cfg *config.Config) error {
	var allPlans []entity.VacationPlanJSON
	var potentialPlans []entity.PotentialVacation
	var allPlansWithPotential []entity.EnhancedJSONPlanResponse

	for _, year := range cfg.Years {
		labeledCategories, err := s.storage.LoadLabeledCategories(year)
		if err != nil {
			return fmt.Errorf(
				"failed to load labeled categories for year %d: %w",
				year,
				err,
			)
		}

		publicHolidays := make(map[time.Time]struct{})
		dataConfig, err := s.storage.LoadCategoryByYear(year)
		if err != nil {
			return fmt.Errorf(
				"failed to load data config for year %d: %w",
				year,
				err,
			)
		}
		if holidays, exists := dataConfig.Categories["public_holidays"]; exists {
			publicHolidays = holidays.Dates
		}

		for _, category := range labeledCategories {
			for _, entry := range category.Entries {
				weekendCount, holidayCount := s.countWeekendsAndHolidays(
					entry.DateStart,
					entry.DateEnd,
					publicHolidays,
				)

				totalDays := int(
					entry.DateEnd.Sub(entry.DateStart).Hours()/24,
				) + 1

				plan := entity.VacationPlanJSON{
					DateStart:    entry.DateStart.Format("2006-01-02"),
					DateEnd:      entry.DateEnd.Format("2006-01-02"),
					Label:        entry.Label,
					WeekendCount: weekendCount,
					HolidayCount: holidayCount,
					TotalDays:    totalDays,
				}

				allPlans = append(allPlans, plan)
			}
		}

		yearPotentialPlans, err := s.findConsecutiveWeekendHolidayBlocks(
			year,
			publicHolidays,
			allPlans,
		)
		if err != nil {
			return fmt.Errorf(
				"failed to find potential vacation plans for year %d: %w",
				year,
				err,
			)
		}

		potentialPlans = append(potentialPlans, yearPotentialPlans...)
		allPlansWithPotential = append(
			allPlansWithPotential,
			entity.EnhancedJSONPlanResponse{
				ExistingVacations:  allPlans,
				PotentialVacations: potentialPlans,
				Year:               year,
			},
		)
	}

	jsonData, err := json.Marshal(allPlansWithPotential)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

func (s *Service) findConsecutiveWeekendHolidayBlocks(
	year int,
	holidays map[time.Time]struct{},
	existingPlans []entity.VacationPlanJSON,
) ([]entity.PotentialVacation, error) {
	nonWorkingDays := make(map[time.Time]bool)

	weekendDays := generateWeekendDays(year)
	for date := range weekendDays {
		nonWorkingDays[date] = true
	}

	for date := range holidays {
		nonWorkingDays[date] = true
	}

	for _, plan := range existingPlans {
		startDate, err := time.ParseInLocation("2006-01-02", plan.DateStart, time.Local)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start date: %w", err)
		}
		endDate, err := time.ParseInLocation("2006-01-02", plan.DateEnd, time.Local)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end date: %w", err)
		}

		for current := startDate; !current.After(endDate); current = current.AddDate(0, 0, 1) {
			nonWorkingDays[current] = false
		}
	}

	startDate := time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(year, time.December, 31, 0, 0, 0, 0, time.Local)

	var potentialVacations []entity.PotentialVacation
	var currentSequence []time.Time

	for current := startDate; !current.After(endDate); current = current.AddDate(0, 0, 1) {
		if nonWorkingDays[current] {
			currentSequence = append(currentSequence, current)
		} else {
			if len(currentSequence) > 2 {
				firstDay := currentSequence[0]
				lastDay := currentSequence[len(currentSequence)-1]

				weekendCount := 0
				holidayCount := 0
				for _, date := range currentSequence {
					if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
						weekendCount++
					}
					if _, isHoliday := holidays[date]; isHoliday {
						holidayCount++
					}
				}

				potential := entity.PotentialVacation{
					DateStart:    firstDay.Format("2006-01-02"),
					DateEnd:      lastDay.Format("2006-01-02"),
					WeekendCount: weekendCount,
					HolidayCount: holidayCount,
					TotalDays:    len(currentSequence),
					Description:  fmt.Sprintf("%d-day natural break: %d weekends, %d holidays", len(currentSequence), weekendCount, holidayCount),
				}

				potentialVacations = append(potentialVacations, potential)
			}

			currentSequence = []time.Time{}
		}
	}

	return potentialVacations, nil
}
