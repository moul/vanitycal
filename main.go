package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	ical "github.com/arran4/golang-ical"
)

type Event struct {
	Date        string `toml:"date"`        // YYYY-MM-DD for anniversaries or countdowns
	MonthDay    string `toml:"month_day"`   // MM-DD for recurring annual events
	Title       string `toml:"title"`
	Description string `toml:"description"`
}

type Anniversary struct {
	Years  []int `toml:"years"`
	Months []int `toml:"months"`
	Days   []int `toml:"days"`
}

type Config struct {
	Timezone      string      `toml:"timezone"`
	CalendarName  string      `toml:"calendar_name"`
	Anniversaries Anniversary `toml:"anniversaries"`
	Events        []Event     `toml:"events"`
}

func main() {
	configFile := flag.String("config", "-", "Path to the config file (use '-' for stdin)")
	outputFile := flag.String("output", "-", "Path to the output file (use '-' for stdout)")
	flag.Parse()

	if err := run(*configFile, *outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(configFile, outputFile string) error {
	config, err := loadConfig(configFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	output, cleanup, err := createOutput(outputFile)
	if err != nil {
		return fmt.Errorf("creating output: %w", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	if err := generateICal(config, output); err != nil {
		return fmt.Errorf("generating calendar: %w", err)
	}

	return nil
}

func loadConfig(configFile string) (Config, error) {
	var config Config
	var err error

	if configFile == "-" {
		_, err = toml.NewDecoder(os.Stdin).Decode(&config)
	} else {
		_, err = toml.DecodeFile(configFile, &config)
	}

	return config, err
}

func validateConfig(config Config) error {
	if len(config.Events) == 0 {
		return fmt.Errorf("no events found in configuration")
	}

	for i, event := range config.Events {
		if event.Title == "" {
			return fmt.Errorf("event %d: title is required", i+1)
		}
		
		// Check that exactly one of date or month_day is specified
		if event.Date == "" && event.MonthDay == "" {
			return fmt.Errorf("event %d: either date or month_day is required", i+1)
		}
		if event.Date != "" && event.MonthDay != "" {
			return fmt.Errorf("event %d: cannot specify both date and month_day", i+1)
		}
		
		// Validate date format
		if event.Date != "" {
			if _, err := time.Parse("2006-01-02", event.Date); err != nil {
				return fmt.Errorf("event %d: invalid date format '%s' (expected YYYY-MM-DD)", i+1, event.Date)
			}
		}
		
		// Validate month_day format
		if event.MonthDay != "" {
			if _, err := time.Parse("01-02", event.MonthDay); err != nil {
				return fmt.Errorf("event %d: invalid month_day format '%s' (expected MM-DD)", i+1, event.MonthDay)
			}
		}
		
	}

	return nil
}

func createOutput(outputFile string) (io.Writer, func(), error) {
	if outputFile == "-" {
		return os.Stdout, nil, nil
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return nil, nil, err
	}

	return file, func() { file.Close() }, nil
}

func generateICal(config Config, output io.Writer) error {
	// Apply defaults
	config = applyDefaults(config)

	cal := ical.NewCalendar()
	cal.SetMethod(ical.MethodPublish)
	cal.SetName(config.CalendarName)
	cal.SetDescription("")
	cal.SetTimezoneId(config.Timezone)
	cal.SetTzid(config.Timezone)
	cal.SetCalscale("GREGORIAN")
	cal.SetLastModified(time.Now()) // XXX: take last modification date of this binary AND the input.

	currentYear := time.Now().Year()
	
	for _, event := range config.Events {
		if event.Date != "" {
			// Handle anniversary events (with full date)
			date, err := time.Parse("2006-01-02", event.Date)
			if err != nil {
				return fmt.Errorf("Error parsing date: %w", err)
			}
			// For future dates, generate BOTH countdown and anniversary events
			if date.After(time.Now()) {
				// First, generate countdown events
				countdowns := getCountdowns(date, config.Anniversaries)
				for _, countdown := range countdowns {
					duration := getCountdownDuration(countdown, date)
					uuid := fmt.Sprintf("vanitycal-countdown-%s", countdown.Format("20060102"))
					icalEvent := cal.AddEvent(uuid)
					summary := fmt.Sprintf("%s - %s 💚", event.Title, duration)
					icalEvent.SetSummary(summary)
					if event.Description != "" {
						icalEvent.SetDescription(event.Description)
					}

					// fullday
					icalEvent.SetProperty(ical.ComponentPropertyDtStart, countdown.UTC().Format("20060102"), ical.WithValue("DATE"))
				}
			}
			
			// Always generate anniversary events (for both past and future dates)
			anniversaries := getAnniversaries(date, config.Anniversaries)
			for _, anniv := range anniversaries {
				duration := getDuration(date, anniv)
				uuid := fmt.Sprintf("vanitycal-%s", anniv.Format("20060102"))
				icalEvent := cal.AddEvent(uuid)
				summary := fmt.Sprintf("%s - %s 💚", event.Title, duration)
				icalEvent.SetSummary(summary)
				if event.Description != "" {
					icalEvent.SetDescription(event.Description)
				}

				// fullday
				icalEvent.SetProperty(ical.ComponentPropertyDtStart, anniv.UTC().Format("20060102"), ical.WithValue("DATE"))

				// XXX: specific hours
				//icalEvent.SetStartAt(anniv)
				//icalEvent.SetEndAt(anniv.Add(24 * time.Hour))
			}
		} else if event.MonthDay != "" {
			// Handle recurring annual events (month-day only)
			// Generate for previous, current, and next year
			monthDay, _ := time.Parse("01-02", event.MonthDay)
			month := monthDay.Month()
			day := monthDay.Day()
			
			for yearOffset := -1; yearOffset <= 1; yearOffset++ {
				year := currentYear + yearOffset
				eventDate := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
				
				uuid := fmt.Sprintf("vanitycal-recurring-%d%02d%02d", year, month, day)
				icalEvent := cal.AddEvent(uuid)
				summary := fmt.Sprintf("%s 💚", event.Title)
				icalEvent.SetSummary(summary)
				if event.Description != "" {
					icalEvent.SetDescription(event.Description)
				}
				
				// fullday
				icalEvent.SetProperty(ical.ComponentPropertyDtStart, eventDate.Format("20060102"), ical.WithValue("DATE"))
			}
		}
	}

	_, err := output.Write([]byte(cal.Serialize()))
	return err
}

func applyDefaults(config Config) Config {
	if config.Timezone == "" {
		config.Timezone = "Europe/Paris"
	}
	if config.CalendarName == "" {
		config.CalendarName = "VanityCal 💚"
	}

	// Apply default anniversary patterns if not specified
	if len(config.Anniversaries.Years) == 0 {
		config.Anniversaries.Years = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 15, 20, 25, 30, 35, 40, 45, 50}
	}
	if len(config.Anniversaries.Months) == 0 {
		config.Anniversaries.Months = []int{1, 2, 3, 6, 9}
	}
	if len(config.Anniversaries.Days) == 0 {
		config.Anniversaries.Days = []int{0, 7, 100, 1000, 10000} // 0 means D-Day
	}

	return config
}

func getAnniversaries(date time.Time, patterns Anniversary) []time.Time {
	var anniversaries []time.Time

	// Add day-based anniversaries
	for _, days := range patterns.Days {
		if days == 0 {
			anniversaries = append(anniversaries, date) // D-Day
		} else {
			anniversaries = append(anniversaries, date.AddDate(0, 0, days))
		}
	}

	// Add month-based anniversaries
	for _, months := range patterns.Months {
		anniversaries = append(anniversaries, date.AddDate(0, months, 0))
	}

	// Add year-based anniversaries
	for _, years := range patterns.Years {
		anniversaries = append(anniversaries, date.AddDate(years, 0, 0))
	}

	return anniversaries
}

func getCountdowns(targetDate time.Time, patterns Anniversary) []time.Time {
	var countdowns []time.Time
	today := time.Now()

	// Only generate countdowns for future dates
	if targetDate.Before(today) || targetDate.Equal(today) {
		return countdowns
	}

	// Add day-based countdowns
	for _, days := range patterns.Days {
		if days == 0 {
			continue // Skip D-Day for countdowns
		}
		countdown := targetDate.AddDate(0, 0, -days)
		if countdown.After(today) {
			countdowns = append(countdowns, countdown)
		}
	}

	// Add month-based countdowns
	for _, months := range patterns.Months {
		countdown := targetDate.AddDate(0, -months, 0)
		if countdown.After(today) {
			countdowns = append(countdowns, countdown)
		}
	}

	// Add year-based countdowns
	for _, years := range patterns.Years {
		countdown := targetDate.AddDate(-years, 0, 0)
		if countdown.After(today) {
			countdowns = append(countdowns, countdown)
		}
	}

	return countdowns
}

func getCountdownDuration(from, to time.Time) string {
	if from.Equal(to) || from.After(to) {
		return "D-DAY"
	}

	// For countdowns, we show time remaining (from is before to)
	totalDays := int(to.Sub(from).Hours() / 24)

	// Check for specific day milestones first
	switch totalDays {
	case 1:
		return "D-1"
	case 2:
		return "D-2"
	case 3:
		return "D-3" 
	case 5:
		return "D-5"
	case 7:
		return "D-7"
	case 10:
		return "D-10"
	case 30:
		return "D-30"
	case 60:
		return "D-60"
	case 90:
		return "D-90"
	case 100:
		return "D-100"
	case 365:
		return "D-365"
	case 1000:
		return "D-1000"
	}

	// Check for exact month/year matches
	years := to.Year() - from.Year()
	months := int(to.Month() - from.Month())
	days := to.Day() - from.Day()

	// Adjust for negative months
	if months < 0 {
		years--
		months += 12
	}

	// Adjust for negative days
	if days < 0 {
		months--
		if months < 0 {
			years--
			months += 12
		}
		prevMonth := to.AddDate(0, -1, 0)
		days += time.Date(prevMonth.Year(), prevMonth.Month()+1, 0, 0, 0, 0, 0, prevMonth.Location()).Day()
	}

	// Return formatted countdown
	if years > 0 && months == 0 && days == 0 {
		return fmt.Sprintf("D-%dy", years)
	} else if years == 0 && months > 0 && days == 0 {
		return fmt.Sprintf("D-%dm", months)
	} else if totalDays < 30 {
		return fmt.Sprintf("D-%d", totalDays)
	} else {
		// Calculate total months for cleaner display
		totalMonths := months + years*12
		if totalMonths > 0 && totalMonths < 12 {
			return fmt.Sprintf("D-%dm", totalMonths)
		} else if totalMonths >= 12 {
			return fmt.Sprintf("D-%dy", totalMonths/12)
		} else {
			return fmt.Sprintf("D-%d", totalDays)
		}
	}
}

func getDuration(start, end time.Time) string {
	if end.Equal(start) {
		return "D-DAY"
	}

	// Calculate total days
	totalDays := int(end.Sub(start).Hours() / 24)

	// Check for exact year matches (including leap year edge case)
	years := end.Year() - start.Year()
	if years > 0 {
		testDate := start.AddDate(years, 0, 0)
		if testDate.Equal(end) {
			return fmt.Sprintf("%dy", years)
		}
		// Special case for leap year Feb 29 -> Feb 28
		if start.Month() == 2 && start.Day() == 29 && end.Month() == 2 && end.Day() == 28 {
			if testDate.Month() == 3 && testDate.Day() == 1 {
				// AddDate moved us to March 1, but we want Feb 28
				return fmt.Sprintf("%dy", years)
			}
		}
	}

	// Check for exact month matches
	// Try different month counts to find exact matches
	for months := 1; months <= years*12+12; months++ {
		if start.AddDate(0, months, 0).Equal(end) {
			if months >= 12 {
				y := months / 12
				m := months % 12
				if m == 0 {
					return fmt.Sprintf("%dy", y)
				}
				return fmt.Sprintf("%dy %dm", y, m)
			}
			return fmt.Sprintf("%dm", months)
		}
	}

	// Check for specific day milestones
	switch totalDays {
	case 7:
		return "7d"
	case 100:
		return "100d"
	case 1000:
		return "1000d"
	case 10000:
		return "10000d"
	}

	// For other cases, calculate years, months, and days
	years = end.Year() - start.Year()
	months := int(end.Month() - start.Month())
	days := end.Day() - start.Day()

	// Adjust for negative months
	if months < 0 {
		years--
		months += 12
	}

	// Adjust for negative days
	if days < 0 {
		months--
		if months < 0 {
			years--
			months += 12
		}
		// Get the last day of the previous month
		prevMonth := end.AddDate(0, -1, 0)
		days += time.Date(prevMonth.Year(), prevMonth.Month()+1, 0, 0, 0, 0, 0, prevMonth.Location()).Day()
	}

	// Format the output based on what's non-zero
	if years > 0 && months == 0 && days == 0 {
		return fmt.Sprintf("%dy", years)
	} else if years > 0 && months > 0 && days == 0 {
		return fmt.Sprintf("%dy %dm", years, months)
	} else if years > 0 && days > 0 && months == 0 {
		return fmt.Sprintf("%dy %dd", years, days)
	} else if months > 0 && days == 0 {
		return fmt.Sprintf("%dm", months)
	} else if months > 0 && days > 0 && years == 0 {
		return fmt.Sprintf("%dm %dd", months, days)
	} else if days > 0 {
		return fmt.Sprintf("%dd", days)
	}

	// Fallback for any edge case
	return fmt.Sprintf("%dy %dm %dd", years, months, days)
}
