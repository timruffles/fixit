package auth

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"path/filepath"

	"github.com/aarondl/authboss/v3"
	"github.com/pkg/errors"

	"fixit/web/layouts"
)

//go:embed templates/*.gohtml
var templatesFS embed.FS
var templates = template.Must(template.ParseFS(templatesFS, "templates/*.gohtml"))

type Renderer struct{}

func NewRenderer() *Renderer {
	return &Renderer{}
}

var _ authboss.Renderer = (*Renderer)(nil)

func (r *Renderer) Load(names ...string) error {
	return nil
}

func (r *Renderer) Render(ctx context.Context, page string, data authboss.HTMLData) (output []byte, contentType string, err error) {
	templateName := filepath.Base(page) + ".gohtml"
	
	// Execute the content template
	content, err := templatesExecute(templateName, data)
	if err != nil {
		return nil, "", err
	}

	// Get the page title based on template name
	title := "Login"
	if templateName == "register.gohtml" {
		title = "Sign Up"
	}

	// Render with layout
	html, err := layouts.WithGeneral(layouts.LayoutData{
		Title:   title + " - FixIt",
		Content: template.HTML(content),
	})
	if err != nil {
		return nil, "", err
	}
	
	return html, "text/html", nil
}

func templatesExecute(name string, data any) ([]byte, error) {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, name, data); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}