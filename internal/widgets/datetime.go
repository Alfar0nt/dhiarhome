package widgets

import (
	"fmt"
	"time"

	"dhiarhome/internal/config"
)

// DateTimeWidget displays current date and time.
type DateTimeWidget struct {
	cfg      config.DateTimeWidgetConfig
	location *time.Location
}

// NewDateTimeWidget creates a new datetime widget from config.
func NewDateTimeWidget(cfg config.DateTimeWidgetConfig) *DateTimeWidget {
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		loc = time.Local
	}
	return &DateTimeWidget{cfg: cfg, location: loc}
}

func (d *DateTimeWidget) Name() string { return "Date & Time" }
func (d *DateTimeWidget) Type() string { return "datetime" }

func (d *DateTimeWidget) Fetch() (*WidgetData, error) {
	now := time.Now().In(d.location)

	var timeStr string
	if d.cfg.Format24h {
		timeStr = now.Format("15:04:05")
	} else {
		timeStr = now.Format("3:04:05 PM")
	}

	dayOfWeek := now.Format("Monday")
	fullDate := now.Format("Monday, January 2, 2006") // include weekday to match JS clock format

	return &WidgetData{
		Type:  "datetime",
		Label: dayOfWeek,
		Icon:  "🕐",
		Values: map[string]interface{}{
			"time":      timeStr,
			"day":       dayOfWeek,
			"date":      fullDate,
			"timezone":  d.cfg.Timezone,
			"format24h": d.cfg.Format24h,
		},
	}, nil
}

// TimezoneOffset returns the UTC offset string for client-side JS.
func (d *DateTimeWidget) TimezoneOffset() string {
	_, offset := time.Now().In(d.location).Zone()
	hours := offset / 3600
	mins := (offset % 3600) / 60
	if mins < 0 {
		mins = -mins
	}
	return fmt.Sprintf("%+d:%02d", hours, mins)
}
