package widgets

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"dhiarhome/internal/config"
)

// WeatherWidget fetches weather data from the Open-Meteo API.
type WeatherWidget struct {
	cfg   config.WeatherWidgetConfig
	cache struct {
		sync.RWMutex
		data      *WidgetData
		fetchedAt time.Time
	}
	mockCache struct {
		sync.RWMutex
		data      *WidgetData
		fetchedAt time.Time
	}
}

// openMeteoResponse represents the API response from Open-Meteo.
type openMeteoResponse struct {
	Current struct {
		Temperature2m float64 `json:"temperature_2m"`
		WeatherCode   int     `json:"weather_code"`
		WindSpeed10m  float64 `json:"wind_speed_10m"`
	} `json:"current"`
	CurrentUnits struct {
		Temperature2m string `json:"temperature_2m"`
		WindSpeed10m  string `json:"wind_speed_10m"`
	} `json:"current_units"`
}

// NewWeatherWidget creates a new weather widget from config.
func NewWeatherWidget(cfg config.WeatherWidgetConfig) *WeatherWidget {
	return &WeatherWidget{cfg: cfg}
}

func (w *WeatherWidget) Name() string { return "Weather" }
func (w *WeatherWidget) Type() string { return "weather" }

func (w *WeatherWidget) Fetch() (*WidgetData, error) {
	if w.cfg.Mock {
		return w.mockData(), nil
	}

	// Check cache
	w.cache.RLock()
	if w.cache.data != nil && time.Since(w.cache.fetchedAt) < time.Duration(w.cfg.CacheMinutes)*time.Minute {
		data := w.cache.data
		w.cache.RUnlock()
		return data, nil
	}
	w.cache.RUnlock()

	// Fetch from API
	data, err := w.fetchFromAPI()
	if err != nil {
		return nil, err
	}

	// Update cache
	w.cache.Lock()
	w.cache.data = data
	w.cache.fetchedAt = time.Now()
	w.cache.Unlock()

	return data, nil
}

func (w *WeatherWidget) fetchFromAPI() (*WidgetData, error) {
	units := "celsius"
	if w.cfg.Units == "fahrenheit" {
		units = "fahrenheit"
	}

	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,weather_code,wind_speed_10m&temperature_unit=%s&wind_speed_unit=kmh",
		w.cfg.Latitude, w.cfg.Longitude, units,
	)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("open-meteo request failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResp openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("open-meteo decode failed: %w", err)
	}

	icon, desc := wmoCodeToDescription(apiResp.Current.WeatherCode)
	tempUnit := "°C"
	if units == "fahrenheit" {
		tempUnit = "°F"
	}

	return &WidgetData{
		Type:  "weather",
		Label: "Weather",
		Icon:  icon,
		Values: map[string]interface{}{
			"temperature": fmt.Sprintf("%.1f%s", apiResp.Current.Temperature2m, tempUnit),
			"condition":   desc,
			"icon":        icon,
			"wind_speed":  fmt.Sprintf("%.0f km/h", apiResp.Current.WindSpeed10m),
		},
	}, nil
}

func (w *WeatherWidget) mockData() *WidgetData {
	// Cache mock data for 5 minutes to prevent random changes on every HTMX poll
	w.mockCache.RLock()
	if w.mockCache.data != nil && time.Since(w.mockCache.fetchedAt) < 5*time.Minute {
		data := w.mockCache.data
		w.mockCache.RUnlock()
		return data
	}
	w.mockCache.RUnlock()

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

	w.mockCache.Lock()
	w.mockCache.data = result
	w.mockCache.fetchedAt = time.Now()
	w.mockCache.Unlock()

	return result
}

// wmoCodeToDescription maps WMO weather codes to icons and descriptions.
func wmoCodeToDescription(code int) (string, string) {
	switch {
	case code == 0:
		return "☀️", "Clear sky"
	case code == 1:
		return "🌤️", "Mainly clear"
	case code == 2:
		return "⛅", "Partly cloudy"
	case code == 3:
		return "☁️", "Overcast"
	case code >= 45 && code <= 48:
		return "🌫️", "Foggy"
	case code >= 51 && code <= 55:
		return "🌦️", "Drizzle"
	case code >= 56 && code <= 57:
		return "🌧️", "Freezing drizzle"
	case code >= 61 && code <= 65:
		return "🌧️", "Rain"
	case code >= 66 && code <= 67:
		return "🌧️", "Freezing rain"
	case code >= 71 && code <= 77:
		return "❄️", "Snow"
	case code >= 80 && code <= 82:
		return "🌧️", "Rain showers"
	case code >= 85 && code <= 86:
		return "🌨️", "Snow showers"
	case code == 95:
		return "⛈️", "Thunderstorm"
	case code >= 96 && code <= 99:
		return "⛈️", "Thunderstorm with hail"
	default:
		return "🌡️", "Unknown"
	}
}
