package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

func RenderTemplateFile(templatePath string, data any) (string, error) {
	resolvedPath, err := resolveTemplatePath(templatePath)
	if err != nil {
		return "", err
	}

	templateContent, err := os.ReadFile(resolvedPath)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", resolvedPath, err)
	}

	tmpl, err := template.New(filepath.Base(resolvedPath)).Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", resolvedPath, err)
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", resolvedPath, err)
	}

	return rendered.String(), nil
}

func resolveTemplatePath(templatePath string) (string, error) {
	if templatePath == "" {
		return "", fmt.Errorf("template path cannot be empty")
	}

	if filepath.IsAbs(templatePath) {
		if _, err := os.Stat(templatePath); err == nil {
			return templatePath, nil
		}
		return "", fmt.Errorf("template file not found: %s", templatePath)
	}

	candidates := []string{templatePath}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, templatePath))

		for dir := cwd; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
			candidates = append(candidates, filepath.Join(dir, templatePath))
		}
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("template file not found: %s", templatePath)
}
