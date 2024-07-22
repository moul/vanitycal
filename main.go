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
	Date        string `toml:"date"`
	Title       string `toml:"title"`
	Description string `toml:"description"`
}

type Config struct {
	Events []Event `toml:"events"`
}

func main() {
	configFile := flag.String("config", "-", "Path to the config file (use '-' for stdin)")
	outputFile := flag.String("output", "-", "Path to the output file (use '-' for stdout)")
	flag.Parse()

	if *configFile == "" || *outputFile == "" {
		fmt.Println("Both config and output flags are required")
		flag.Usage()
		return
	}

	var config Config
	var err error

	if *configFile == "-" {
		_, err = toml.NewDecoder(os.Stdin).Decode(&config)
	} else {
		_, err = toml.DecodeFile(*configFile, &config)
	}

	if err != nil {
		panic(fmt.Errorf("Error reading config file: %w", err))
	}

	var output io.Writer
	if *outputFile == "-" {
		output = os.Stdout
	} else {
		file, err := os.Create(*outputFile)
		if err != nil {
			panic(fmt.Errorf("Error creating output file: %w", err))
		}
		defer file.Close()
		output = file
	}

	err = generateICal(config, output)
	if err != nil {
		panic(fmt.Errorf("Error generating ics file: %w", err))
	}
}

func generateICal(config Config, output io.Writer) error {
	cal := ical.NewCalendar()
	cal.SetMethod(ical.MethodPublish)
	cal.SetName("VanityCal 💚")
	cal.SetDescription("")
	cal.SetTimezoneId("Europe/Paris")
	cal.SetTzid("Europe/Paris")
	cal.SetCalscale("GREGORIAN")
	cal.SetLastModified(time.Now()) // XXX: take last modification date of this binary AND the input.

	for _, event := range config.Events {
		date, err := time.Parse("2006-01-02", event.Date)
		if err != nil {
			return fmt.Errorf("Error parsing date: %w", err)
		}
		anniversaries := getAnniversaries(date)
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
	}

	_, err := output.Write([]byte(cal.Serialize()))
	return err
}

func getAnniversaries(date time.Time) []time.Time {
	return []time.Time{
		date,                       // d day
		date.AddDate(1, 0, 0),      // 1 year
		date.AddDate(2, 0, 0),      // 2 years
		date.AddDate(3, 0, 0),      // 3 years
		date.AddDate(4, 0, 0),      // 4 years
		date.AddDate(5, 0, 0),      // 5 years
		date.AddDate(6, 0, 0),      // 6 years
		date.AddDate(7, 0, 0),      // 7 years
		date.AddDate(8, 0, 0),      // 8 years
		date.AddDate(9, 0, 0),      // 9 years
		date.AddDate(10, 0, 0),     // 10 years
		date.AddDate(15, 0, 0),     // 15 years
		date.AddDate(20, 0, 0),     // 20 years
		date.AddDate(25, 0, 0),     // 25 years
		date.AddDate(30, 0, 0),     // 30 years
		date.AddDate(35, 0, 0),     // 35 years
		date.AddDate(40, 0, 0),     // 40 years
		date.AddDate(45, 0, 0),     // 45 years
		date.AddDate(50, 0, 0),     // 50 years
		date.AddDate(0, 0, 7),      // 7 days
		date.AddDate(0, 0, 100),    // 100 days
		date.AddDate(0, 0, 1_000),  // 1 000 days
		date.AddDate(0, 0, 10_000), // 10 000 days
		date.AddDate(0, 1, 0),      // 1 month
		date.AddDate(0, 2, 0),      // 2 month
		date.AddDate(0, 3, 0),      // 3 month
		date.AddDate(0, 6, 0),      // 6 months
		date.AddDate(0, 9, 0),      // 9 months
	}
}

func getDuration(start, end time.Time) string {
	years := end.Year() - start.Year()
	months := int(end.Sub(start).Hours() / (24 * 30))
	days := int(end.Sub(start).Hours() / 24)

	if end == start {
		return "D-DAY"
	}
	if years > 0 && end.AddDate(-years, 0, 0).Equal(start) {
		return fmt.Sprintf("%dy", years)
	} else if months >= 12 && end.AddDate(0, -months, 0).Equal(start) {
		return fmt.Sprintf("%dy", months/12)
	} else if months > 0 && end.AddDate(0, -months, 0).Equal(start) {
		return fmt.Sprintf("%dm", months)
	} else {
		return fmt.Sprintf("%dd", days)
	}
}
