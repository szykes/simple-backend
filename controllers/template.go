package controllers

import "net/http"

type template interface {
	Execute(w http.ResponseWriter, r *http.Request, data any, errs ...error)
}
