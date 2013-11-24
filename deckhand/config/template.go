package config

import (
	"fmt"
	"os"
	"path"
)

type Template struct {
	File   string
	Source string
}

// SetYAML parses the YAML tree into the configuration object.
func (t *Template) SetYAML(tag string, value interface{}) bool {
	AssertIsMap("template", value)
	for file, source := range value.(map[interface{}](interface{})) {
		AssertIsString("file", file)
		AssertIsString("source", source)
		t.File = file.(string)
		t.Source = source.(string)
		return true
	}
	panic(ParseError(fmt.Sprintf(`config template %+v cannot be parsed`, value)))
}

// Validate the configuration object.
func (t Template) Validate() []error {
	errors := make([]error, 0, 2)
	if fi, err := os.Stat(t.Source); err != nil {
		var msg string
		switch {
		case os.IsNotExist(err):
			msg = fmt.Sprintf("template source not found: %s", t.Source)
		case os.IsPermission(err):
			msg = fmt.Sprintf("template source permission denied: %s", t.Source)
		default:
			msg = fmt.Sprintf("template source error: %s: %s", err, t.Source)
		}
		errors = append(errors, ValidationError(msg))
	} else if !fi.Mode().IsRegular() {
		msg := fmt.Sprintf("template source is not a file: %s", t.Source)
		errors = append(errors, ValidationError(msg))
	}

	dir := path.Dir(t.File)
	if fi, err := os.Stat(dir); err != nil {
		var msg string
		switch {
		case os.IsNotExist(err):
			msg = fmt.Sprintf("template file directory not found: %s", t.File)
		case os.IsPermission(err):
			msg = fmt.Sprintf("template file directory permission denied: %s", t.File)
		default:
			msg = fmt.Sprintf("template file directory error: %s: %s", err, t.File)
		}
		errors = append(errors, ValidationError(msg))
	} else if !fi.Mode().IsDir() {
		msg := fmt.Sprintf("template file directory is not a directory: %s", t.File)
		errors = append(errors, ValidationError(msg))
	}
	return errors
}
