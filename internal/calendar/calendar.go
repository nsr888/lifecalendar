package calendar

import (
	"time"

	"github.com/nsr888/lifecalendar/internal/entity"
)

func NewRenderContext() *entity.RenderContext {
	return &entity.RenderContext{
		FirstWeekday: 0,
		WeekendDays:  getDefaultWeekendDays(),
		MonthNames:   getDefaultMonthNames(),
		WeekdayNames: getDefaultWeekdayNames(),
	}
}

func MonthCalendar(year int, month time.Month) [][]int {
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	weekdayFirst := (int(firstOfMonth.Weekday()) + 6) % 7

	dim := daysInMonth(year, month)

	var weeks [][]int
	curDay := 1 - weekdayFirst

	for {
		week := make([]int, 7)
		for i := range 7 {
			if curDay >= 1 && curDay <= dim {
				week[i] = curDay
			} else {
				week[i] = 0
			}
			curDay++
		}
		weeks = append(weeks, week)

		if curDay > dim || len(weeks) >= 6 {
			break
		}
	}
	return weeks
}

// CountDaysInYear counts vacation and personal days in a year.
func CountDaysInYear(cfg *entity.CategoryName, year int) (int, int) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.Local)
	return CountDaysInPeriod(cfg, start, end)
}

// CountDaysInPeriod counts vacation and personal days in a period.
func CountDaysInPeriod(cfg *entity.CategoryName, start, end time.Time) (int, int) {
	var vacDays, persDays int
	vacCat := cfg.Categories["vacations"]
	persCat := cfg.Categories["personal_days"]
	holCat := cfg.Categories["public_holidays"]

	ctx := NewRenderContext()

	for cur := start; cur.Before(end); cur = cur.AddDate(0, 0, 1) {
		weekday := (int(cur.Weekday()) + 6) % 7

		// Skip weekends
		if _, isWknd := ctx.WeekendDays[weekday]; isWknd {
			continue
		}

		// Skip public holidays
		if holCat != nil {
			if _, isHol := holCat.Dates[cur]; isHol {
				continue
			}
		}

		if vacCat != nil {
			if _, ok := vacCat.Dates[cur]; ok {
				vacDays++
			}
		}
		if persCat != nil {
			if _, ok := persCat.Dates[cur]; ok {
				persDays++
			}
		}
	}
	return vacDays, persDays
}

func daysInMonth(year int, month time.Month) int {
	switch month {
	case time.January, time.March, time.May, time.July, time.August, time.October, time.December:
		return 31
	case time.April, time.June, time.September, time.November:
		return 30
	case time.February:
		if isLeapYear(year) {
			return 29
		}
		return 28
	default:
		return 0
	}
}

func isLeapYear(y int) bool {
	if y%400 == 0 {
		return true
	}
	if y%100 == 0 {
		return false
	}
	return y%4 == 0
}

func getDefaultMonthNames() map[int]string {
	return map[int]string{
		1:  "January",
		2:  "February",
		3:  "March",
		4:  "April",
		5:  "May",
		6:  "June",
		7:  "July",
		8:  "August",
		9:  "September",
		10: "October",
		11: "November",
		12: "December",
	}
}

func getDefaultWeekdayNames() []string {
	return []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
}

func getDefaultWeekendDays() map[int]struct{} {
	return map[int]struct{}{
		5: {},
		6: {},
	}
}
