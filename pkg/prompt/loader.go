package prompt

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"
)

type Prompt struct {
	tpl *template.Template
}

func (p *Prompt) Name() string {
	return p.tpl.Name()
}

func (p *Prompt) Execute(data map[string]any) (string, error) {
	sb := strings.Builder{}

	if err := p.tpl.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("execute template %q: %w", p.tpl.Name(), err)
	}

	s := sb.String()

	if strings.Contains(s, "<no value>") {
		return "", errors.New("some template arguments are not provided")
	}

	return s, nil
}

func Load(dir fs.FS, patterns ...string) (map[string]Prompt, error) {
	tpl, err := template.ParseFS(dir, patterns...)
	if err != nil {
		return nil, fmt.Errorf("parse fs templates: %w", err)
	}

	prompts := map[string]Prompt{}

	for _, t := range tpl.Templates() {
		ext := filepath.Ext(t.Name())
		name := strings.TrimSuffix(t.Name(), ext)

		prompts[name] = Prompt{tpl: t}
	}

	return prompts, nil
}
