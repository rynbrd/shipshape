package config

import (
	"io/ioutil"
	"launchpad.net/goyaml"
	"syscall"
	"time"
)

const (
	DefaultSystemEventsEnable  = true
	DefaultSystemEventsUrl     = "http://%s:9001/event"
	DefaultServiceStartTimeout = 1 * time.Second
	DefaultServiceStartRetries = 3
	DefaultServiceStopSignal   = syscall.SIGTERM
	DefaultServiceStopTimeout  = 5 * time.Second
	DefaultServiceStopRestart  = true
	DefaultServiceStdout       = "/dev/null"
	DefaultServiceStderr       = "stdout"
	DefaultServiceExitCode     = 0
)

// Validator exposes validation on a configuration object.
type Validater interface {
	Validate() []error
}

// Config is the root of the Deckhand configuration tree.
type Config struct {
	System   System
	Services map[string]Service
}

// Load the Deckhand configuration file.
func Load(file string) (config Config, err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	err = goyaml.Unmarshal(data, &config)
	return
}

// SetYAML parses the YAML tree into the object.
func (c *Config) SetYAML(tag string, data interface{}) bool {
	AssertHasKeys(data, []string{"system", "services"}, "config")

	value, _ := GetMapItem(data, "system")
	AssertIsMap("system", value)
	system := &System{}
	if system.SetYAML("!!map", value) {
		c.System = *system
	}

	values, _ := GetMapItem(data, "services")
	AssertIsMap("services", values)
	c.Services = make(map[string]Service)
	for name, value := range values.(map[interface{}]interface{}) {
		AssertIsString("name", name)
		AssertIsMap("service", value)
		service := &Service{Name: name.(string)}
		if service.SetYAML("!!map", value) {
			c.Services[name.(string)] = *service
		}
	}

	return true
}

func (c Config) Validate() []error {
	errors := make([]error, 0, 10)
	validateItem := func(item Validater) {
		newErrors := item.Validate()
		if len(newErrors) > 0 {
			errors = append(errors, newErrors...)
		}
	}

	validateItem(c.System)
	for _, service := range c.Services {
		validateItem(service)
	}
	return errors
}
