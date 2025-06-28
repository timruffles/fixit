package layouts

import (
	"bytes"
	_ "embed"
	"html/template"

	"github.com/pkg/errors"
)

//go:embed general.gohtml
var genTplS string

var genTpl = template.Must(template.New("general").Parse(genTplS))

type LayoutData struct {
	Title   string
	Content template.HTML
}

func WithGeneral(dat LayoutData) ([]byte, error) {
	var out bytes.Buffer
	err := genTpl.Execute(&out, dat)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return out.Bytes(), nil
}
