package cache

import (
	"sync"
	"time"
)

// ServiceState represents a snapshot of a service's health
type ServiceState struct {
	Name         string
	Status       string // "Online", "Offline", "Unknown"
	ResponseTime time.Duration
	Timestamp    time.Time
}

// Node is a single element in the double linked list
type Node struct {
	Value ServiceState
	Prev  *Node
	Next  *Node
}

// HistoryCache is a thread-safe double linked list
type HistoryCache struct {
	head *Node
	tail *Node
	size int
	max  int
	mu   sync.RWMutex
}

func NewHistoryCache(maxSize int) *HistoryCache {
	return &HistoryCache{
		max: maxSize,
	}
}

// Add inserts a new state at the head of the list.
// If the list exceeds max size, the tail is removed.
func (c *HistoryCache) Add(state ServiceState) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node := &Node{Value: state}

	if c.size == 0 {
		c.head = node
		c.tail = node
	} else {
		node.Next = c.head
		c.head.Prev = node
		c.head = node
	}
	c.size++

	if c.size > c.max {
		// Remove tail
		if c.tail != nil && c.tail.Prev != nil {
			c.tail = c.tail.Prev
			c.tail.Next = nil
			c.size--
		}
	}
}

// GetAll returns all cached states from newest to oldest.
func (c *HistoryCache) GetAll() []ServiceState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var states []ServiceState
	curr := c.head
	for curr != nil {
		states = append(states, curr.Value)
		curr = curr.Next
	}
	return states
}

// GetLatest returns the most recent state for a specific service name
func (c *HistoryCache) GetLatest(name string) (ServiceState, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	curr := c.head
	for curr != nil {
		if curr.Value.Name == name {
			return curr.Value, true
		}
		curr = curr.Next
	}
	return ServiceState{}, false
}
