package widgets

import (
	"html"

	"dhiarhome/internal/config"
)

// CustomTextWidget displays user-defined text/HTML content.
type CustomTextWidget struct {
	cfg config.CustomTextWidgetConfig
}

// NewCustomTextWidget creates a new custom text widget from config.
func NewCustomTextWidget(cfg config.CustomTextWidgetConfig) *CustomTextWidget {
	return &CustomTextWidget{cfg: cfg}
}

func (c *CustomTextWidget) Name() string { return c.cfg.Title }
func (c *CustomTextWidget) Type() string { return "custom_text" }

func (c *CustomTextWidget) Fetch() (*WidgetData, error) {
	// Sanitize content: escape HTML entities to prevent XSS
	// but allow basic formatting via line breaks
	content := html.EscapeString(c.cfg.Content)

	return &WidgetData{
		Type:  "custom_text",
		Label: c.cfg.Title,
		Icon:  "📝",
		Values: map[string]interface{}{
			"title":   c.cfg.Title,
			"content": content,
		},
	}, nil
}
