package utils

import (
	"bytes"
	tpl "text/template"
)

/////////////////////////////////////////////////////////////////////

// Template interface
type Template interface {
	Context() TemplateContext
}

// TemplateContext inner Template class processor
type TemplateContext interface {
	// Sets variables in the Template context
	Set(string, interface{}) TemplateContext

	// Parse returns the parsed template as a byte slice
	Parse() []byte

	// ParseToString returns the parsed template as a string
	ParseToString() string
}

/////////////////////////////////////////////////////////////////////

type template struct {
	t *tpl.Template
}

type templateContext struct {
	ctx map[string]interface{}
	t   *tpl.Template
}

/////////////////////////////////////////////////////////////////////

// NewTemplate constructor
func NewTemplate(content string) Template {
	t, err := tpl.New("html").Parse(content)
	if err != nil {
		panic(err)
	}
	return &template{t}

}

func (t *template) Context() TemplateContext {
	return &templateContext{
		ctx: make(map[string]interface{}),
		t:   t.t,
	}
}

func (t *templateContext) Set(name string, value interface{}) TemplateContext {
	t.ctx[name] = value
	return t
}

func (t *templateContext) Parse() []byte {
	var buf bytes.Buffer
	err := t.t.Execute(&buf, t.ctx)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func (t *templateContext) ParseToString() string {
	var buf bytes.Buffer
	err := t.t.Execute(&buf, t.ctx)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
