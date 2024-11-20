package controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/szykes/simple-backend/custctx"
	"github.com/szykes/simple-backend/models"
)

type Users struct {
	Templates struct {
		New    template
		SignIn template
	}
	UserService    *models.UserService
	SessionService *models.SessionService
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
		return
	}

	session, err := u.SessionService.Create(context.Background(), user.ID)
	if err != nil {
		fmt.Println(err)
		// TODO: show warning about blocked signin
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	setCookie(w, CookieSessionName, session.Token)

	http.Redirect(w, r, "/users/me", http.StatusFound)
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
		return
	}

	session, err := u.SessionService.Create(context.Background(), user.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	setCookie(w, CookieSessionName, session.Token)

	http.Redirect(w, r, "/users/me", http.StatusFound)
}

func (u *Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := custctx.User(ctx)

	fmt.Fprintf(w, "Current user: %s", user.Email)
}

func (u *Users) ProcessSignOut(w http.ResponseWriter, r *http.Request) {
	token, err := readCookie(r, CookieSessionName)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	err = u.SessionService.Delete(context.Background(), token)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	deleteCookie(w, CookieSessionName)
	http.Redirect(w, r, "/signin", http.StatusFound)
}

type UserMiddleware struct {
	SessionService *models.SessionService
}

func (u *UserMiddleware) SetUser(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := readCookie(r, CookieSessionName)
		if err != nil {
			handler.ServeHTTP(w, r)
			return
		}

		user, err := u.SessionService.User(context.Background(), token)
		if err != nil {
			fmt.Println(err)
			handler.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		ctx = custctx.WithUser(ctx, user)
		r = r.WithContext(ctx)
		handler.ServeHTTP(w, r)
	})
}

func (u *UserMiddleware) RequireUser(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := custctx.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}
		handler.ServeHTTP(w, r)
	})
}
