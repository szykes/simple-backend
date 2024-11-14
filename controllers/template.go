package controllers

import "net/http"

type template interface {
	Execute(w http.ResponseWriter, data any)
}
