package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/szykes/simple-backend/custctx"
	"github.com/szykes/simple-backend/errors"
	"github.com/szykes/simple-backend/models"
)

type Users struct {
	Templates struct {
		New            template
		SignIn         template
		ForgotPassword template
		CheckYourEmail template
		ResetPassword  template
	}
	UserService          *models.UserService
	SessionService       *models.SessionService
	PasswordResetService *models.PasswordResetService
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

func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	newUser := models.NewUser{
		Name:            r.FormValue("name"),
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		ConfirmPassword: r.FormValue("confirmPassword"),
	}
	user, err := u.UserService.Create(r.Context(), newUser)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrEmailTaken):
			err = errors.Public(err, "That email address is already associated with an account.")
		case errors.Is(err, models.ErrPwMismatch):
			err = errors.Public(err, "The given passwords do not match.")
		default:
			log.Printf("ERROR: create user: %v\n", err.Error())
		}
		u.Templates.New.Execute(w, r, newUser, err)
		return
	}

	session, err := u.SessionService.Create(r.Context(), user.ID)
	if err != nil {
		log.Printf("DEBUG: create user: %v\n", err.Error())
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	setCookie(w, CookieSessionName, session.Token)

	http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (u *Users) SignIn(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Email string
	}{
		Email: r.FormValue("email"),
	}
	u.Templates.SignIn.Execute(w, r, data)
}

func (u *Users) DoSignIn(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Email    string
		Password string
	}{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}
	user, err := u.UserService.Authenticate(r.Context(), data.Email, data.Password)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			err = errors.Public(err, "Wrong email and/or password")
		} else {
			log.Printf("ERROR: do sign in: %v\n", err.Error())
		}
		u.Templates.SignIn.Execute(w, r, data, err)
		return
	}

	session, err := u.SessionService.Create(r.Context(), user.ID)
	if err != nil {
		log.Printf("ERROR: do sign in: %v\n", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	setCookie(w, CookieSessionName, session.Token)

	http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (u *Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := custctx.User(ctx)

	fmt.Fprintf(w, "Current user: %s", user.Email)
}

func (u *Users) DoSignOut(w http.ResponseWriter, r *http.Request) {
	token, err := readCookie(r, CookieSessionName)
	if err != nil {
		log.Printf("ERROR: do sign out: %v\n", err.Error())
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	err = u.SessionService.Delete(r.Context(), token)
	if err != nil {
		log.Printf("ERROR: do sign out: %v\n", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	deleteCookie(w, CookieSessionName)
	http.Redirect(w, r, "/signin", http.StatusFound)
}

func (u *Users) ForgetPassword(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Email string
	}{
		Email: r.FormValue("email"),
	}
	u.Templates.ForgotPassword.Execute(w, r, data)
}

func (u *Users) DoForgetPassword(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Email     string
		ResetLink string
	}{
		Email: r.FormValue("email"),
	}
	pwReset, err := u.PasswordResetService.Create(r.Context(), data.Email)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			err = errors.Public(err, "Wrong email")
		} else {
			log.Printf("ERROR: do forgot password: %v\n", err.Error())
		}
		u.Templates.ForgotPassword.Execute(w, r, data, err)
		return
	}

	data.ResetLink = "http://localhost:3000/reset-password?token=" + pwReset.Token
	// TODO: here should be the emailing part

	u.Templates.CheckYourEmail.Execute(w, r, data)
}

func (u *Users) ResetPassword(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Token string
	}{
		Token: r.FormValue("token"),
	}
	u.Templates.ResetPassword.Execute(w, r, data)
}

func (u *Users) DoResetPassword(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Token           string
		Password        string
		ConfirmPassword string
	}{
		Token:           r.FormValue("token"),
		Password:        r.FormValue("newPassword"),
		ConfirmPassword: r.FormValue("confirmPassword"),
	}

	if data.Password != data.ConfirmPassword {
		err := errors.Public(nil, "The given passwords do not match.")
		u.Templates.ResetPassword.Execute(w, r, data, err)
		return
	}

	user, err := u.PasswordResetService.Consume(r.Context(), data.Token)
	if err != nil {
		log.Printf("ERROR: do reset password: %v\n", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	err = u.UserService.UpdatePassword(r.Context(), user.ID, data.Password)
	if err != nil {
		log.Printf("ERROR: do reset password: %v\n", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	session, err := u.SessionService.Create(r.Context(), user.ID)
	if err != nil {
		log.Printf("ERROR: do reset password: %v\n", err.Error())
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	setCookie(w, CookieSessionName, session.Token)
	http.Redirect(w, r, "/galleries", http.StatusFound)
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

		user, err := u.SessionService.User(r.Context(), token)
		if err != nil {
			log.Printf("ERROR: set user: %v\n", err.Error())
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
