package widgets

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"dhiarhome/internal/config"
)

// WeatherWidget displays mock weather data (demo only).
type WeatherWidget struct {
	cfg   config.WeatherWidgetConfig
	cache struct {
		sync.RWMutex
		data      *WidgetData
		fetchedAt time.Time
	}
}

func NewWeatherWidget(cfg config.WeatherWidgetConfig) *WeatherWidget {
	return &WeatherWidget{cfg: cfg}
}

func (w *WeatherWidget) Name() string { return "Weather" }
func (w *WeatherWidget) Type() string { return "weather" }

func (w *WeatherWidget) Fetch() (*WidgetData, error) {
	// Cache mock data for 5 minutes to prevent random changes on every HTMX poll
	w.cache.RLock()
	if w.cache.data != nil && time.Since(w.cache.fetchedAt) < 5*time.Minute {
		data := w.cache.data
		w.cache.RUnlock()
		return data, nil
	}
	w.cache.RUnlock()

	conditions := []struct {
		icon string
		desc string
	}{
		{"☀️", "Clear sky"},
		{"⛅", "Partly cloudy"},
		{"🌧️", "Light rain"},
		{"🌤️", "Mainly clear"},
	}
	c := conditions[rand.Intn(len(conditions))]

	tempUnit := "°C"
	temp := 15.0 + rand.Float64()*15.0
	if w.cfg.Units == "fahrenheit" {
		tempUnit = "°F"
		temp = temp*9/5 + 32
	}

	result := &WidgetData{
		Type:  "weather",
		Label: "Weather",
		Icon:  c.icon,
		Values: map[string]interface{}{
			"temperature": fmt.Sprintf("%.1f%s", temp, tempUnit),
			"condition":   c.desc,
			"icon":        c.icon,
			"wind_speed":  fmt.Sprintf("%.0f km/h", 5.0+rand.Float64()*20.0),
		},
	}

	w.cache.Lock()
	w.cache.data = result
	w.cache.fetchedAt = time.Now()
	w.cache.Unlock()

	return result, nil
}
