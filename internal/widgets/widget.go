package widgets

// WidgetData represents the data for a single widget to be rendered in the dashboard.
type WidgetData struct {
	Type   string                 // Widget type identifier (e.g., "weather", "datetime")
	Label  string                 // Display label
	Icon   string                 // Icon name or emoji
	Values map[string]interface{} // Arbitrary key-value data for the template
}

// Widget is the interface that all widgets must implement.
type Widget interface {
	Name() string                // Human-readable name
	Type() string                // Unique type identifier
	Fetch() (*WidgetData, error) // Fetch current data
}
