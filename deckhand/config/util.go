package config

import (
	"fmt"
	"strings"
	"syscall"
	"time"
)

// LookupSignal returns a POSIX signal by name, i.e. "SIGHUP".
func LookupSignal(name string) (signal syscall.Signal, ok bool) {
	// I cried a little when I wrote this.
	ok = true
	switch strings.TrimSpace(strings.ToUpper(name)) {
	case "SIGABRT":
		signal = syscall.SIGABRT
	case "SIGALRM":
		signal = syscall.SIGALRM
	case "SIGBUS":
		signal = syscall.SIGBUS
	case "SIGCHLD":
		signal = syscall.SIGCHLD
	case "SIGCONT":
		signal = syscall.SIGCONT
	case "SIGFPE":
		signal = syscall.SIGFPE
	case "SIGHUP":
		signal = syscall.SIGHUP
	case "SIGILL":
		signal = syscall.SIGILL
	case "SIGINT":
		signal = syscall.SIGINT
	case "SIGKILL":
		signal = syscall.SIGKILL
	case "SIGPIPE":
		signal = syscall.SIGPIPE
	case "SIGPOLL":
		signal = syscall.SIGPOLL
	case "SIGPROF":
		signal = syscall.SIGPROF
	case "SIGQUIT":
		signal = syscall.SIGQUIT
	case "SIGSEGV":
		signal = syscall.SIGSEGV
	case "SIGSTOP":
		signal = syscall.SIGSTOP
	case "SIGSYS":
		signal = syscall.SIGSYS
	case "SIGTERM":
		signal = syscall.SIGTERM
	case "SIGTRAP":
		signal = syscall.SIGTRAP
	case "SIGTSTP":
		signal = syscall.SIGTSTP
	case "SIGTTIN":
		signal = syscall.SIGTTIN
	case "SIGTTOU":
		signal = syscall.SIGTTOU
	case "SIGURG":
		signal = syscall.SIGURG
	case "SIGUSR1":
		signal = syscall.SIGUSR1
	case "SIGUSR2":
		signal = syscall.SIGUSR2
	case "SIGVTALRM":
		signal = syscall.SIGVTALRM
	case "SIGXCPU":
		signal = syscall.SIGXCPU
	case "SIGXFSZ":
		signal = syscall.SIGXFSZ
	default:
		ok = false
	}
	return
}

// GetYamlItem returns a value out of a YAML map.
func GetMapItem(data interface{}, key string) (value interface{}, ok bool) {
	value, ok = (data.(map[interface{}]interface{}))[key]
	return
}

// GetBool parses a YAML value as a boolean.
func GetBool(data interface{}, key string, dflt bool) bool {
	if value, ok := GetMapItem(data, key); ok {
		AssertIsBool(key, value)
		return value.(bool)
	}
	return dflt
}

// GetString parses a YAML value as a string.
func GetString(data interface{}, key string, dflt string) string {
	if value, ok := GetMapItem(data, key); ok {
		AssertIsString(key, value)
		return value.(string)
	}
	return dflt
}

func GetStringArray(data interface{}, key string, dflt []string) []string {
	if value, ok := GetMapItem(data, key); ok {
		AssertIsStringArray(key, value)
		array := make([]string, len(value.([]interface{})))
		for n, item := range value.([]interface{}) {
			array[n] = item.(string)
		}
		return array
	}
	return dflt
}

// GetInt parses a YAML value as an int.
func GetInt(data interface{}, key string, dflt int) int {
	if value, ok := GetMapItem(data, key); ok {
		AssertIsInt(key, value)
		return value.(int)
	}
	return dflt
}

// GetDuration parses a YAML integer (seconds) as a Duration.
func GetDuration(data interface{}, key string, dflt time.Duration) time.Duration {
	if value, ok := GetMapItem(data, key); ok {
		AssertIsInt(key, value)
		return time.Duration(value.(int)) * time.Second
	}
	return dflt
}

// GetSignal parses a YAML string as a syscall.Signal.
func GetSignal(data interface{}, key string, dflt syscall.Signal) syscall.Signal {
	if value, ok := GetMapItem(data, key); ok {
		AssertIsString(key, value)
		if signal, ok := LookupSignal(value.(string)); ok {
			return signal
		} else {
			panic(ParseError(fmt.Sprintf(`config key {%s: %s} is not a valid POSIX signal`, key, value)))
		}
	}
	return dflt
}
