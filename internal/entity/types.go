package entity

import "time"

// Date represents a calendar date.
type Date = time.Time

// CategoryType represents a category name as string.
// This allows for dynamic categories based on CSV filenames.
type CategoryType string

// Predefined category constants for backward compatibility.
const (
	CategoryPlans    CategoryType = "plans"
	CategoryWeekends CategoryType = "weekends"
)

// CategoryEntry represents a single entry in a category.
type CategoryEntry struct {
	DateStart time.Time
	DateEnd   time.Time
	Label     string
}

// Category represents a type of day with dates and entries.
type Category struct {
	Type    CategoryType
	Desc    string
	Dates   map[time.Time]struct{} // For backward compatibility
	Entries []CategoryEntry        // New unified format
}

// CategoryName represents the application configuration.
type CategoryName struct {
	BaseYear   int
	Categories map[string]*Category
}

// DayInfo contains styling information for a specific date.
type DayInfo struct {
	Category string
	Priority int
}

// RenderContext contains information needed for rendering.
type RenderContext struct {
	FirstWeekday int              // Monday = 0
	WeekendDays  map[int]struct{} // {5,6} for Sat/Sun
	MonthNames   map[int]string
	WeekdayNames []string
}
