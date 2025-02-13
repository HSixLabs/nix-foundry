package progress

import (
	"fmt"
)

type Spinner struct {
	message string
	active  bool
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
	}
}

func (s *Spinner) Start() {
	s.active = true
	fmt.Printf("⏳ %s\n", s.message)
}

func (s *Spinner) Stop() {
	s.active = false
}

func (s *Spinner) Success(message string) {
	s.active = false
	fmt.Printf("✅ %s\n", message)
}

func (s *Spinner) Fail(message string) {
	s.active = false
	fmt.Printf("❌ %s\n", message)
}
