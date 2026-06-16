package widgets

import "log"

// Registry manages all enabled widgets.
type Registry struct {
	widgets []Widget
}

// NewRegistry creates an empty widget registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a widget to the registry.
func (r *Registry) Register(w Widget) {
	r.widgets = append(r.widgets, w)
	log.Printf("Widget registered: %s (%s)", w.Name(), w.Type())
}

// FetchAll collects data from all registered widgets.
func (r *Registry) FetchAll() []WidgetData {
	var data []WidgetData
	for _, w := range r.widgets {
		wd, err := w.Fetch()
		if err != nil {
			log.Printf("Widget %s fetch error: %v", w.Name(), err)
			continue
		}
		if wd != nil {
			data = append(data, *wd)
		}
	}
	return data
}

// Count returns the number of registered widgets.
func (r *Registry) Count() int {
	return len(r.widgets)
}
