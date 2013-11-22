package main

import (
	"syscall"
	"go-supervisor/supervisor"
	"os"
	"runtime"
	"shipshape/events"
	"os/signal"
)

func getState(process *supervisor.Process) string {
	switch process.State {
	case supervisor.Running:
		return events.Running
	case supervisor.Stopped:
		return events.Stopped
	case supervisor.Exited:
		fallthrough
	case supervisor.Fatal:
		return events.Failed
	}
	return ""
}

func Run(config *Config) {
	done := make(chan bool)
	monitorEvents := make(chan interface{}, config.MonitorQueueSize)
	monitor, err := supervisor.NewMonitor(config.SupervisorUrl, config.InStream, config.OutStream, monitorEvents)
	if err != nil {
		Fatal(err, 1)
	}

	eventClient := events.NewClient(config.GreaserUrl)

	send := func(name string, state string) {
		Debug("%s -> %s", name, state)
		if err := eventClient.Send(&events.Event{name, state}); err != nil {
			Error(err)
		}
	}

	shutdown := func() {
		for _, process := range monitor.Processes {
			if getState(process) != events.Stopped {
				send(process.Name, events.Stopped)
			}
		}
		Info("Exiting.")
		os.Exit(0)
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT)
	signal.Notify(sigchan, syscall.SIGTERM)
	go func() {
		signal := <-sigchan
		Info("Received signal: %s", signal)
		shutdown()
	}()

	go func() {
		var process supervisor.Process
		var state string

		for event := range monitorEvents {
			switch event.(type) {
			case supervisor.ProcessAddEvent:
				process = (event.(supervisor.ProcessAddEvent)).Process
			case supervisor.ProcessRemoveEvent:
				process = (event.(supervisor.ProcessRemoveEvent)).Process
			case supervisor.ProcessStateEvent:
				process = (event.(supervisor.ProcessStateEvent)).Process
			default:
				continue
			}

			if state = getState(&process); state != "" {
				send(process.Name, state)
			}
		}
		done <- true
	}()

	if err = monitor.Run(); err != nil {
		Fatal(err, 1)
	}

	<-done
	Info("Monitor completed.")
	shutdown()
}

func main() {
	runtime.GOMAXPROCS(2)
	config := LoadConfig()
	Run(&config)
}
