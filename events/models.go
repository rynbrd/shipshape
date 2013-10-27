package events

import (
	"errors"
	"fmt"
)

const (
	Running string = "running"
	Stopped string = "stopped"
	Success string = "success"
	Failure string = "failure"
)

// Event is sent by the client to update a process status.
type Event struct {
	Process string `json:"process"`
	Status  string `json:"status"`
}

// String returns the string representation of the Event.
func (event Event) String() string {
	if event.Process == "" && event.Status == "" {
		return "Event{}"
	} else {
		return fmt.Sprintf(`Event{"%s", "%s"}`, event.Process, event.Status)
	}
}

// Validate returns an error if the event is invalid.
func (event Event) Validate() error {
	switch {
	case event.Process == "":
		return errors.New("event has no process value")
	case event.Status == "":
		return errors.New("event has no status value")
	case event.Status != Running && event.Status != Stopped:
		return errors.New("event has invalid status value")
	default:
		return nil
	}
}

// EventResult is sent from the server in response to an Event sent by a client.
type EventResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// String returns the string representation of the Event.
func (result EventResult) String() string {
	return fmt.Sprintf("EventResult(%s: %s)", result.Status, result.Message)
}
