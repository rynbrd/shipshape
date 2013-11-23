package main

import (
	"go-supervisor/supervisor"
	"io"
	"net/http"
	"shipshape/events"
	"strconv"
	"testing"
)


func TestOiler(t *testing.T) {
	stdin, stdinWriter := io.Pipe()
	stdoutReader, stdout := io.Pipe()

	serial := 0

	ch := make(chan *events.Event, 10)
	defer func() {
		close(ch)
	}()

	go func() {
		handler := events.MakeHandler(ch)
		http.HandleFunc("/event", handler)
		http.ListenAndServe("127.0.0.1:5009", nil)
	}()

	config := Config {
		DefaultGreaserUrl,
		DefaultSupervisorUrl,
		stdin,
		stdout,
		DefaultMonitorQueueSize,
	}

	go Run(&config)

	sendAndVerify := func(eventname string, meta map[string]string, payload []byte, state string) {
		serialstr := strconv.Itoa(serial)
		sentEvent := supervisor.Event{
			map[string]string{
				"ver":    "3.0",
				"server": "supervisor",
				"pool":   "listener",
				"serial": serialstr,
				"poolserial": serialstr,
				"eventname": eventname,
			},
			map[string]string{
				"processname": "test",
				"groupname":   "test",
			},
			payload,
		}

		for k, v := range meta {
			sentEvent.Meta[k] = v
		}

		stdinWriter.Write(sentEvent.ToBytes())
		serial += 1

		if result, err := supervisor.ReadResult(stdoutReader); err != nil {
			t.Errorf(`supervisor.ReadResult() => error{"%v"}, want result="OK"`, err)
		} else if string(result) != "OK" {
			t.Errorf(`supervisor.ReadResult() => "%s", want "OK"`, result)
		}

		if state == "" {
			return
		}

		if receivedEvent, ok := <-ch; !ok {
			t.Errorf(`(event, ok := <-ch) => channel closed, want event`)
		} else if receivedEvent.State != state {
			t.Errorf(`(event, ok := <-ch) => got state="%s", want state="%s"`, receivedEvent.State, state)
		}
	}

	var meta map[string]string

	meta = map[string]string{"pid": "12942"}
	sendAndVerify("PROCESS_LOG_STDERR", meta, []byte("made up log data"), "")

	meta = map[string]string{"from_state": "STARTING"}
	sendAndVerify("PROCESS_STATE_RUNNING", meta, []byte{}, events.Running)

	meta = map[string]string{"from_state": "RUNNING"}
	sendAndVerify("PROCESS_STATE_STOPPED", meta, []byte{}, events.Stopped)

	meta = map[string]string{"from_state": "RUNNING"}
	sendAndVerify("PROCESS_STATE_FATAL", meta, []byte{}, events.Failed)
}
