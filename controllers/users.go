package controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/szykes/simple-backend/models"
)

type Users struct {
	Templates struct {
		New    template
		SignIn template
	}
	UserService *models.UserService
}

func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Name  string
		Email string
	}{
		Name:  r.FormValue("name"),
		Email: r.FormValue("email"),
	}
	u.Templates.New.Execute(w, r, data)
}

// TODO: introduce conxtext
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	newUser := models.NewUser{
		Name:         r.FormValue("name"),
		Email:        r.FormValue("email"),
		Password:     r.FormValue("password"),
		PasswordConf: r.FormValue("confirmPassword"),
	}
	user, err := u.UserService.Create(context.Background(), newUser)
	if err != nil {
		// TODO: proper error logging and don't use fmt.Println
		fmt.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "User created: %+v", user)
}

func (u *Users) SignIn(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Email string
	}{
		Email: r.FormValue("email"),
	}
	u.Templates.SignIn.Execute(w, r, data)
}

func (u *Users) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Email    string
		Password string
	}{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}
	user, err := u.UserService.Authenticate(context.Background(), data.Email, data.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}

	cookie := http.Cookie{
		Name:     "email",
		Value:    user.Email,
		Path:     "/",
		HttpOnly: true, // proctected from JS, no way to do XSS
	}
	http.SetCookie(w, &cookie)

	fmt.Fprintf(w, "User authenticated: %+v", user)
}

func (u *Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("email")
	if err != nil {
		fmt.Fprint(w, "The email cookie could not be read")
		return
	}

	fmt.Fprintf(w, "Email cookie: %v\n", cookie.Value)
}
