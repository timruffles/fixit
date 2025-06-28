package auth

import (
	"bytes"
	"context"
	"html/template"
	"path/filepath"

	"github.com/aarondl/authboss/v3"
)

type Renderer struct {
	templates map[string]*template.Template
}

func NewRenderer() *Renderer {
	r := &Renderer{
		templates: make(map[string]*template.Template),
	}

	// Load login template
	loginTmpl := template.Must(template.New("login").Parse(loginTemplate))
	r.templates["login"] = loginTmpl

	// Load register template
	registerTmpl := template.Must(template.New("register").Parse(registerTemplate))
	r.templates["register"] = registerTmpl

	return r
}

var _ authboss.Renderer = (*Renderer)(nil)

func (r *Renderer) Load(names ...string) error {
	// Templates are already loaded in NewRenderer
	return nil
}

func (r *Renderer) Render(ctx context.Context, page string, data authboss.HTMLData) (output []byte, contentType string, err error) {
	templateName := filepath.Base(page)

	tmpl, ok := r.templates[templateName]
	if !ok {
		// If template not found, use a default template
		tmpl = r.templates["login"]
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, "", err
	}

	return buf.Bytes(), "text/html", nil
}
