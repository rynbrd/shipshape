package main

import (
	"./config"
	"fmt"
	"github.com/BlueDragonX/go-service/service"
	"os"
	"syscall"
	"time"
)

const (
	EventQueueSize   = 10
	CommandQueueSize = 1
)

type Deckhand struct {
	Services map[string]*Service
	Events   chan service.Event
	config   *config.Config
}

// NewDeckhand returns a new Deckhand instance with the provided configuration.
func NewDeckhand(config *config.Config) (deckhand *Deckhand, err error) {
	defer func() {
		if err != nil && deckhand != nil {
			deckhand.Close()
		}
	}()

	deckhand = &Deckhand{
		make(map[string]*Service, len(config.Services)),
		make(chan service.Event, EventQueueSize),
		config,
	}

	var svc *Service
	for name, svcConfig := range config.Services {
		if svc, err = NewService(&svcConfig, deckhand.Events); err != nil {
			break
		}
		deckhand.Services[name] = svc
	}
	return
}

// Close frees resources associated with the Deckhand object.
func (d Deckhand) Close() (err error) {
	close(d.Events)
	for _, service := range d.Services {
		if err = service.Close(); err != nil {
			break
		}
	}
	return
}
