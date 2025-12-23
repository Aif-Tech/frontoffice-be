package mail

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
)

type TemplateRenderer struct {
	templates map[string]*template.Template
}

func NewTemplateRenderer(templateDir string) (*TemplateRenderer, error) {
	tmplMap := make(map[string]*template.Template)

	files, err := filepath.Glob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		name := filepath.Base(file)

		tmpl, err := template.ParseFiles(
			filepath.Join(templateDir, "base.html"),
			file,
		)
		if err != nil {
			return nil, fmt.Errorf("parse template %s: %w", name, err)
		}

		tmplMap[name] = tmpl
	}

	return &TemplateRenderer{
		templates: tmplMap,
	}, nil
}

func (r *TemplateRenderer) Render(
	templateName string,
	data any,
) (string, error) {
	tmpl, ok := r.templates[templateName]
	if !ok {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
