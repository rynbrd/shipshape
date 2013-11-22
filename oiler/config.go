package main

import (
	"flag"
	"io"
	"os"
)

const (
	DefaultGreaserUrl = "http://127.0.0.1:5009/event"
	DefaultSupervisorUrl = "http://127.0.0.1:9001/RPC2"
	DefaultMonitorQueueSize = 10
)

type Config struct {
	GreaserUrl string
	SupervisorUrl string
	InStream io.Reader
	OutStream io.Writer
	MonitorQueueSize int
}

// LoadConfig parses the commandline arguments into a Config struct.
func LoadConfig() Config {
	greaser := flag.String("greaser", DefaultGreaserUrl, "Greaser URL to send events to.")
	supervisor := flag.String("supervisor", DefaultSupervisorUrl, "Supervisor URL to poll.")
	flag.Parse()

	return Config{
		*greaser,
		*supervisor,
		os.Stdin,
		os.Stdout,
		DefaultMonitorQueueSize,
	}
}
