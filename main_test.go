package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestGetCountdownDuration(t *testing.T) {
	tests := []struct {
		name     string
		from     string
		to       string
		expected string
	}{
		{"D-Day", "2024-01-01", "2024-01-01", "D-DAY"},
		{"1 day before", "2024-01-01", "2024-01-02", "D-1"},
		{"2 days before", "2024-01-01", "2024-01-03", "D-2"},
		{"3 days before", "2024-01-01", "2024-01-04", "D-3"},
		{"5 days before", "2024-01-01", "2024-01-06", "D-5"},
		{"7 days before", "2024-01-01", "2024-01-08", "D-7"},
		{"10 days before", "2024-01-01", "2024-01-11", "D-10"},
		{"100 days before", "2024-01-01", "2024-04-10", "D-100"},
		{"1 month before", "2024-01-01", "2024-02-01", "D-1m"},
		{"1 year before", "2024-01-01", "2025-01-01", "D-1y"},
		{"Past date", "2024-01-02", "2024-01-01", "D-DAY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, _ := time.Parse("2006-01-02", tt.from)
			to, _ := time.Parse("2006-01-02", tt.to)
			result := getCountdownDuration(from, to)
			if result != tt.expected {
				t.Errorf("getCountdownDuration(%s, %s) = %s; want %s", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestGetDuration(t *testing.T) {
	tests := []struct {
		name     string
		start    string
		end      string
		expected string
	}{
		{"D-Day", "2023-01-01", "2023-01-01", "D-DAY"},
		{"7 days", "2023-01-01", "2023-01-08", "7d"},
		{"1 month", "2023-01-01", "2023-02-01", "1m"},
		{"1 year", "2023-01-01", "2024-01-01", "1y"},
		{"2 years 3 months", "2023-01-01", "2025-04-01", "2y 3m"},
		{"1 year 15 days", "2023-01-01", "2024-01-16", "1y 15d"},
		{"6 months 10 days", "2023-01-01", "2023-07-11", "6m 10d"},
		{"100 days", "2023-01-01", "2023-04-11", "100d"},
		{"Leap year handling", "2020-02-29", "2021-02-28", "1y"},
		{"Month boundary", "2023-01-31", "2023-02-28", "28d"},
		{"Year boundary", "2022-12-31", "2023-01-01", "1d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, _ := time.Parse("2006-01-02", tt.start)
			end, _ := time.Parse("2006-01-02", tt.end)
			result := getDuration(start, end)
			if result != tt.expected {
				t.Errorf("getDuration(%s, %s) = %s; want %s", tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Empty config",
			config:  Config{},
			wantErr: true,
			errMsg:  "no events found",
		},
		{
			name: "Missing title",
			config: Config{
				Events: []Event{{Date: "2023-01-01", Title: ""}},
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "Missing date",
			config: Config{
				Events: []Event{{Date: "", Title: "Test"}},
			},
			wantErr: true,
			errMsg:  "either date or month_day is required",
		},
		{
			name: "Invalid date format",
			config: Config{
				Events: []Event{{Date: "01/01/2023", Title: "Test"}},
			},
			wantErr: true,
			errMsg:  "invalid date format",
		},
		{
			name: "Valid config",
			config: Config{
				Events: []Event{{Date: "2023-01-01", Title: "Test Event"}},
			},
			wantErr: false,
		},
		{
			name: "Valid config with description",
			config: Config{
				Events: []Event{{Date: "2023-01-01", Title: "Test Event", Description: "A test"}},
			},
			wantErr: false,
		},
		{
			name: "Valid recurring event",
			config: Config{
				Events: []Event{{MonthDay: "12-25", Title: "Christmas"}},
			},
			wantErr: false,
		},
		{
			name: "Invalid month_day format",
			config: Config{
				Events: []Event{{MonthDay: "25/12", Title: "Christmas"}},
			},
			wantErr: true,
			errMsg:  "invalid month_day format",
		},
		{
			name: "Both date and month_day specified",
			config: Config{
				Events: []Event{{Date: "2023-01-01", MonthDay: "01-01", Title: "Test"}},
			},
			wantErr: true,
			errMsg:  "cannot specify both date and month_day",
		},
		{
			name: "Neither date nor month_day specified",
			config: Config{
				Events: []Event{{Title: "Test"}},
			},
			wantErr: true,
			errMsg:  "either date or month_day is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateConfig() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		check  func(Config) bool
	}{
		{
			name:   "Empty config gets defaults",
			config: Config{},
			check: func(c Config) bool {
				return c.Timezone == "Europe/Paris" &&
					c.CalendarName == "VanityCal 💚" &&
					len(c.Anniversaries.Years) > 0 &&
					len(c.Anniversaries.Months) > 0 &&
					len(c.Anniversaries.Days) > 0
			},
		},
		{
			name: "Custom timezone preserved",
			config: Config{
				Timezone: "America/New_York",
			},
			check: func(c Config) bool {
				return c.Timezone == "America/New_York"
			},
		},
		{
			name: "Custom calendar name preserved",
			config: Config{
				CalendarName: "My Calendar",
			},
			check: func(c Config) bool {
				return c.CalendarName == "My Calendar"
			},
		},
		{
			name: "Custom anniversaries preserved",
			config: Config{
				Anniversaries: Anniversary{
					Years:  []int{1, 5, 10},
					Months: []int{6},
					Days:   []int{100},
				},
			},
			check: func(c Config) bool {
				return len(c.Anniversaries.Years) == 3 &&
					len(c.Anniversaries.Months) == 1 &&
					len(c.Anniversaries.Days) == 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyDefaults(tt.config)
			if !tt.check(result) {
				t.Errorf("applyDefaults() failed check for %s", tt.name)
			}
		})
	}
}

func TestGetAnniversaries(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2023-01-01")
	patterns := Anniversary{
		Years:  []int{1, 2, 5},
		Months: []int{1, 6},
		Days:   []int{0, 7, 100},
	}

	anniversaries := getAnniversaries(date, patterns)

	// Check we have the right number of anniversaries
	expected := len(patterns.Years) + len(patterns.Months) + len(patterns.Days)
	if len(anniversaries) != expected {
		t.Errorf("getAnniversaries() returned %d anniversaries, want %d", len(anniversaries), expected)
	}

	// Check specific dates
	expectedDates := []string{
		"2023-01-01", // D-Day (days: 0)
		"2023-01-08", // 7 days
		"2023-04-11", // 100 days
		"2023-02-01", // 1 month
		"2023-07-01", // 6 months
		"2024-01-01", // 1 year
		"2025-01-01", // 2 years
		"2028-01-01", // 5 years
	}

	for _, expectedDate := range expectedDates {
		expected, _ := time.Parse("2006-01-02", expectedDate)
		found := false
		for _, anniv := range anniversaries {
			if anniv.Equal(expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected anniversary date %s not found", expectedDate)
		}
	}
}

func TestGenerateICal(t *testing.T) {
	t.Run("Anniversary events", func(t *testing.T) {
		config := Config{
			Timezone:     "UTC",
			CalendarName: "Test Calendar",
			Anniversaries: Anniversary{
				Years:  []int{1},
				Months: []int{1},
				Days:   []int{0},
			},
			Events: []Event{
				{
					Date:        "2023-01-01",
					Title:       "Test Event",
					Description: "Test Description",
				},
			},
		}

		var buf bytes.Buffer
		err := generateICal(config, &buf)
		if err != nil {
			t.Fatalf("generateICal() error = %v", err)
		}

		output := buf.String()

		// Check for required iCal components
		checks := []string{
			"BEGIN:VCALENDAR",
			"END:VCALENDAR",
			"NAME:Test Calendar",
			"TIMEZONE-ID:UTC",
			"BEGIN:VEVENT",
			"END:VEVENT",
			"SUMMARY:Test Event - D-DAY 💚",
			"DESCRIPTION:Test Description",
			"SUMMARY:Test Event - 1m 💚",
			"SUMMARY:Test Event - 1y 💚",
		}

		for _, check := range checks {
			if !strings.Contains(output, check) {
				t.Errorf("generateICal() output missing %q", check)
			}
		}
	})

	t.Run("Recurring annual events", func(t *testing.T) {
		config := Config{
			Timezone:     "UTC",
			CalendarName: "Test Calendar",
			Events: []Event{
				{
					MonthDay:    "07-04",
					Title:       "Independence Day",
					Description: "Annual celebration",
				},
			},
		}

		var buf bytes.Buffer
		err := generateICal(config, &buf)
		if err != nil {
			t.Fatalf("generateICal() error = %v", err)
		}

		output := buf.String()
		currentYear := time.Now().Year()

		// Check for required iCal components
		checks := []string{
			"BEGIN:VCALENDAR",
			"END:VCALENDAR",
			"NAME:Test Calendar",
			"SUMMARY:Independence Day 💚",
			"DESCRIPTION:Annual celebration",
			fmt.Sprintf("DTSTART;VALUE=DATE:%d0704", currentYear-1),
			fmt.Sprintf("DTSTART;VALUE=DATE:%d0704", currentYear),
			fmt.Sprintf("DTSTART;VALUE=DATE:%d0704", currentYear+1),
		}

		for _, check := range checks {
			if !strings.Contains(output, check) {
				t.Errorf("generateICal() output missing %q", check)
			}
		}

		// Ensure no duration is shown for recurring events
		if strings.Contains(output, " - ") {
			t.Error("Recurring events should not show duration")
		}
	})

	t.Run("Countdown events", func(t *testing.T) {
		// Set a fixed future date for testing
		futureDate := time.Now().AddDate(0, 3, 10) // 3 months and 10 days from now
		
		config := Config{
			Timezone:     "UTC",
			CalendarName: "Test Calendar",
			Anniversaries: Anniversary{
				Years:  []int{1},
				Months: []int{1, 3},
				Days:   []int{7, 100},
			},
			Events: []Event{
				{
					Date:        futureDate.Format("2006-01-02"),
					Title:       "Big Launch",
					Description: "Product launch date",
				},
			},
		}

		var buf bytes.Buffer
		err := generateICal(config, &buf)
		if err != nil {
			t.Fatalf("generateICal() error = %v", err)
		}

		output := buf.String()

		// Check for countdown markers and anniversary markers
		hasCountdown := strings.Contains(output, "Big Launch - D-7") ||
			strings.Contains(output, "Big Launch - D-100") ||
			strings.Contains(output, "Big Launch - D-1m") ||
			strings.Contains(output, "Big Launch - D-3m")
		
		hasAnniversary := strings.Contains(output, "Big Launch - D-DAY") ||
			strings.Contains(output, "Big Launch - 7d") ||
			strings.Contains(output, "Big Launch - 1m") ||
			strings.Contains(output, "Big Launch - 1y")
		
		if !hasCountdown {
			t.Error("generateICal() should have countdown events for future dates")
		}
		
		if !hasAnniversary {
			t.Error("generateICal() should have anniversary events for future dates")
		}
		
		if !strings.Contains(output, "DESCRIPTION:Product launch date") {
			t.Error("generateICal() should include event description")
		}
	})

	t.Run("No past events", func(t *testing.T) {
		config := Config{
			Timezone:     "UTC",
			CalendarName: "Test Calendar",
			Anniversaries: Anniversary{
				Years:  []int{1},
				Months: []int{1},
				Days:   []int{0, 7},
			},
			Events: []Event{
				{
					Date:        "2023-01-01",
					Title:       "Past Event",
					NoPast:      true,
				},
			},
		}

		var buf bytes.Buffer
		err := generateICal(config, &buf)
		if err != nil {
			t.Fatalf("generateICal() error = %v", err)
		}

		output := buf.String()

		// Should not contain any events since it's a past date with no_past=true
		if strings.Contains(output, "Past Event") {
			t.Error("no_past flag not working - past events should be skipped")
		}
	})

	t.Run("No future countdown", func(t *testing.T) {
		futureDate := time.Now().AddDate(0, 3, 0)
		
		config := Config{
			Timezone:     "UTC",
			CalendarName: "Test Calendar",
			Anniversaries: Anniversary{
				Years:  []int{1},
				Months: []int{1},
				Days:   []int{7, 100},
			},
			Events: []Event{
				{
					Date:        futureDate.Format("2006-01-02"),
					Title:       "Future Event",
					NoFuture:    true,
				},
			},
		}

		var buf bytes.Buffer
		err := generateICal(config, &buf)
		if err != nil {
			t.Fatalf("generateICal() error = %v", err)
		}

		output := buf.String()

		// Should not contain countdown events
		if strings.Contains(output, "D-7") || strings.Contains(output, "D-1") {
			t.Error("no_future flag not working - countdown events should be skipped")
		}

		// Should not contain any future events including D-DAY since the date is in the future
		if strings.Contains(output, "Future Event") {
			t.Error("no_future flag should skip all future events")
		}
	})

	t.Run("Mixed past and future with flags", func(t *testing.T) {
		futureDate := time.Now().AddDate(0, 6, 0)
		
		config := Config{
			Timezone:     "UTC",
			CalendarName: "Test Calendar",
			Anniversaries: Anniversary{
				Years:  []int{1},
				Months: []int{1},
				Days:   []int{7},
			},
			Events: []Event{
				{
					Date:        futureDate.Format("2006-01-02"),
					Title:       "Countdown Only",
					NoPast:      true,  // Only countdown, no anniversaries
				},
				{
					Date:        "2023-01-01",
					Title:       "Anniversary Only",
					NoFuture:    true,  // Only past anniversaries
				},
			},
		}

		var buf bytes.Buffer
		err := generateICal(config, &buf)
		if err != nil {
			t.Fatalf("generateICal() error = %v", err)
		}

		output := buf.String()

		// Check countdown only event
		if !strings.Contains(output, "Countdown Only - D-") {
			t.Error("Countdown Only event should have countdown events")
		}
		if strings.Contains(output, "Countdown Only - 7d") || strings.Contains(output, "Countdown Only - 1m") {
			t.Error("Countdown Only event should not have anniversary events")
		}

		// Check anniversary only event
		if strings.Contains(output, "Anniversary Only - D-") && !strings.Contains(output, "Anniversary Only - D-DAY") {
			t.Error("Anniversary Only event should not have countdown events")
		}
		if !strings.Contains(output, "Anniversary Only - 1y") {
			t.Error("Anniversary Only event should have anniversary events")
		}
	})
}

func TestLoadConfig(t *testing.T) {
	// Test invalid file
	_, err := loadConfig("/nonexistent/file.toml")
	if err == nil {
		t.Error("loadConfig should fail for nonexistent file")
	}
}
