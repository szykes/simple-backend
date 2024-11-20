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

func (u *Users) ForgetPassword(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Email string
	}{
		Email: r.FormValue("email"),
	}
	u.Templates.ForgotPassword.Execute(w, r, data)
}

func (u *Users) ProcessForgetPassword(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Email     string
		ResetLink string
	}{
		Email: r.FormValue("email"),
	}
	pwReset, err := u.PasswordResetService.Create(context.Background(), data.Email)
	if err != nil {
		// TODO: what if the user does not exist?
		fmt.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	// TODO: change to localhost
	data.ResetLink = "http://192.168.1.2:3000/reset-password?token=" + pwReset.Token
	// TODO: here should be the emailing part

	u.Templates.CheckYourEmail.Execute(w, r, data)
	// 	fmt.Fprint(w, `
	// Subject: Reset your password
	// To: `+data.Email+`
	// Body: <p>To reset your passowrd, please visit the following link: <a href"`+pwReset+`">`+pwReset+`</a></p>`)
	// TODO: print this info

}

func (u *Users) ResetPassword(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Token string
	}{
		Token: r.FormValue("token"),
	}
	u.Templates.ResetPassword.Execute(w, r, data)
}

func (u *Users) ProcessResetPassword(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Token           string
		Password        string
		ConfirmPassword string
	}{
		Token:           r.FormValue("token"),
		Password:        r.FormValue("newPassword"),
		ConfirmPassword: r.FormValue("confirmPassword"),
	}

	// TODO: is this ok?
	if data.Password != data.ConfirmPassword {
		fmt.Println("reset password: mismatching password")
		return
	}

	user, err := u.PasswordResetService.Consume(context.Background(), data.Token)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	err = u.UserService.UpdatePassword(context.Background(), user.ID, data.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	session, err := u.SessionService.Create(context.Background(), user.ID)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	setCookie(w, CookieSessionName, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
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
