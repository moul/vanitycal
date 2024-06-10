package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	ical "github.com/arran4/golang-ical"
	"github.com/BurntSushi/toml"
)

type Event struct {
	Date  string `toml:"date"`
	Title string `toml:"title"`
}

type Config struct {
	Events []Event `toml:"events"`
}

func main() {
	configFile := flag.String("config", "", "Path to the config file (use '-' for stdin)")
	outputFile := flag.String("output", "", "Path to the output file (use '-' for stdout)")
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
		fmt.Println("Error reading config file:", err)
		return
	}

	var output io.Writer
	if *outputFile == "-" {
		output = os.Stdout
	} else {
		file, err := os.Create(*outputFile)
		if err != nil {
			fmt.Println("Error creating output file:", err)
			return
		}
		defer file.Close()
		output = file
	}

	err = generateICal(config, output)
	if err != nil {
		fmt.Println("Error generating iCal file:", err)
	}
}

func generateICal(config Config, output io.Writer) error {
	cal := ical.NewCalendar()
	cal.SetMethod(ical.MethodPublish)

	for _, event := range config.Events {
		date, err := time.Parse("2006-01-02", event.Date)
		if err != nil {
			fmt.Println("Error parsing date:", err)
			continue
		}
		anniversaries := getAnniversaries(date)
		for _, anniv := range anniversaries {
			duration := getDuration(date, anniv)
			icalEvent := cal.AddEvent(fmt.Sprintf("anniv-%s", anniv.Format("20060102")))
			icalEvent.SetSummary(fmt.Sprintf("%s - %s", event.Title, duration))
			icalEvent.SetStartAt(anniv)
			icalEvent.SetEndAt(anniv.Add(24 * time.Hour))
		}
	}

	_, err := output.Write([]byte(cal.Serialize()))
	return err
}

func getAnniversaries(date time.Time) []time.Time {
	return []time.Time{
		date.AddDate(1, 0, 0),    // 1 year
		date.AddDate(2, 0, 0),    // 2 years
		date.AddDate(3, 0, 0),    // 3 years
		date.AddDate(4, 0, 0),    // 4 years
		date.AddDate(5, 0, 0),    // 5 years
		date.AddDate(6, 0, 0),    // 6 years
		date.AddDate(7, 0, 0),    // 7 years
		date.AddDate(8, 0, 0),    // 8 years
		date.AddDate(9, 0, 0),    // 9 years
		date.AddDate(10, 0, 0),    // 10 years
		date.AddDate(15, 0, 0),    // 15 years
		date.AddDate(20, 0, 0),    // 20 years
		date.AddDate(25, 0, 0),    // 25 years
		date.AddDate(30, 0, 0),    // 30 years
		date.AddDate(35, 0, 0),    // 35 years
		date.AddDate(40, 0, 0),    // 40 years
		date.AddDate(45, 0, 0),    // 45 years
		date.AddDate(50, 0, 0),    // 50 years
		date.AddDate(0, 0, 100),   // 100 days
		date.AddDate(0, 0, 1_000), // 1 000 days
		date.AddDate(0, 0, 10_000), // 10 000 days
		date.AddDate(0, 0, 100_000), // 100 000 days
		date.AddDate(0, 1, 0),    // 1 month
		date.AddDate(0, 2, 0),    // 2 month
		date.AddDate(0, 3, 0),    // 3 month
		date.AddDate(0, 6, 0),    // 6 months
	}
}

func getDuration(start, end time.Time) string {
	years := end.Year() - start.Year()
	months := int(end.Sub(start).Hours() / (24 * 30))
	days := int(end.Sub(start).Hours() / 24)

	if years > 0 && end.AddDate(-years, 0, 0).Equal(start) {
		return fmt.Sprintf("%dy", years)
	} else if months >= 12 && end.AddDate(0, -months, 0).Equal(start) {
		return fmt.Sprintf("%dy", months/12)
	} else if months > 0 && end.AddDate(0, -months, 0).Equal(start) {
		return fmt.Sprintf("%dm", months)
	} else if days >= 1000 && end.AddDate(0, 0, -days).Equal(start) {
		return fmt.Sprintf("%dd", days)
	} else if days >= 100 && end.AddDate(0, 0, -days).Equal(start) {
		return fmt.Sprintf("%dd", days)
	} else {
		return fmt.Sprintf("%dd", days)
	}
}
