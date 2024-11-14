package views

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

type Template struct {
	htmlTemplate *template.Template
}

func MustParseFS(fs fs.FS, pattern ...string) *Template {
	t, err := template.ParseFS(fs, pattern...)
	if err != nil {
		panic(err)
	}
	return &Template{
		htmlTemplate: t,
	}
}

func (t *Template) Execute(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := t.htmlTemplate.Execute(w, data)
	if err != err {
		log.Printf("executing template: %v", err)
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
}
