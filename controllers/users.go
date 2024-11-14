package controllers

import (
	"fmt"
	"net/http"
)

type Users struct {
	Templates struct {
		New template
	}
}

func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Name  string
		Email string
	}{
		Name:  r.FormValue("name"),
		Email: r.FormValue("email"),
	}
	u.Templates.New.Execute(w, data)
}

func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Name: ", r.FormValue("name"))
	fmt.Fprint(w, "Email: ", r.FormValue("email"))
	fmt.Fprint(w, "Password: ", r.FormValue("password"))
	fmt.Fprint(w, "Password conf: ", r.FormValue("confirmPassword"))
}
