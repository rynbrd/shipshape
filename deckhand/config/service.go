package config

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type Service struct {
	Name         string
	Command      []string
	Ports        []Port
	Templates    []Template
	Directory    string
	Environment  map[string]string
	StartTimeout time.Duration
	StartRetries int
	StopSignal   syscall.Signal
	StopTimeout  time.Duration
	StopRestart  bool
	Stdout       string
	Stderr       string
}

// SetYAML parses the YAML tree into the configuration object.
func (s *Service) SetYAML(tag string, data interface{}) bool {
	cwd, err := os.Getwd()
	if err != nil {
		panic("failed to get current working directory")
	}

	AssertIsMap("service", data)
	AssertHasKeys(data, []string{"command"}, "service")
	s.Command = GetStringArray(data, "command", nil)
	s.Directory = GetString(data, "directory", cwd)
	s.StartTimeout = GetDuration(data, "start-timeout", DefaultServiceStartTimeout)
	s.StartRetries = GetInt(data, "start-retries", DefaultServiceStartRetries)
	s.StopSignal = GetSignal(data, "stop-signal", DefaultServiceStopSignal)
	s.StopTimeout = GetDuration(data, "stop-timeout", DefaultServiceStopTimeout)
	s.StopRestart = GetBool(data, "stop-restart", DefaultServiceStopRestart)
	s.Stdout = GetString(data, "stdout", DefaultServiceStdout)
	s.Stderr = GetString(data, "stderr", DefaultServiceStderr)

	s.Environment = make(map[string]string)
	if values, ok := GetMapItem(data, "environment"); ok {
		AssertIsStringMap("environment", values)
		for name, value := range values.(map[interface{}]interface{}) {
			s.Environment[name.(string)] = value.(string)
		}
	}

	if values, ok := GetMapItem(data, "ports"); ok {
		AssertIsArray("ports", values)
		for _, value := range values.([]interface{}) {
			port := &Port{}
			if port.SetYAML("!!map", value) {
				s.Ports = append(s.Ports, *port)
			}
		}
	}

	if values, ok := GetMapItem(data, "templates"); ok {
		AssertIsArray("templates", values)
		for _, value := range values.([]interface{}) {
			template := &Template{}
			if template.SetYAML("!!map", value) {
				s.Templates = append(s.Templates, *template)
			}
		}
	}

	return true
}

// Validate the configuration object.
func (s Service) Validate() []error {
	errors := make([]error, 0, 10)

	if len(s.Command) == 0 {
		msg := fmt.Sprintf("command is invalid: %+v", s.Command)
		errors = append(errors, ValidationError(msg))
	}

	if _, err := exec.LookPath(s.Command[0]); err != nil {
		msg := fmt.Sprintf("command not found: %+v", s.Command)
		errors = append(errors, ValidationError(msg))
	}

	validateItem := func(item Validater) {
		newErrors := item.Validate()
		if len(newErrors) > 0 {
			errors = append(errors, newErrors...)
		}
	}

	for _, port := range s.Ports {
		validateItem(port)
	}

	for _, template := range s.Templates {
		validateItem(template)
	}

	validateTimeout := func(what string, t time.Duration) {
		if t < 0*time.Second {
			msg := fmt.Sprintf("%s must be n >= 0s: %s", what, t)
			errors = append(errors, ValidationError(msg))
		}
	}

	validateTimeout("start-timeout", s.StartTimeout)
	validateTimeout("stop-timeout", s.StopTimeout)

	if s.StartRetries < 0 {
		msg := fmt.Sprintf("start-retries must be n >= 0: %d", s.StartRetries)
		errors = append(errors, ValidationError(msg))
	}

	validateDest := func(what string, path string) {
		if fi, err := os.Stat(path); err != nil && !os.IsNotExist(err) {
			msg := fmt.Sprintf("%s is not writable: %s", what, path)
			errors = append(errors, ValidationError(msg))
		} else if s.Stdout != "/dev/null" && !fi.Mode().IsRegular() {
			msg := fmt.Sprintf("%s must be a regular file or /dev/null: %s", what, path)
			errors = append(errors, ValidationError(msg))
		}
	}

	validateDest("stdout", s.Stdout)
	if s.Stderr != "stdout" {
		validateDest("stderr", s.Stderr)
	}

	if s.StartRetries < 0 {
		msg := fmt.Sprintf("start-retries must be n >= 0: %d", s.StartRetries)
		errors = append(errors, ValidationError(msg))
	}

	return errors
}
