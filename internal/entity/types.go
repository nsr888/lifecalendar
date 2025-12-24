package entity

import "time"

type Date = time.Time

type CategoryType string

const (
	CategoryPlans    CategoryType = "plans"
	CategoryWeekends CategoryType = "weekends"
)

type CategoryEntry struct {
	DateStart time.Time
	DateEnd   time.Time
	Label     string
}

type Category struct {
	Type    CategoryType
	Desc    string
	Dates   map[time.Time]struct{} // For backward compatibility
	Entries []CategoryEntry        // New unified format
}

type CategoryName struct {
	BaseYear   int
	Categories map[string]*Category
}

type DayInfo struct {
	Category string
	Priority int
}

type VacationPlanJSON struct {
	DateStart    string `json:"date_start"`
	DateEnd      string `json:"date_end"`
	Label        string `json:"label"`
	WeekendCount int    `json:"weekend_count"`
	HolidayCount int    `json:"holiday_count"`
	TotalDays    int    `json:"total_days"`
}

type PotentialVacation struct {
	DateStart    string `json:"date_start"`
	DateEnd      string `json:"date_end"`
	WeekendCount int    `json:"weekend_count"`
	HolidayCount int    `json:"holiday_count"`
	TotalDays    int    `json:"total_days"`
	Description  string `json:"description"`
}

type EnhancedJSONPlanResponse struct {
	ExistingVacations  []VacationPlanJSON  `json:"existing_vacations"`
	PotentialVacations []PotentialVacation `json:"potential_vacations"`
	Year               int                 `json:"year"`
}

type RenderContext struct {
	FirstWeekday int              // Monday = 0
	WeekendDays  map[int]struct{} // {5,6} for Sat/Sun
	MonthNames   map[int]string
	WeekdayNames []string
}
