package community

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"

	"github.com/pkg/errors"

	"github.com/timr/oss/fixit/web/handler"
	"github.com/timr/oss/fixit/web/layouts"
)

//go:embed templates/create.gohtml
var createTplS string

var createTpl = template.Must(template.New("create").Parse(createTplS))

type CreateData struct {
	Name      string
	Slug      string
	Latitude  string
	Longitude string
	Error     string
}

func CreateHandler(r *http.Request) (handler.Response, error) {
	if r.Method == "GET" {
		return showCreateForm(CreateData{})
	}

	return handler.BadInput([]byte("Method not allowed")), nil
}

func showCreateForm(data CreateData) (handler.Response, error) {
	var contentBuf bytes.Buffer
	err := createTpl.Execute(&contentBuf, data)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	layoutData := layouts.LayoutData{
		Title:   "Create Community",
		Content: template.HTML(contentBuf.String()),
	}

	content, err := layouts.WithGeneral(layoutData)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return handler.Ok(content), nil
}