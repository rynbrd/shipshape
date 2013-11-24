package config

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	UDP     = "udp"
	TCP     = "tcp"
	PortMin = 1
	PortMax = 65535
)

type Port struct {
	Number   int
	Protocol string
}

// SetYAML parses the YAML tree into the configuration object.
func (p *Port) SetYAML(tag string, value interface{}) bool {
	fail := func() {
		panic(ParseError(fmt.Sprintf(`config port %+v cannot be parsed`, value)))
	}

	AssertIsString("port", value)
	tokens := strings.SplitN(value.(string), "/", 2)
	if len(tokens) != 2 {
		fail()
	}

	var err error
	p.Number, err = strconv.Atoi(strings.TrimSpace(tokens[0]))
	if err != nil {
		fail()
	}

	p.Protocol = strings.TrimSpace(strings.ToLower(tokens[1]))
	return true
}

// Validate the configuration object.
func (p Port) Validate() []error {
	errors := make([]error, 0, 2)
	if p.Number < PortMin || p.Number > PortMax {
		msg := fmt.Sprintf("port number is invalid: %d/%s", p.Number, p.Protocol)
		errors = append(errors, ValidationError(msg))
	}
	if p.Protocol != TCP && p.Protocol != UDP {
		msg := fmt.Sprintf("port protocol is invalid: %d/%s", p.Number, p.Protocol)
		errors = append(errors, ValidationError(msg))
	}
	return []error{}
}
