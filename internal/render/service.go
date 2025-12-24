package render

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/nsr888/lifecalendar/internal/calendar"
	"github.com/nsr888/lifecalendar/internal/config"
	"github.com/nsr888/lifecalendar/internal/entity"
	"github.com/nsr888/lifecalendar/internal/storage"
	"github.com/nsr888/lifecalendar/internal/styles"
	"github.com/nsr888/lifecalendar/pkg/colors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// DateCategoryInfo holds category information for a specific date.
type DateCategoryInfo struct {
	CategoryType entity.CategoryType
	Priority     int
	HasUnderline bool
}

// Service provides calendar rendering functionality with unified style management.
type Service struct {
	year            int
	config          *entity.CategoryName
	appConfig       *config.Config
	ctx             *entity.RenderContext
	styleService    styles.StyleService
	maxWidthInChars int
	monthWidth      int
	separatorWidth  int
}

func NewService(
	year int,
	cfg *entity.CategoryName,
	appConfig *config.Config,
	styleService styles.StyleService,
) *Service {
	rs := &Service{
		year:            year,
		config:          cfg,
		appConfig:       appConfig,
		ctx:             calendar.NewRenderContext(),
		styleService:    styleService,
		maxWidthInChars: 80,
		monthWidth:      20,
		separatorWidth:  2,
	}

	return rs
}

func (rs *Service) SetMaxWidth(maxWidth int) {
	if maxWidth < 20 {
		maxWidth = 20
	}
	rs.maxWidthInChars = maxWidth
	rs.monthWidth = 20
}

func (rs *Service) calculateColumnsPerWidth() int {
	if rs.maxWidthInChars < rs.monthWidth {
		return 1
	}

	columnsCalc := (rs.maxWidthInChars + rs.separatorWidth) / (rs.monthWidth + rs.separatorWidth)
	columns := min(max(columnsCalc, 1), 4)
	return columns
}

// calculateLayout determines if side panel is possible and returns layout info.
func (rs *Service) calculateLayout() (bool, int, int) {
	maxColsForCalendar := 4
	fullWidthCols := rs.calculateColumnsPerWidth()

	calendarCols := min(fullWidthCols, maxColsForCalendar)

	var useSidePanel bool
	var sidePanelWidth int
	if calendarCols > 0 {
		calendarWidth := calendarCols*rs.monthWidth + (calendarCols-1)*rs.separatorWidth
		availableRightSpace := rs.maxWidthInChars - calendarWidth

		useSidePanel = availableRightSpace >= 40
		if useSidePanel {
			sidePanelWidth = availableRightSpace
		}
	}

	return useSidePanel, calendarCols, sidePanelWidth
}

func (rs *Service) RenderYearTitle(year int) {
	borderString := strings.Repeat("─", rs.maxWidthInChars)
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#595959"))
	fmt.Println(borderStyle.Render(borderString))

	title := strconv.Itoa(year)
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#d8d8d8")).
		Bold(true).
		Width(rs.maxWidthInChars).
		AlignHorizontal(lipgloss.Center)
	fmt.Println(titleStyle.Render(title))

	fmt.Println(borderStyle.Render(borderString))
}

func (rs *Service) computeMonthBlocks() [][]string {
	allMonths := make([][]string, 12)
	for m := 1; m <= 12; m++ {
		name := rs.ctx.MonthNames[m]
		if name == "" {
			name = time.Month(m).String()
		}

		calData := calendar.MonthCalendar(rs.year, time.Month(m))
		lines := rs.generateMonthLines(name, calData, time.Month(m))
		allMonths[m-1] = lines
	}
	return allMonths
}

func (rs *Service) getDayDisplay(dayDate time.Time) string {
	if dayDate.Year() != rs.year {
		return "  "
	}

	dayNum := strconv.Itoa(dayDate.Day())

	style := lipgloss.NewStyle().
		Width(2).
		Align(lipgloss.Right).
		Foreground(lipgloss.Color("#999999"))

	if info, exists := rs.styleService.GetDayStyle(dayDate); exists {
		categoryStyle := rs.styleService.GetCategoryStyle(info.Category)

		style = categoryStyle.
			Width(2).
			Align(lipgloss.Right)

		return style.Render(dayNum)
	}

	return style.Render(dayNum)
}

// generateMonthLines creates the text lines for a single month.
func (rs *Service) generateMonthLines(
	name string,
	calData [][]int,
	month time.Month,
) []string {
	var lines []string
	monthHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#909090")).
		Bold(true).
		Italic(true).
		Width(rs.monthWidth).
		AlignHorizontal(lipgloss.Center)
	header := monthHeaderStyle.Render(name)
	lines = append(lines, header)
	weekdayHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4d4d4d")).
		Bold(false)
	weekdayHeader := weekdayHeaderStyle.Render("Mo Tu We Th Fr Sa Su")
	lines = append(lines, weekdayHeader)

	for _, week := range calData {
		var cells []string
		for _, dayNum := range week {
			if dayNum == 0 {
				cells = append(cells, "  ")
				continue
			}
			d := time.Date(rs.year, month, dayNum, 0, 0, 0, 0, time.Local)
			cells = append(cells, rs.getDayDisplay(d))
		}
		line := strings.Join(cells, " ")
		lines = append(lines, line)
	}
	return lines
}

// printMonthRow prints a row of months side by side.
func (rs *Service) printMonthRow(rowMonths [][]string) string {
	var content strings.Builder

	maxLines := 0
	for _, mLines := range rowMonths {
		if len(mLines) > maxLines {
			maxLines = len(mLines)
		}
	}

	for li := range maxLines {
		var parts []string
		for _, mLines := range rowMonths {
			if li < len(mLines) {
				parts = append(
					parts,
					fmt.Sprintf("%-*s", rs.monthWidth, mLines[li]),
				)
			} else {
				parts = append(parts, strings.Repeat(" ", rs.monthWidth))
			}
		}
		separator := strings.Repeat(" ", rs.separatorWidth)
		content.WriteString(strings.Join(parts, separator) + "\n")
	}

	return content.String()
}

// generateLegendLines creates legend lines for side panel.
func (rs *Service) generateLegendLines() string {
	var lines strings.Builder

	type legendItem struct {
		name  string
		style lipgloss.Style
	}

	var legendItems []legendItem

	for categoryName, category := range rs.config.Categories {
		if len(category.Dates) == 0 {
			continue
		}

		style := rs.styleService.GetCategoryStyle(categoryName)

		noColor := lipgloss.NoColor{}
		if style.GetBackground() == noColor {
			continue
		}

		displayName := strings.ReplaceAll(categoryName, "_", " ")
		legendItems = append(legendItems, legendItem{
			name:  displayName,
			style: style,
		})
	}

	if len(legendItems) == 0 {
		return ""
	}

	sort.Slice(legendItems, func(i, j int) bool {
		return legendItems[i].name < legendItems[j].name
	})

	if len(legendItems) > 0 {
		lines.WriteString(colors.Header().Render("Legend:"))
		lines.WriteString("\n")
	}

	for _, item := range legendItems {
		line := fmt.Sprintf("%s %s  ",
			item.style.Render("  "),
			colors.Text().Render(item.name),
		)
		lines.WriteString(line)
	}

	return lines.String()
}

func (rs *Service) generateCategoriesLines(
	labeledCategories []storage.LabeledCategory,
	width int,
) string {
	var lines strings.Builder

	if len(labeledCategories) == 0 {
		return ""
	}

	for _, category := range labeledCategories {
		header := colors.Header().Render(category.Name + ":")
		lines.WriteString(header + "\n")

		l := list.New().
			Enumerator(list.Bullet).
			EnumeratorStyle(colors.Text().MarginRight(1)).
			ItemStyle(colors.Text().Width(width - 4))

		for _, entry := range category.Entries {
			l.Item(entry.String())
		}

		lines.WriteString(l.String() + "\n")
	}

	return lines.String()
}

// calculateCategoryStats calculates statistics for all categories, considering priority.
func (rs *Service) calculateCategoryStats() map[string]int {
	stats := make(map[string]int)

	startDate := time.Date(rs.year, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(rs.year, 12, 31, 0, 0, 0, 0, time.Local)

	for currentDate := startDate; !currentDate.After(endDate); currentDate = currentDate.AddDate(0, 0, 1) {
		var winningCategory string
		winningPriority := 999

		for categoryName, category := range rs.config.Categories {
			if _, exists := category.Dates[currentDate]; exists {
				categoryConfig := rs.appConfig.GetCategoryConfig(categoryName)
				if categoryConfig.Priority < winningPriority {
					winningCategory = categoryName
					winningPriority = categoryConfig.Priority
				}
			}
		}

		if winningCategory != "" {
			stats[winningCategory]++
		}
	}

	return stats
}

func (rs *Service) generateStatsLines(width int) string {
	var lines strings.Builder

	lines.WriteString(colors.Header().Render("Statistics:") + "\n")

	categoryStats := rs.calculateCategoryStats()

	l := list.New().
		Enumerator(list.Bullet).
		EnumeratorStyle(colors.Text().MarginRight(1)).
		ItemStyle(colors.Text().Width(width - 4))

	type categoryStat struct {
		name     string
		priority int
		days     int
	}

	var sortedStats []categoryStat
	for categoryName, days := range categoryStats {
		categoryConfig := rs.appConfig.GetCategoryConfig(categoryName)
		sortedStats = append(sortedStats, categoryStat{
			name:     categoryName,
			priority: categoryConfig.Priority,
			days:     days,
		})
	}

	sort.Slice(sortedStats, func(i, j int) bool {
		if sortedStats[i].priority == sortedStats[j].priority {
			return sortedStats[i].name < sortedStats[j].name
		}
		return sortedStats[i].priority < sortedStats[j].priority
	})

	for _, stat := range sortedStats {
		if stat.days > 0 {
			displayName := strings.ReplaceAll(stat.name, "_", " ")
			displayName = cases.Title(language.English).String(displayName)
			line := fmt.Sprintf("%s: %d", displayName, stat.days)
			l.Item(line)
		}
	}

	lines.WriteString(l.String() + "\n")

	return lines.String()
}

// RenderCompactYearView renders the calendar in a compact grid with configurable columns.
func (rs *Service) RenderCompactYearView() {
	rs.RenderCompactYearViewWithSidePanel(nil)
}

// RenderThreeColumnView renders the calendar in a 3-column format:
// Column 1: Month names aligned with 1st day
// Column 2: Continuous calendar without month gaps
// Column 3: Plans aligned to start dates
func (rs *Service) RenderThreeColumnView(
	labeledCategories []storage.LabeledCategory,
) {
	monthNamesColumn := rs.generateMonthNamesColumn()
	continuousCalendarColumn := rs.generateContinuousCalendarColumn()
	plansColumn := rs.generateCatColumnInVerticalLayout(labeledCategories)

	maxHeight := max(
		len(strings.Split(monthNamesColumn, "\n")),
		len(strings.Split(continuousCalendarColumn, "\n")),
		len(strings.Split(plansColumn, "\n")),
	)

	for i := range maxHeight {
		var monthLine, calendarLine, plansLine string

		monthLines := strings.Split(monthNamesColumn, "\n")
		calendarLines := strings.Split(continuousCalendarColumn, "\n")
		plansLines := strings.Split(plansColumn, "\n")

		if i < len(monthLines) {
			monthLine = monthLines[i]
		} else {
			monthLine = strings.Repeat(" ", 12)
		}

		if i < len(calendarLines) {
			calendarLine = calendarLines[i]
		} else {
			calendarLine = strings.Repeat(" ", 53)
		}

		if i < len(plansLines) {
			plansLine = plansLines[i]
		} else {
			plansLine = ""
		}

		fmt.Printf("%-12s  %-53s  %s\n", monthLine, calendarLine, plansLine)
	}

	fmt.Println()

	rs.renderLegendAndStatistics()
}

// generateMonthNamesColumn creates the first column with month names
func (rs *Service) generateMonthNamesColumn() string {
	var lines []string

	lines = append(lines, "")

	currentDate := time.Date(rs.year, 1, 1, 0, 0, 0, 0, time.Local)

	// Calculate starting weekday offset
	startWeekday := int(currentDate.Weekday())
	if rs.appConfig.Rendering.FirstWeekday == 0 { // Monday as first day
		startWeekday = (startWeekday - 1 + 7) % 7
	}

	// Add empty lines for initial offset to align with first week
	for i := 0; i < startWeekday; i++ {
		lines = append(lines, "")
	}

	// Add month names at the start of each month
	for month := 1; month <= 12; month++ {
		firstDayOfMonth := time.Date(
			rs.year,
			time.Month(month),
			1,
			0,
			0,
			0,
			0,
			time.Local,
		)

		// Calculate which week line this month starts on
		daysSinceStart := int(
			firstDayOfMonth.Sub(time.Date(rs.year, 1, 1, 0, 0, 0, 0, time.Local)).
				Hours() /
				24,
		)
		weekIndex := 1 + (startWeekday+daysSinceStart)/7 // +1 for header line

		// Ensure we have enough lines
		for len(lines) <= weekIndex {
			lines = append(lines, "")
		}

		// Place month name
		if lines[weekIndex] == "" {
			lines[weekIndex] = firstDayOfMonth.Month().String()
		}
	}

	// Fill remaining lines to match calendar height
	calendarLines := strings.Split(rs.generateContinuousCalendarColumn(), "\n")
	for len(lines) < len(calendarLines) {
		lines = append(lines, "")
	}

	return colors.Text().Render(strings.Join(lines, "\n"))
}

// generateContinuousCalendarColumn creates the continuous calendar without month breaks
func (rs *Service) generateContinuousCalendarColumn() string {
	var lines []string

	weekdayNames := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	if rs.appConfig.Rendering.FirstWeekday == 0 {
		weekdayNames = []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	}

	header := strings.Join(weekdayNames, " ")
	header = colors.Text().Render(header)
	lines = append(lines, header)

	// Start from first day of year
	currentDate := time.Date(rs.year, 1, 1, 0, 0, 0, 0, time.Local)

	// Calculate starting weekday offset
	startWeekday := int(currentDate.Weekday())
	if rs.appConfig.Rendering.FirstWeekday == 0 {
		startWeekday = (startWeekday - 1 + 7) % 7
	}

	// Create the first week with leading spaces
	var weekDays []string
	for i := 0; i < startWeekday; i++ {
		weekDays = append(weekDays, "  ")
	}

	// Generate all days of the year
	for currentDate.Year() == rs.year {
		dayStr := fmt.Sprintf("%2d", currentDate.Day())

		// Apply styling if the day has a category
		if info, exists := rs.styleService.GetDayStyle(currentDate); exists {
			categoryStyle := rs.styleService.GetCategoryStyle(info.Category)
			dayStr = categoryStyle.Render(dayStr)
		} else {
			dayStr = colors.Text().Render(dayStr)
		}

		weekDays = append(weekDays, dayStr)

		// When we have 7 days, complete the week
		if len(weekDays) == 7 {
			lines = append(lines, strings.Join(weekDays, " "))
			weekDays = []string{}
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	// Add the last week if it has remaining days
	if len(weekDays) > 0 {
		// Pad remaining days with spaces
		for len(weekDays) < 7 {
			weekDays = append(weekDays, "  ")
		}
		lines = append(lines, strings.Join(weekDays, " "))
	}

	return strings.Join(lines, "\n")
}

func (rs *Service) generateCatColumnInVerticalLayout(
	labeledCategories []storage.LabeledCategory,
) string {
	if len(labeledCategories) == 0 {
		return strings.Repeat(
			"\n",
			54,
		)
	}

	weekPlans := make(map[int][]string)

	for _, category := range labeledCategories {
		for _, entry := range category.Entries {
			_, weekNum := entry.DateStart.ISOWeek()
			planText := colors.Text().Render(entry.String())
			weekPlans[weekNum] = append(weekPlans[weekNum], planText)
		}
	}

	var lines []string

	lines = append(lines, "") // Header line

	maxWeeksInYear := 53

	for week := 1; week <= maxWeeksInYear; week++ {
		if plans, exists := weekPlans[week]; exists && len(plans) > 0 {
			for _, plan := range plans {
				lines = append(lines, plan)
			}
		} else {
			lines = append(lines, "")
		}
	}

	return strings.Join(lines, "\n")
}

// renderLegendAndStatistics adds legend and statistics at the bottom
func (rs *Service) renderLegendAndStatistics() {
	maxWidth := rs.maxWidthInChars

	legend := rs.generateLegendLines()
	fmt.Println(legend)

	stats := rs.generateStatsLines(maxWidth)
	fmt.Println(stats)
}

func (rs *Service) RenderCompactYearViewWithSidePanel(
	labeledCategories []storage.LabeledCategory,
) {
	allMonths := rs.computeMonthBlocks()

	useSidePanel, calendarCols, sidePanelWidth := rs.calculateLayout()

	if useSidePanel && labeledCategories != nil {
		rs.renderTwoColumnLayout(
			allMonths,
			calendarCols,
			sidePanelWidth,
			labeledCategories,
		)
		return
	}

	leftContent := rs.createLeftSidePanelContent(allMonths, calendarCols)
	rightContent := rs.createRightSidePanelContent(
		labeledCategories,
		rs.maxWidthInChars,
	)
	mergedCols := lipgloss.JoinVertical(
		lipgloss.Left,
		leftContent,
		rightContent,
	)
	fmt.Println(mergedCols)
}

func (rs *Service) renderTwoColumnLayout(
	allMonths [][]string,
	calendarCols int,
	sidePanelWidth int,
	labeledCategories []storage.LabeledCategory,
) {
	paddingWidth := 4
	leftContent := rs.createLeftSidePanelContent(allMonths, calendarCols)
	rightContent := rs.createRightSidePanelContent(
		labeledCategories,
		sidePanelWidth-paddingWidth,
	)

	mergedCols := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftContent,
		strings.Repeat(" ", paddingWidth),
		rightContent,
	)

	fmt.Println(mergedCols)
}

// createLeftSidePanelContent creates content for left side column of side panel.
func (rs *Service) createLeftSidePanelContent(
	allMonths [][]string,
	calendarCols int,
) string {
	var content strings.Builder

	for rowStart := 0; rowStart < 12; rowStart += calendarCols {
		end := min(rowStart+calendarCols, 12)
		s := rs.printMonthRow(allMonths[rowStart:end])
		content.WriteString(s)
	}

	return content.String()
}

func (rs *Service) createRightSidePanelContent(
	labeledCategories []storage.LabeledCategory,
	width int,
) string {
	var content strings.Builder

	legend := rs.generateLegendLines()
	legend2 := lipgloss.NewStyle().Width(width).Render(legend)
	content.WriteString(legend2)

	if len(labeledCategories) > 0 {
		content.WriteString("\n")
		categoriesLines := rs.generateCategoriesLines(labeledCategories, width)
		content.WriteString(categoriesLines)
	}

	content.WriteString(rs.generateStatsLines(width))

	return content.String()
}
