package pongo2

import (
	"io/ioutil"
)

// Helper function which panics, if a Template couldn't
// successfully parsed. This is how you would use it:
//     var baseTemplate = pongo2.Must(pongo2.FromFile("templates/base.html"))
func Must(tpl *Template, err error) *Template {
	if err != nil {
		panic(err)
	}
	return tpl
}

// Loads  a template from string and returns a Template instance.
func FromString(tpl string) (*Template, error) {
	t, err := newTemplateString(tpl)
	return t, err
}

// Loads  a template from a filename and returns a Template instance.
// The filename must either be relative to the application's directory
// or be an absolute path.
func FromFile(filename string) (*Template, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	t, err := newTemplate(filename, string(buf))
	return t, err
}
