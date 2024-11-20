package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"

	"github.com/szykes/simple-backend/controllers"
	"github.com/szykes/simple-backend/migrations"
	"github.com/szykes/simple-backend/models"
	"github.com/szykes/simple-backend/templates"
	"github.com/szykes/simple-backend/views"
)

func main() {
	// setup DB
	cfg := models.DefaultPostgresCfg()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	// setup services
	userService := models.UserService{
		DB: db,
	}
	sessionService := models.SessionService{
		DB: db,
	}
	passwordResetService := models.PasswordResetService{
		DB: db,
	}

	// setup middleware
	userMw := controllers.UserMiddleware{
		SessionService: &sessionService,
	}

	csrfKey := []byte("ashlKfD8U8ui3xAfLk78Jh10AslKuHbH")
	csrfMw := csrf.Protect(csrfKey, csrf.Path("/"), csrf.Secure(false))

	// setup contollers
	users := controllers.Users{
		UserService:          &userService,
		SessionService:       &sessionService,
		PasswordResetService: &passwordResetService,
	}
	users.Templates.New = views.MustParseFS(templates.FS, "base.html", "signup.html")
	users.Templates.SignIn = views.MustParseFS(templates.FS, "base.html", "signin.html")
	users.Templates.ForgotPassword = views.MustParseFS(templates.FS, "base.html", "forgot-password.html")
	users.Templates.CheckYourEmail = views.MustParseFS(templates.FS, "base.html", "check-your-email.html")
	users.Templates.ResetPassword = views.MustParseFS(templates.FS, "base.html", "reset-password.html")

	// setup router
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(csrfMw)
	r.Use(userMw.SetUser)

	t := views.MustParseFS(templates.FS, "base.html", "home.html")
	r.Get("/", controllers.StaticHandler(t))

	t = views.MustParseFS(templates.FS, "base.html", "contact.html")
	r.Get("/contact", controllers.StaticHandler(t))

	t = views.MustParseFS(templates.FS, "base.html", "faq.html")
	r.Get("/faq", controllers.FAQ(t))

	r.Get("/signup", users.New)
	r.Post("/users", users.Create)
	r.Get("/signin", users.SignIn)
	r.Post("/signin", users.ProcessSignIn)
	r.Post("/signout", users.ProcessSignOut)
	r.Get("/forgot-password", users.ForgetPassword)
	r.Post("/forgot-password", users.ProcessForgetPassword)
	r.Get("/reset-password", users.ResetPassword)
	r.Post("/reset-password", users.ProcessResetPassword)

	r.Route("/users/me", func(r chi.Router) {
		r.Use(userMw.RequireUser)
		r.Get("/", users.CurrentUser)
	})

	r.Get("/joke ", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusTeapot), http.StatusTeapot)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	// start webserver
	fmt.Println("Starting server on :3000")
	err = http.ListenAndServe(":3000", r)
	if err != nil {
		panic(err)
	}
}
