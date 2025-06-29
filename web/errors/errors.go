package errors

import (
	"bytes"
	"embed"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/pkg/errors"

	"fixit/web/layouts"
)

//go:embed templates/*.gohtml
var templatesFS embed.FS

func getTemplates() *template.Template {
	// In development, parse templates from disk for hot reloading
	// In production, use embedded templates
	if templates, err := template.ParseGlob("web/errors/templates/*.gohtml"); err == nil {
		return templates
	}
	// Fallback to embedded templates
	return template.Must(template.ParseFS(templatesFS, "templates/*.gohtml"))
}

type ErrorData struct {
	Error     string
	ShowError bool
}

// Handle500 renders a 500 error page
func Handle500(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusInternalServerError)

	showError := os.Getenv("ENV") == "development" || os.Getenv("DEBUG") == "true"

	data := ErrorData{
		ShowError: showError,
	}

	if err != nil && showError {
		data.Error = err.Error()
	}

	content, renderErr := templatesExecute("500.gohtml", data)
	if renderErr != nil {
		log.Printf("Error rendering 500 page: %v", renderErr)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	html, layoutErr := layouts.WithGeneral(layouts.LayoutData{
		Title:   "Internal Server Error",
		Content: template.HTML(content),
	})
	if layoutErr != nil {
		log.Printf("Error applying layout to 500 page: %v", layoutErr)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(html)
}

// Handle404 renders a 404 error page
func Handle404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	content, err := templatesExecute("404.gohtml", nil)
	if err != nil {
		log.Printf("Error rendering 404 page: %v", err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	html, layoutErr := layouts.WithGeneral(layouts.LayoutData{
		Title:   "Page Not Found",
		Content: template.HTML(content),
	})
	if layoutErr != nil {
		log.Printf("Error applying layout to 404 page: %v", layoutErr)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(html)
}

// PanicRecoveryMiddleware recovers from panics and renders the 500 error page
func PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovery := recover(); recovery != nil {
				var err error
				if er, ok := recovery.(error); ok {
					err = er
				} else {
					err = errors.Errorf("non-error panic: %+v", recovery)
				}
				slog.Error("panic recovered", "method", r.Method, "path", r.URL.Path, "err", err)

				Handle500(w, r, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// NotFoundHandler is a fallback handler for 404s
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	// Only handle GET requests, others should return plain 404
	if r.Method != http.MethodGet {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	Handle404(w, r)
}

func templatesExecute(name string, data any) ([]byte, error) {
	var buf bytes.Buffer
	templates := getTemplates()
	if err := templates.ExecuteTemplate(&buf, name, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
