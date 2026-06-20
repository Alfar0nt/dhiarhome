package todo

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Todo represents a single to-do item.
type Todo struct {
	ID        int64  `json:"id"`
	Text      string `json:"text"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
	DoneAt    string `json:"done_at,omitempty"`
}

// Store provides thread-safe CRUD for to-do items with JSON file persistence.
type Store struct {
	mu       sync.RWMutex
	todos    []Todo
	nextID   int64
	filePath string
}

// NewStore creates a new todo store, loading existing data from filePath.
func NewStore(filePath string) *Store {
	s := &Store{
		filePath: filePath,
		todos:    make([]Todo, 0),
		nextID:   1,
	}
	s.load()
	return s
}

func (s *Store) load() {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return // file doesn't exist yet, start empty
	}
	if err := json.Unmarshal(data, &s.todos); err != nil {
		return
	}
	// Find the max ID to set nextID
	for _, t := range s.todos {
		if t.ID >= s.nextID {
			s.nextID = t.ID + 1
		}
	}
}

func (s *Store) save() {
	data, err := json.MarshalIndent(s.todos, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(s.filePath, data, 0644)
}

// GetByID returns a copy of a single to-do item by ID, or nil if not found.
func (s *Store) GetByID(id int64) *Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := range s.todos {
		if s.todos[i].ID == id {
			t := s.todos[i]
			return &t
		}
	}
	return nil
}

// GetAll returns a copy of all to-do items.
func (s *Store) GetAll() []Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Todo, len(s.todos))
	copy(result, s.todos)
	return result
}

// Add creates a new to-do item and returns it.
func (s *Store) Add(text string) Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	t := Todo{
		ID:        s.nextID,
		Text:      text,
		Done:      false,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	s.nextID++
	s.todos = append(s.todos, t)
	s.save()
	return t
}

// Update modifies the text of an existing to-do item.
func (s *Store) Update(id int64, newText string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.todos {
		if s.todos[i].ID == id {
			s.todos[i].Text = newText
			s.save()
			return true
		}
	}
	return false
}

// Toggle flips the done state of a to-do item.
func (s *Store) Toggle(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.todos {
		if s.todos[i].ID == id {
			s.todos[i].Done = !s.todos[i].Done
			if s.todos[i].Done {
				s.todos[i].DoneAt = time.Now().Format(time.RFC3339)
			} else {
				s.todos[i].DoneAt = ""
			}
			s.save()
			return true
		}
	}
	return false
}

// Delete removes a to-do item by ID.
func (s *Store) Delete(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.todos {
		if s.todos[i].ID == id {
			s.todos = append(s.todos[:i], s.todos[i+1:]...)
			s.save()
			return true
		}
	}
	return false
}
