package main

import (
	"fmt"
	"os"
	"time"

	ical "github.com/arran4/golang-ical"
)

type Event struct {
	Date  time.Time
	Title string
}

type DateList struct {
	Events []Event
}

func main() {
	// Example events
	events := DateList{
		Events: []Event{
			{Date: time.Date(2022, time.January, 25, 0, 0, 0, 0, time.UTC), Title: "Project Launch"},
			{Date: time.Date(2023, time.March, 14, 0, 0, 0, 0, time.UTC), Title: "Product Release"},
		},
	}

	// Generate iCal file
	err := generateICal(events)
	if err != nil {
		fmt.Println("Error generating iCal file:", err)
	}
}

func generateICal(dates DateList) error {
	cal := ical.NewCalendar()
	cal.SetMethod(ical.MethodPublish)

	for _, event := range dates.Events {
		anniversaries := getAnniversaries(event.Date)
		for _, anniv := range anniversaries {
			duration := getDuration(event.Date, anniv)
			icalEvent := cal.AddEvent(fmt.Sprintf("anniv-%s", anniv.Format("20060102")))
			icalEvent.SetSummary(fmt.Sprintf("%s - %s", event.Title, duration))
			icalEvent.SetStartAt(anniv)
			icalEvent.SetEndAt(anniv.Add(24 * time.Hour))
		}
	}

	file, err := os.Create("anniversaries.ics")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(cal.Serialize()))
	return err
}

func getAnniversaries(date time.Time) []time.Time {
	return []time.Time{
		date.AddDate(1, 0, 0),    // 1 year
		date.AddDate(2, 0, 0),    // 2 years
		date.AddDate(3, 0, 0),    // 3 years
		date.AddDate(0, 0, 100),  // 100 days
		date.AddDate(0, 0, 1000), // 1000 days
		date.AddDate(0, 1, 0),    // 1 month
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
