package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nsr888/lifecalendar/internal/entity"
)

const (
	dateLayout   = "2006-01-02"
	dateStartCol = "date_start"
	dateEndCol   = "date_end"
	dateCol      = "date"
	labelCol     = "label"
	descCol      = "desc"
)

type CSVStorage struct {
	dataFolder string
}

func NewCSVStorage(dataFolder string) *CSVStorage {
	return &CSVStorage{dataFolder: dataFolder}
}

func (s *CSVStorage) IsYearDataExists(year int) bool {
	dataDir := fmt.Sprintf("%s/%d", s.dataFolder, year)
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		return false
	}

	return true
}

// GetCategoryNames returns all category names for a given year (excluding weekends)
func (s *CSVStorage) GetCategoryNames(year int) ([]string, error) {
	dataDir := fmt.Sprintf("%s/%d", s.dataFolder, year)

	categoryFiles, err := s.getCategoryToFileMap(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to discover category files: %w", err)
	}

	var categoryNames []string
	for categoryName := range categoryFiles {
		categoryNames = append(categoryNames, categoryName)
	}

	return categoryNames, nil
}

type LabeledCategory struct {
	Name        string
	Description string
	Entries     []CategoryEntry
}

type CategoryEntry struct {
	DateStart time.Time
	DateEnd   time.Time
	Label     string
}

func (s *CSVStorage) LoadCategoryByYear(year int) (*entity.CategoryName, error) {
	dataDir := fmt.Sprintf("%s/%d", s.dataFolder, year)

	categoryFiles, err := s.getCategoryToFileMap(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to discover category files: %w", err)
	}

	categories := make(map[string]*entity.Category)

	for categoryName, filename := range categoryFiles {
		category, err := s.loadCategoryFromFile(filename, entity.CategoryType(categoryName))
		if err != nil {
			return nil, fmt.Errorf("failed to load category %s: %w", categoryName, err)
		}
		categories[categoryName] = category
	}

	return &entity.CategoryName{
		BaseYear:   year,
		Categories: categories,
	}, nil
}

// getCategoryToFileMap scans the data directory for CSV files and returns a map of category name -> filename
func (s *CSVStorage) getCategoryToFileMap(dataDir string) (map[string]string, error) {
	categoryFiles := make(map[string]string)

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".csv") {
			continue
		}

		categoryName := strings.TrimSuffix(entry.Name(), ".csv")
		filename := filepath.Join(dataDir, entry.Name())
		categoryFiles[categoryName] = filename
	}

	return categoryFiles, nil
}

func (s *CSVStorage) loadCategoryFromFile(
	filename string,
	categoryType entity.CategoryType,
) (*entity.Category, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	category := &entity.Category{
		Type:    categoryType,
		Desc:    string(categoryType),
		Dates:   make(map[time.Time]struct{}),
		Entries: []entity.CategoryEntry{},
	}

	if len(records) == 0 {
		return category, nil
	}

	headers := records[0]
	if len(headers) == 0 {
		return category, nil
	}

	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) == 0 || record[0] == "" {
			continue
		}

		entry, err := s.parseCSVRecord(record, headers)
		if err != nil {
			return nil, fmt.Errorf("failed to parse record on line %d: %w", i+1, err)
		}

		category.Entries = append(category.Entries, entry)

		for cur := entry.DateStart; !cur.After(entry.DateEnd); cur = cur.AddDate(0, 0, 1) {
			category.Dates[cur] = struct{}{}
		}
	}

	return category, nil
}

func (s *CSVStorage) parseCSVRecord(record, headers []string) (entity.CategoryEntry, error) {
	var entry entity.CategoryEntry

	// Map headers to record indices
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(header))] = i
	}

	if dateStartIdx, exists := headerMap[dateStartCol]; exists && dateStartIdx < len(record) {
		if date, err := time.ParseInLocation(dateLayout, strings.TrimSpace(record[dateStartIdx]), time.Local); err == nil {
			entry.DateStart = date
		}
	}

	if dateEndIdx, exists := headerMap[dateEndCol]; exists && dateEndIdx < len(record) {
		if date, err := time.ParseInLocation(dateLayout, strings.TrimSpace(record[dateEndIdx]), time.Local); err == nil {
			entry.DateEnd = date
		}
	}

	// TODO: move to constants headerMap keys
	if entry.DateEnd.IsZero() {
		if dateIdx, exists := headerMap[dateCol]; exists && dateIdx < len(record) {
			if date, err := time.ParseInLocation(dateLayout, strings.TrimSpace(record[dateIdx]), time.Local); err == nil {
				entry.DateStart = date
				entry.DateEnd = date
			}
		} else if !entry.DateStart.IsZero() {
			entry.DateEnd = entry.DateStart
		}
	}

	// Parse label (try "label", then "desc")
	if labelIdx, exists := headerMap[labelCol]; exists && labelIdx < len(record) {
		entry.Label = strings.TrimSpace(record[labelIdx])
	} else if descIdx, exists := headerMap[descCol]; exists && descIdx < len(record) {
		entry.Label = strings.TrimSpace(record[descIdx])
	}

	// Validate entry
	if entry.DateStart.IsZero() || entry.DateEnd.IsZero() {
		return entity.CategoryEntry{}, fmt.Errorf("invalid date range")
	}

	if entry.Label == "" {
		entry.Label = "Event"
	}

	return entry, nil
}

// LoadLabeledCategories loads all categories that have entries with non-empty labels
func (s *CSVStorage) LoadLabeledCategories(year int) ([]LabeledCategory, error) {
	data, err := s.LoadCategoryByYear(year)
	if err != nil {
		return nil, err
	}

	var labeledCategories []LabeledCategory

	// Iterate through all categories and collect those with labeled entries
	for categoryName, category := range data.Categories {
		var labeledEntries []CategoryEntry

		for _, entry := range category.Entries {
			if entry.Label != "" && entry.Label != "Event" {
				labeledEntries = append(labeledEntries, CategoryEntry{
					DateStart: entry.DateStart,
					DateEnd:   entry.DateEnd,
					Label:     entry.Label,
				})
			}
		}

		if len(labeledEntries) > 0 {
			displayName := strings.ReplaceAll(categoryName, "_", " ")

			labeledCategories = append(labeledCategories, LabeledCategory{
				Name:        displayName,
				Description: category.Desc,
				Entries:     labeledEntries,
			})
		}
	}

	sort.Slice(labeledCategories, func(i, j int) bool {
		return labeledCategories[i].Name < labeledCategories[j].Name
	})

	return labeledCategories, nil
}
