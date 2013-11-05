package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"shipshape/events"
	"supervisor"
)

func Print(format string, params ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", params...)
}

func Info(format string, params ...interface{}) {
	fmt.Fprintf(os.Stderr, "INFO: "+format+"\n", params...)
}

func Error(format string, params ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", params...)
}

func run(in io.Reader, out io.Writer, url string) (err error) {
	ch := make(chan *supervisor.Event)
	go func() {
		defer func() {
			close(ch)
		}()

		err := supervisor.Listen(in, out, ch)
		if err != nil {
			Error(err.Error())
		}
	}()

	var shipev *events.Event
	var superev *supervisor.Event
	client := events.NewClient(url)
	for {
		superev = <-ch
		switch {
		case superev.Name() == "PROCESS_STATE_RUNNING":
			shipev = &events.Event{superev.Meta["processname"], events.Running}
		case superev.Name()[:14] == "PROCESS_STATE_":
			shipev = &events.Event{superev.Meta["processname"], events.Stopped}
		default:
			continue
		}

		Info("Sending %s to %v\n", shipev, client.Url)
		err := client.Send(shipev)
		if err != nil {
			Error(err.Error())
		}
	}
}

func main() {
	runtime.GOMAXPROCS(2)

	urlp := flag.String("url", "http://127.0.0.1:5009/event", "URL to send events to.")
	flag.Parse()
	run(os.Stdin, os.Stdout, *urlp)
}
