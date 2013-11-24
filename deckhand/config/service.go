package config

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
)

type Service struct {
	Name        string
	Command     []string
	Ports       []Port
	Templates   []Template
	Dir         string
	Env         map[string]string
	Metadata    map[interface{}]interface{}
	StopSignal  syscall.Signal
	StopTimeout time.Duration
	StdoutDest  string
	StderrDest  string
	Restart     bool
	Retries     int
	ExitCode    int
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
	s.Dir = GetString(data, "dir", cwd)
	s.StopSignal = GetSignal(data, "stop-signal", DefaultServiceStopSignal)
	s.StopTimeout = GetDuration(data, "stop-timeout", DefaultServiceStopTimeout)
	s.StdoutDest = GetString(data, "stdout-dest", DefaultServiceStdoutDest)
	s.Restart = GetBool(data, "restart", DefaultServiceRestart)
	s.Retries = GetInt(data, "retries", DefaultServiceRetries)
	s.ExitCode = GetInt(data, "exit-code", DefaultServiceExitCode)

	stderrDest := GetString(data, "stderr-dest", DefaultServiceStderrDest)
	if strings.Trim(strings.ToUpper(stderrDest)) == "STDOUT" {
		s.StderrDest = "STDOUT"
	} else {
		s.StderrDest = stderrDest
	}

	if values, ok := GetMapItem(data, "env"); ok {
		AssertIsMap("env", values)
		s.Metadata = values.(map[interface{}]interface{})
	} else {
		s.Metadata = make(map[interface{}]interface{})
	}

	s.Env = make(map[string]string)
	if values, ok := GetMapItem(data, "env"); ok {
		AssertIsStringMap("env", values)
		for name, value := range values.(map[interface{}]interface{}) {
			s.Env[name.(string)] = value.(string)
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

	if len(s.Command) == 0 || strings.TrimSpace(s.Command[0]) == "" {
		msg := fmt.Sprintf("command is invalid: %+v", s.Command)
		errors = append(errors, ValidationError(msg))
	}

	if s.StopTimeout < 0 {
		msg := fmt.Sprintf("stop-timeout must be n >= 0: %d", s.StopTimeout)
		errors = append(errors, ValidationError(msg))
	}

	validateDest := func(what string, path string) {
		if fi, err := os.Stat(path); err != nil && !os.IsNotExist(err) {
			msg := fmt.Sprintf("%s is not writable: %s", what, path)
			errors = append(errors, ValidationError(msg))
		} else if s.StdoutDest != "/dev/null" && !fi.Mode().IsRegular() {
			msg := fmt.Sprintf("%s must be a regular file or /dev/null: %s", what, path)
			errors = append(errors, ValidationError(msg))
		}
	}

	validateDest("stdout-dest", s.StdoutDest)
	if s.StderrDest != "stdout" {
		validateDest("stderr-dest", s.StderrDest)
	}

	if s.Retries < 0 {
		msg := fmt.Sprintf("retries must be n >= 0: %d", s.Retries)
		errors = append(errors, ValidationError(msg))
	}

	if s.ExitCode < 0 || s.ExitCode > 255 {
		msg := fmt.Sprintf("exit-code must be 0 <= n <= 255: %d", s.Retries)
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

	return errors
}
