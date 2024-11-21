package views

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/szykes/simple-backend/custctx"
	"github.com/szykes/simple-backend/models"
)

type publicErorr interface {
	Public() string
}

type Template struct {
	htmlTemplate *template.Template
}

func MustParseFS(fs fs.FS, patterns ...string) *Template {
	t := template.New(patterns[0])
	t = t.Funcs(template.FuncMap{
		"csrfField": func() (template.HTML, error) {
			return "", fmt.Errorf("not implemented")
		},
		"user": func() (template.HTML, error) {
			return "", fmt.Errorf("not implemented")
		},
		"errors": func() (template.HTML, error) {
			return "", fmt.Errorf("not implemented")
		},
	})

	t, err := t.ParseFS(fs, patterns...)
	if err != nil {
		// TODO: eliminate panic
		panic(err)
	}
	return &Template{
		htmlTemplate: t,
	}
}

func (t *Template) Execute(w http.ResponseWriter, r *http.Request, data any, errs ...error) {
	tpl, err := t.htmlTemplate.Clone()
	if err != nil {
		log.Printf("cloning template: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	tpl = tpl.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrf.TemplateField(r)
		},
		"user": func() *models.User {
			return custctx.User(r.Context())
		},
		"errors": func() []string {
			var errMsgs []string
			for _, err := range errs {
				var pubErr publicErorr
				if errors.As(err, &pubErr) {
					errMsgs = append(errMsgs, pubErr.Public())
				} else {
					log.Println(err)
					errMsgs = append(errMsgs, "Internal error.")
				}
			}
			return errMsgs
		},
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != err {
		log.Printf("executing template: %v", err)
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(w, &buf)
	if err != nil {
		log.Printf("writing page: %v", err)
		return
	}
}
