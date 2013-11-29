package main

import (
	"./config"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/template"
	"time"
)

const (
	// Service commands.
	Start    = "start"
	Stop     = "stop"
	Restart  = "restart"
	Shutdown = "shutdown"

	// Service command result.
	Success = "success"
	Failure = "failure"

	// Service states.
	Starting = "starting"
	Running  = "running"
	Stopping = "stopping"
	Stopped  = "stopped"
	Exited   = "exited"
	//TODO: Implement a Backoff state.
)

type InvalidStateError string

func (err InvalidStateError) Error() string {
	return string(err)
}

type ServiceCommandResult struct {
	Service *Service
	Name    string
	Error   error
}

type ServiceCommand struct {
	Name   string
	Result chan<- ServiceCommandResult
}

// Respond creates and sends a command result.
func (cmd ServiceCommand) Respond(service *Service, err error) {
	if cmd.Result != nil {
		cmd.Result <- ServiceCommandResult{service, cmd.Name, err}
	}
}

type ServiceEvent struct {
	Service *Service
	State   string
}

type Service struct {
	command *exec.Cmd
	config  *config.Service
	state   string
}

// NewService creates a new service with the provided configuration.
func NewService(cfg *config.Service) *Service {
	return &Service{nil, cfg, Stopped}
}

// Name gets the name of the service.
func (s Service) Name() string {
	return s.config.Name
}

// State gets the current state of the service.
func (s Service) State() string {
	return s.state
}

// Pid gets the PID of the service or 0 if not Running or Stopping.
func (s Service) Pid() int {
	if s.state != Running && s.state != Stopping {
		return 0
	}
	return s.command.Process.Pid
}

// Command gets the running command.
func (s Service) Command() []string {
	return s.config.Command
}

func (s Service) makeCommand() (cmd *exec.Cmd, err error) {
	cmd = exec.Command(s.config.Command[0], s.config.Command[1:]...)

	getWriter := func(path string) (*os.File, error) {
		if path == "/dev/null" {
			return nil, nil
		}
		return os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 644)
	}

	defer func() {
		if err != nil {
			if cmd.Stderr != nil && cmd.Stderr != cmd.Stdout {
				cmd.Stderr.(*os.File).Close()
			}
			if cmd.Stdout != nil {
				cmd.Stdout.(*os.File).Close()
			}
		}
	}()

	if cmd.Stdout, err = getWriter(s.config.StdoutDest); err != nil {
		return
	}
	if strings.ToUpper(s.config.StderrDest) == "STDOUT" {
		cmd.Stderr = cmd.Stdout
	} else if cmd.Stderr, err = getWriter(s.config.StderrDest); err != nil {
		return
	}

	env := make([]string, 0, len(s.config.Env))
	for k, v := range s.config.Env {
		env = append(env, fmt.Sprintf(`{}={}`, k, v))
	}

	cmd.Stdin = nil
	cmd.Env = env
	cmd.Dir = s.config.Dir
	return
}

func (s *Service) startProcess(states chan string) (err error) {
	defer func() {
		if err != nil {
			//TODO: Log this error?
			states <- Exited
		}
	}()

	//TODO: Render templates here.
	if s.command, err = s.makeCommand(); err != nil {
		return
	}
	if err = s.command.Start(); err != nil {
		return
	}

	states <- Running
	s.command.Wait()
	states <- Exited
	return
}

func (s *Service) Run(commands <-chan ServiceCommand, events chan<- ServiceEvent) {
	var lastCommand *ServiceCommand
	states := make(chan string)
	quit := make(chan bool, 2)
	kill := make(chan int, 2)
	retries := 0

	defer func() {
		close(states)
		close(quit)
		close(kill)
	}()

	sendEvent := func(state string) {
		s.state = state
		events <- ServiceEvent{s, state}
	}

	sendInvalidCmd := func(cmd *ServiceCommand, state string) {
		if cmd != nil {
			cmd.Respond(s, InvalidStateError(fmt.Sprintf("invalid state transition: %s -> %s", s.state, state)))
		}
	}

	start := func(cmd *ServiceCommand) {
		if s.state != Stopped && s.state != Exited {
			sendInvalidCmd(cmd, Starting)
			return
		}

		sendEvent(Starting)
		go s.startProcess(states)
	}

	stop := func(cmd *ServiceCommand) {
		if s.state != Running {
			sendInvalidCmd(cmd, Stopping)
			return
		}

		sendEvent(Stopping)
		pid := s.Pid()
		s.command.Process.Signal(s.config.StopSignal) //TODO: Check for error.
		go func() {
			time.Sleep(s.config.StopTimeout)
			defer func() {
				if err := recover(); err != nil {
					if _, ok := err.(runtime.Error); !ok {
						panic(err)
					}
				}
			}()
			kill <- pid
		}()
	}

	shutdown := func(cmd *ServiceCommand, lastCmd *ServiceCommand) {
		if lastCmd != nil {
			lastCmd.Respond(s, errors.New("service is shutting down"))
		}
		if s.state == Stopped || s.state == Exited {
			quit <- true
		} else if s.state == Running {
			stop(cmd)
		}
	}

	onRunning := func(cmd *ServiceCommand) {
		sendEvent(Running)
		if cmd != nil {
			switch cmd.Name {
			case Start:
				fallthrough
			case Restart:
				cmd.Respond(s, nil)
			case Shutdown:
				stop(cmd)
			}
		}
		retries = 0
	}

	onStopped := func(cmd *ServiceCommand) {
		sendEvent(Stopped)
		if cmd != nil {
			switch cmd.Name {
			case Restart:
				start(cmd)
			case Stop:
				cmd.Respond(s, nil)
			case Shutdown:
				quit <- true
			}
		}
	}

	onExited := func(cmd *ServiceCommand, retries int) bool {
		sendEvent(Exited)
		if s.config.Restart && retries < s.config.Retries {
			start(cmd)
			return true
		}
		return false
	}

loop:
	for {
		select {
		case state := <-states:
			// running, exited
			switch state {
			case Running:
				onRunning(lastCommand)
			case Exited:
				if s.state == Stopping {
					onStopped(lastCommand)
				} else {
					if onExited(lastCommand, retries) {
						retries++
					}
				}
			}
			if lastCommand != nil {
				if lastCommand.Name == Restart && s.state == Running {
					lastCommand = nil
				} else if lastCommand.Name != Restart && lastCommand.Name != Shutdown {
					lastCommand = nil
				}
			}
		case command := <-commands:
			if lastCommand == nil || lastCommand.Name != Shutdown { // Shutdown cannot be overriden!
				switch command.Name {
				case Start:
					start(&command)
				case Stop:
					stop(&command)
				case Restart:
					stop(&command)
				case Shutdown:
					shutdown(&command, lastCommand)
				}
				lastCommand = &command
			} else {
				command.Respond(s, errors.New("service is shutting down"))
			}
		case <-quit:
			if lastCommand != nil {
				lastCommand.Respond(s, nil)
			}
			break loop
		case pid := <-kill:
			if pid == s.Pid() {
				s.command.Process.Kill() //TODO: Check for error.
			}
		}
	}
}

// RenderTemplates generates config files from the service templates.
func (s Service) RenderTemplates(data interface{}) error {
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

		context := struct {
			Config *config.Service
			Data   interface{}
		}{
			s.config,
			data,
		}

		err = renderer.Execute(out, context)
		if err != nil {
			return err
		}
	}
	return nil
}
