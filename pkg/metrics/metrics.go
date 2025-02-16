package metrics

import (
	"sync"
	"time"
)

var (
	collector *Collector
	once      sync.Once
)

type Collector struct {
	mu     sync.Mutex
	events []Event
}

type Event struct {
	Timestamp time.Time
	Type      string
	Success   bool
	Metadata  map[string]interface{}
}

// Add security audit events
const (
	EventTypeBackupCreated = "backup_created"
	EventTypeKeyRotated    = "key_rotated"
	EventTypeConfigChanged = "config_changed"
)

func Record(data map[string]interface{}) {
	once.Do(func() {
		collector = &Collector{}
	})

	event := Event{
		Timestamp: time.Now(),
		Type:      data["type"].(string),
		Success:   data["success"].(bool),
		Metadata:  data,
	}

	collector.mu.Lock()
	defer collector.mu.Unlock()
	collector.events = append(collector.events, event)
}

func RecordSecurityEvent(eventType string, meta map[string]interface{}) {
	Record(map[string]interface{}{
		"type":      "security",
		"subtype":   eventType,
		"timestamp": time.Now().UTC(),
		"metadata":  meta,
	})
}
