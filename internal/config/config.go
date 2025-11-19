package config

import (
	"crypto/md5"
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/configor"
	"golang.org/x/term"
)

const defaultConfigPath = "config.toml"

type ColorStyle struct {
	Fg     string `toml:"fg"`
	Bg     string `toml:"bg"`
	Bold   bool   `toml:"bold"`
	Italic bool   `toml:"italic"`
}

type CategoryConfig struct {
	ColorStyle
	Priority int `toml:"priority"`
}

type Config struct {
	Years     []int `toml:"years"`
	DataFolder string `toml:"data_folder"`
	Rendering struct {
		MaxWidthInChars int   `toml:"max_width_in_chars"`
		FirstWeekday    int   `toml:"first_weekday"`
		WeekendDays     []int `toml:"weekend_days"`
	} `toml:"rendering"`
	Categories map[string]CategoryConfig `toml:"categories"`
}

func Load(configPath string) (*Config, error) {
	config := &Config{}
	currentYear := time.Now().Year()
	config.Years = []int{currentYear}
	config.DataFolder = "data" // Default data folder
	config.Rendering.MaxWidthInChars = getTerminalWidth()
	config.Rendering.FirstWeekday = 0          // Monday
	config.Rendering.WeekendDays = []int{5, 6} // Saturday, Sunday

	config.Categories = make(map[string]CategoryConfig)

	if _, err := os.Stat(configPath); err == nil {
		err = configor.Load(config, configPath)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

func getTerminalWidth() int {
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		if width >= 20 {
			return width
		}
	}

	// Fallback for non-terminal environments or errors
	return 80
}

func generateColorFromHash(categoryName string) string {
	hash := md5.Sum([]byte(categoryName))

	// Generate hue from hash (0-360)
	hue := int(hash[0]) * 141 % 360 // 141 is a prime to distribute colors well

	// Convert HSL to RGB with fixed saturation and lightness for good visibility
	saturation := 0.7
	lightness := 0.4

	return hslToHex(hue, saturation, lightness)
}

func hslToHex(h int, s, l float64) string {
	c := (1.0 - abs(2*l-1)) * s
	x := c * (1.0 - abs(float64((h/60)%2)-1))
	m := l - c/2.0

	var r, g, b float64

	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	r = (r + m) * 255
	g = (g + m) * 255
	b = (b + m) * 255

	return fmt.Sprintf("#%02x%02x%02x", int(r), int(g), int(b))
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func (c *Config) GetCategoryConfig(categoryName string) CategoryConfig {
	if config, exists := c.Categories[categoryName]; exists {
		return config
	}

	bgColor := generateColorFromHash(categoryName)
	fgColor := "#ffffff"

	return CategoryConfig{
		ColorStyle: ColorStyle{
			Bg: bgColor,
			Fg: fgColor,
		},
		Priority: 999, // Low priority for unknown categories
	}
}

func (c *Config) GetDataFolder() string {
	if c.DataFolder != "" {
		return c.DataFolder
	}
	return "data"
}

func (c *Config) GetDataFolderWithFallback() string {
	primaryFolder := c.GetDataFolder()

	if _, err := os.Stat(primaryFolder); err == nil {
		return primaryFolder
	}

	if primaryFolder != "data" {
		if _, err := os.Stat("data"); err == nil {
			return "data"
		}
	}

	return primaryFolder
}

func LoadDefault() (*Config, error) {
	return Load(defaultConfigPath)
}
