package main

import (
	"./config"
	"fmt"
	"github.com/BlueDragonX/go-service/service"
	"os"
	"text/template"
)

type Service struct {
	Commands chan service.Command
	Metadata map[string]string
	config   *config.Service
	stdout   *os.File
	stderr   *os.File
	*service.Service
}

// NewService creates a new service with the provided configuration.
func NewService(cfg *config.Service, events chan<- service.Event) (s *Service, err error) {
	env := make([]string, 0, len(cfg.Environment))
	for key, value := range cfg.Environment {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	var stdout, stderr *os.File

	defer func() {
		if err != nil {
			if stderr != stdout && stderr != nil {
				stderr.Close()
			}
			if stdout != nil {
				stdout.Close()
			}
		}
	}()

	getStream := func(file string) (stream *os.File, err error) {
		if file != "/dev/null" {
			stream, err = os.OpenFile(file, os.O_WRONLY|os.O_APPEND, 644)
		}
		return
	}

	stdout, err = getStream(cfg.Stdout)
	if err != nil {
		return
	}

	if cfg.Stdout == cfg.Stderr {
		stderr = stdout
	} else {
		stderr, err = getStream(cfg.Stderr)
		if err != nil {
			return
		}
	}

	s = &Service{
		make(chan service.Command, CommandQueueSize),
		make(map[string]string),
		cfg,
		stdout,
		stderr,
		nil,
	}

	if s.Service, err = service.NewService(cfg.Command); err != nil {
		return
	}

	s.Service.Directory = cfg.Directory
	s.Service.Environment = env
	s.Service.StartTimeout = cfg.StartTimeout
	s.Service.StartRetries = cfg.StartRetries
	s.Service.StopSignal = cfg.StopSignal
	s.Service.StopTimeout = cfg.StopTimeout
	s.Service.StopRestart = cfg.StopRestart
	s.Service.Stdout = stdout
	s.Service.Stderr = stderr
	s.Service.CommandHook = func(svc *service.Service, command string) (err error) {
		if command == service.Start {
			context := struct {
				Config   *config.Service
				Metadata interface{}
			}{
				s.config,
				s.Metadata,
			}

			err = s.RenderTemplates(context)
		}
		return
	}

	go s.Service.Run(s.Commands, events)
	return
}

// Name gets the name of the service.
func (s Service) Name() string {
	return s.config.Name
}

// Ports returns the ports this service exposes.
func (s Service) Ports() []config.Port {
	return s.config.Ports
}

// Close frees resources associated with the service.
func (s Service) Close() (err error) {
	close(s.Commands)

	if s.stderr != nil && s.stderr != s.stdout {
		if err = s.stderr.Close(); err != nil {
			return
		}
	}
	if s.stdout != nil {
		err = s.stdout.Close()
	}
	return
}

// RenderTemplates generates config files from the service templates.
func (s Service) RenderTemplates(context interface{}) error {
	for _, tpl := range s.config.Templates {
		out, err := os.Create(tpl.File)
		if err != nil {
			return err
		}
		defer out.Close()

		renderer, err := template.ParseFiles(tpl.Source)
		if err != nil {
			return err
		}

		if err = renderer.Execute(out, context); err != nil {
			return err
		}
	}
	return nil
}
