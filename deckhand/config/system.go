package config

import (
	"fmt"
)

func getDefaultEventUrl() string {
	return fmt.Sprintf(DefaultSystemEventsUrl, "localhost")
}

type System struct {
	EventsEnable bool
}

// SetYAML parses the YAML tree into the configuration object.
func (s *System) SetYAML(tag string, data interface{}) bool {
	s.EventsEnable = GetBool(data, "events-enable", DefaultSystemEventsEnable)
	return true
}

// Validate the configuration object.
func (s System) Validate() []error {
	return []error{}
}
