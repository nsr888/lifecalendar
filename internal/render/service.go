package render

import (
	"fmt"
	"sort"
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
)

// DateCategoryInfo holds category information for a specific date
type DateCategoryInfo struct {
	CategoryType entity.CategoryType
	Priority     int
	HasUnderline bool
}

// RenderService provides calendar rendering functionality with unified style management
type RenderService struct {
	year            int
	config          *entity.CategoryName
	appConfig       *config.Config
	ctx             *entity.RenderContext
	styleService    styles.StyleService
	maxWidthInChars int
	monthWidth      int
	separatorWidth  int
}

func NewRenderService(year int, cfg *entity.CategoryName, appConfig *config.Config, styleService styles.StyleService) *RenderService {
	rs := &RenderService{
		year:            year,
		config:          cfg,
		appConfig:       appConfig,
		ctx:             calendar.NewRenderContext(),
		styleService:    styleService,
		maxWidthInChars: 80, // default, will be overridden
		monthWidth:      20, // default, will be calculated
		separatorWidth:  2,  // 2 spaces between months
	}

	return rs
}

func (rs *RenderService) SetMaxWidth(maxWidth int) {
	if maxWidth < 20 {
		maxWidth = 20 // minimum readable width
	}
	rs.maxWidthInChars = maxWidth
	rs.monthWidth = 20 // maintain minimum month width
}

func (rs *RenderService) calculateColumnsPerWidth() int {
	if rs.maxWidthInChars < rs.monthWidth {
		return 1 // at least one column
	}

	columnsCalc := (rs.maxWidthInChars + rs.separatorWidth) / (rs.monthWidth + rs.separatorWidth)
	columns := min(max(columnsCalc, 1), 4) // between 1 and 4 (maximum 4 months per row)
	return columns
}

// calculateLayout determines if side panel is possible and returns layout info
func (rs *RenderService) calculateLayout() (useSidePanel bool, calendarCols int, sidePanelWidth int) {
	maxColsForCalendar := 4

	// Calculate how many calendar columns would fit in full width
	fullWidthCols := rs.calculateColumnsPerWidth()

	// Use the lesser of full width calculation and 4
	calendarCols = min(fullWidthCols, maxColsForCalendar)

	// Calculate width needed for calendar columns
	if calendarCols > 0 {
		calendarWidth := calendarCols*rs.monthWidth + (calendarCols-1)*rs.separatorWidth
		availableRightSpace := rs.maxWidthInChars - calendarWidth

		// Use side panel if we have at least 40 characters available
		useSidePanel = availableRightSpace >= 40
		if useSidePanel {
			sidePanelWidth = availableRightSpace
		}
	}

	return useSidePanel, calendarCols, sidePanelWidth
}

func (rs *RenderService) RenderYearTitle(year int) {
	borderString := strings.Repeat("─", rs.maxWidthInChars)
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#595959"))
	fmt.Println(borderStyle.Render(borderString))

	title := fmt.Sprintf("%d", year)
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#d8d8d8")).
		Bold(true).
		// Italic(true).
		Width(rs.maxWidthInChars).
		AlignHorizontal(lipgloss.Center)
	fmt.Println(titleStyle.Render(title))

	fmt.Println(borderStyle.Render(borderString))
}

func (rs *RenderService) computeMonthBlocks() [][]string {
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

func (rs *RenderService) getDayDisplay(dayDate time.Time) string {
	if dayDate.Year() != rs.year {
		return "  "
	}

	dayNum := fmt.Sprintf("%d", dayDate.Day())

	// Default style for regular days
	style := lipgloss.NewStyle().
		Width(2).
		Align(lipgloss.Right).
		Foreground(lipgloss.Color("#999999"))

	if info, exists := rs.styleService.GetDayStyle(dayDate); exists {
		categoryStyle := rs.styleService.GetCategoryStyle(info.Category)

		// Start with category style as base, then apply our constraints
		style = categoryStyle.
			Width(2).
			Align(lipgloss.Right)

		// Add underline if needed (for plans)
		// if info.HasUnderline {
		// 	style = style.Italic(true)
		// }

		return style.Render(dayNum)
	}

	return style.Render(dayNum)
}

// generateMonthLines creates the text lines for a single month
func (rs *RenderService) generateMonthLines(name string, calData [][]int, month time.Month) []string {
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
		for dow, dayNum := range week {
			if dayNum == 0 {
				cells = append(cells, "  ")
				continue
			}
			d := time.Date(rs.year, month, dayNum, 0, 0, 0, 0, time.Local)
			cells = append(cells, rs.getDayDisplay(d))
			_ = dow // ensure we maintain 7 cols
		}
		line := strings.Join(cells, " ")
		lines = append(lines, line)
	}
	return lines
}

// printMonthRow prints a row of months side by side
func (rs *RenderService) printMonthRow(rowMonths [][]string) string {
	var content strings.Builder

	maxLines := 0
	for _, mLines := range rowMonths {
		if len(mLines) > maxLines {
			maxLines = len(mLines)
		}
	}

	for li := 0; li < maxLines; li++ {
		var parts []string
		for _, mLines := range rowMonths {
			if li < len(mLines) {
				parts = append(parts, fmt.Sprintf("%-*s", rs.monthWidth, mLines[li]))
			} else {
				parts = append(parts, strings.Repeat(" ", rs.monthWidth))
			}
		}
		separator := strings.Repeat(" ", rs.separatorWidth)
		content.WriteString(strings.Join(parts, separator) + "\n")
	}

	return content.String()
}

// printLegend prints the color legend for all categories with background colors
func (rs *RenderService) printLegend() {
	// Collect all categories with background colors
	type legendItem struct {
		name  string
		style lipgloss.Style
	}

	var legendItems []legendItem

	// Iterate through all categories in the data
	for categoryName, category := range rs.config.Categories {
		// Skip weekends (typically just foreground color) and categories with no dates
		if len(category.Dates) == 0 {
			continue
		}

		style := rs.styleService.GetCategoryStyle(categoryName)

		// Only include categories with background colors (check if style is different from default)
		defaultStyle := lipgloss.NewStyle()
		if style.GetBackground() != defaultStyle.GetBackground() {
			// Format the category name for display (capitalize first letter, replace underscores with spaces)
			displayName := strings.ReplaceAll(categoryName, "_", " ")

			legendItems = append(legendItems, legendItem{
				name:  displayName,
				style: style,
			})
		}
	}

	// If no categories with background colors found, return
	if len(legendItems) == 0 {
		return
	}

	// Sort legend items by name for consistent ordering
	sort.Slice(legendItems, func(i, j int) bool {
		return legendItems[i].name < legendItems[j].name
	})

	// Find the maximum item length for formatting
	maxItemLength := 0
	for _, item := range legendItems {
		if len(item.name) > maxItemLength {
			maxItemLength = len(item.name)
		}
	}

	// Calculate how many items fit per row
	itemWidth := maxItemLength + 5
	columns := (rs.maxWidthInChars + 2) / (itemWidth + 2)
	if columns < 1 {
		columns = 1
	}

	// Print legend in rows
	for i := 0; i < len(legendItems); i += columns {
		end := min(i+columns, len(legendItems))

		var rowItems []string
		for j := i; j < end; j++ {
			item := fmt.Sprintf("%s %s",
				legendItems[j].style.Render("  "),
				lipgloss.NewStyle().Render(legendItems[j].name))
			rowItems = append(rowItems, item)
		}

		fmt.Println(strings.Join(rowItems, "  "))
	}
}

// centerString returns s centered to width w (truncate if too long)
func centerString(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	left := (w - len(s)) / 2
	right := w - len(s) - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// countWorkingDays counts days in a date range, excluding weekends and public holidays
func (rs *RenderService) countWorkingDays(start, end time.Time) int {
	count := 0
	for cur := start; !cur.After(end); cur = cur.AddDate(0, 0, 1) {
		// Skip weekends
		if cur.Weekday() == time.Saturday || cur.Weekday() == time.Sunday {
			continue
		}

		// Skip public holidays
		if holidayCategory, exists := rs.config.Categories["public_holidays"]; exists {
			if _, isHoliday := holidayCategory.Dates[cur]; isHoliday {
				continue
			}
		}

		count++
	}
	return count
}

// calculateVacationDaysUsed calculates total vacation days used, excluding weekends and holidays
func (rs *RenderService) calculateVacationDaysUsed() int {
	vacationCategory, exists := rs.config.Categories["vacations"]
	if !exists {
		return 0
	}

	// Get all vacation dates and sort them
	var vacationDates []time.Time
	for date := range vacationCategory.Dates {
		if date.Year() == rs.year {
			vacationDates = append(vacationDates, date)
		}
	}

	if len(vacationDates) == 0 {
		return 0
	}

	// Sort dates
	sort.Slice(vacationDates, func(i, j int) bool {
		return vacationDates[i].Before(vacationDates[j])
	})

	// Group dates into continuous ranges and count working days
	totalDays := 0
	i := 0
	for i < len(vacationDates) {
		rangeStart := vacationDates[i]
		rangeEnd := rangeStart

		// Find continuous range
		for i+1 < len(vacationDates) {
			nextDate := vacationDates[i+1]
			if nextDate.Equal(rangeEnd.AddDate(0, 0, 1)) {
				rangeEnd = nextDate
				i++
			} else {
				break
			}
		}

		totalDays += rs.countWorkingDays(rangeStart, rangeEnd)
		i++
	}

	return totalDays
}

// calculatePersonalDaysUsed calculates total personal days used, excluding weekends and holidays
func (rs *RenderService) calculatePersonalDaysUsed() int {
	personalCategory, exists := rs.config.Categories["personal_days"]
	if !exists {
		return 0
	}

	totalDays := 0
	for date := range personalCategory.Dates {
		if date.Year() == rs.year {
			totalDays += rs.countWorkingDays(date, date)
		}
	}

	return totalDays
}

// RenderCategoriesLabels prints categories with labeled entries respecting max width
func (rs *RenderService) RenderCategoriesLabels(labeledCategories []storage.LabeledCategory) {
	if len(labeledCategories) == 0 {
		return
	}

	for _, category := range labeledCategories {
		colors.PrintHeader(category.Name)

		l := list.New().
			Enumerator(list.Dash).
			ItemStyle(lipgloss.NewStyle().Width(rs.maxWidthInChars - 4))

		for _, entry := range category.Entries {
			startStr := entry.DateStart.Format("02.01")
			endStr := entry.DateEnd.Format("02.01")

			// Calculate total days (including weekends)
			totalDays := int(entry.DateEnd.Sub(entry.DateStart).Hours()/24) + 1
			var daysText string
			if totalDays == 1 {
				daysText = "1 day"
			} else {
				daysText = fmt.Sprintf("%d days", totalDays)
			}

			// Create prefix with date range and days count
			dateRangeText := fmt.Sprintf("%s-%s", startStr, endStr)
			techInfo := fmt.Sprintf("%s (%s)", dateRangeText, daysText)
			techInfoStyle := colors.Text().Bold(true)
			techInfo = techInfoStyle.Render(techInfo)
			entryDescStyle := colors.Text()
			entryDesc := entryDescStyle.Render(entry.Label)
			l.Item(fmt.Sprintf("%s - %s", techInfo, entryDesc))
		}
		fmt.Println(l)
	}
}

// RenderStats prints vacation and personal day usage statistics respecting max width
func (rs *RenderService) RenderStats() {
	colors.PrintHeader("Statistics:")
	vacationDays := rs.calculateVacationDaysUsed()
	personalDays := rs.calculatePersonalDaysUsed()

	vacationLine := fmt.Sprintf("Vacation days used: %d", vacationDays)
	personalLine := fmt.Sprintf("Personal days used: %d", personalDays)

	colors.PrintText(vacationLine)
	colors.PrintText(personalLine)
}

// generateLegendLines creates legend lines for side panel
func (rs *RenderService) generateLegendLines() string {
	var lines strings.Builder

	// Collect all categories with background colors
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

// wrapText wraps text to specified width with proper indentation
func (rs *RenderService) wrapText(text string, width int, indent string) []string {
	var lines []string

	for len(text) > width {
		// Find the best break point
		breakPoint := width
		for i := width - 1; i > width/3 && i > 0; i-- { // Search within last 1/3 of width
			if text[i] == ' ' {
				breakPoint = i
				break
			}
		}

		// If no space found, try dash
		if breakPoint == width {
			for i := width - 1; i > width/3 && i > 0; i-- {
				if text[i] == '-' {
					breakPoint = i
					break
				}
			}
		}

		lines = append(lines, text[:breakPoint])
		text = indent + text[breakPoint:]
		if len(text) > 0 && text[0] == ' ' {
			text = text[1:]
		}
	}

	if len(text) > 0 {
		lines = append(lines, text)
	}

	return lines
}

// generateCategoriesLines creates category lines for side panel
func (rs *RenderService) generateCategoriesLines(labeledCategories []storage.LabeledCategory, width int) string {
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
			startStr := entry.DateStart.Format("02.01")
			endStr := entry.DateEnd.Format("02.01")

			// Calculate total days
			totalDays := int(entry.DateEnd.Sub(entry.DateStart).Hours()/24) + 1
			var daysText string
			if totalDays == 1 {
				daysText = "1 day"
			} else {
				daysText = fmt.Sprintf("%d days", totalDays)
			}

			dateRangeText := fmt.Sprintf("%s-%s", startStr, endStr)
			techInfo := fmt.Sprintf("%s (%s)", dateRangeText, daysText)
			entryLine := fmt.Sprintf("%s - %s", techInfo, entry.Label)
			l.Item(entryLine)
		}

		lines.WriteString(l.String() + "\n")
	}

	return lines.String()
}

// calculateCategoryStats calculates statistics for all categories, considering priority
// Each day is counted only for the highest priority category it belongs to
func (rs *RenderService) calculateCategoryStats() map[string]int {
	stats := make(map[string]int)

	// Get all dates in the year
	startDate := time.Date(rs.year, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(rs.year, 12, 31, 0, 0, 0, 0, time.Local)

	for currentDate := startDate; !currentDate.After(endDate); currentDate = currentDate.AddDate(0, 0, 1) {
		// Find the highest priority category for this date
		var winningCategory string
		var winningPriority int = 999 // Start with high number (low priority)

		for categoryName, category := range rs.config.Categories {
			if _, exists := category.Dates[currentDate]; exists {
				categoryConfig := rs.appConfig.GetCategoryConfig(categoryName)
				if categoryConfig.Priority < winningPriority {
					winningCategory = categoryName
					winningPriority = categoryConfig.Priority
				}
			}
		}

		// If we found a category, count this day for it
		if winningCategory != "" {
			stats[winningCategory]++
		}
	}

	return stats
}

func (rs *RenderService) generateStatsLines(width int) string {
	var lines strings.Builder

	lines.WriteString(colors.Header().Render("Statistics:") + "\n")

	// Calculate statistics automatically for all categories
	categoryStats := rs.calculateCategoryStats()

	l := list.New().
		Enumerator(list.Bullet).
		EnumeratorStyle(colors.Text().MarginRight(1)).
		ItemStyle(colors.Text().Width(width - 4))

	// Sort categories by priority for consistent display
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

	// Sort by priority (ascending) and then by name
	sort.Slice(sortedStats, func(i, j int) bool {
		if sortedStats[i].priority == sortedStats[j].priority {
			return sortedStats[i].name < sortedStats[j].name
		}
		return sortedStats[i].priority < sortedStats[j].priority
	})

	// Display statistics
	for _, stat := range sortedStats {
		if stat.days > 0 { // Only show categories that have days
			// Format category name for display (replace underscores with spaces and capitalize)
			displayName := strings.ReplaceAll(stat.name, "_", " ")
			displayName = strings.Title(displayName)
			line := fmt.Sprintf("%s: %d", displayName, stat.days)
			l.Item(line)
		}
	}

	lines.WriteString(l.String() + "\n")

	return lines.String()
}

// RenderCompactYearView renders the calendar in a compact grid with configurable columns
func (rs *RenderService) RenderCompactYearView() {
	rs.RenderCompactYearViewWithSidePanel(nil)
}

func (rs *RenderService) RenderCompactYearViewWithSidePanel(labeledCategories []storage.LabeledCategory) {
	allMonths := rs.computeMonthBlocks()

	useSidePanel, calendarCols, sidePanelWidth := rs.calculateLayout()

	// Render with side panel if possible
	if useSidePanel && labeledCategories != nil {
		rs.renderTwoColumnLayout(allMonths, calendarCols, sidePanelWidth, labeledCategories)
		return
	}

	// Fallback to single column layout
	leftContent := rs.createLeftSidePanelContent(allMonths, calendarCols)
	rightContent := rs.createRightSidePanelContent(labeledCategories, rs.maxWidthInChars)
	mergedCols := lipgloss.JoinVertical(
		lipgloss.Left,
		leftContent,
		rightContent,
	)
	fmt.Println(mergedCols)
}

func (rs *RenderService) renderTwoColumnLayout(
	allMonths [][]string,
	calendarCols int,
	sidePanelWidth int,
	labeledCategories []storage.LabeledCategory,
) {
	paddingWidth := 4 // spaces between calendar and side panel
	leftContent := rs.createLeftSidePanelContent(allMonths, calendarCols)
	rightContent := rs.createRightSidePanelContent(labeledCategories, sidePanelWidth-paddingWidth)

	mergedCols := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftContent,
		strings.Repeat(" ", paddingWidth),
		rightContent,
	)

	fmt.Println(mergedCols)
}

// createLeftSidePanelContent creates content for left side column of side panel
func (rs *RenderService) createLeftSidePanelContent(
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

func (rs *RenderService) createRightSidePanelContent(
	labeledCategories []storage.LabeledCategory,
	width int,
) string {
	var content strings.Builder

	// legend
	legend := rs.generateLegendLines()
	legend2 := lipgloss.NewStyle().Width(width).Render(legend)
	content.WriteString(legend2)

	// categories
	if len(labeledCategories) > 0 {
		content.WriteString("\n") // Add spacing
		categoriesLines := rs.generateCategoriesLines(labeledCategories, width)
		content.WriteString(categoriesLines)
	}

	// statistics
	content.WriteString(rs.generateStatsLines(width))

	return content.String()
}
