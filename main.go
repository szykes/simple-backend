package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"

	"github.com/szykes/simple-backend/controllers"
	"github.com/szykes/simple-backend/models"
	"github.com/szykes/simple-backend/templates"
	"github.com/szykes/simple-backend/views"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))

	t := views.MustParseFS(templates.FS, "base.html", "home.html")
	r.Get("/", controllers.StaticHandler(t))

	t = views.MustParseFS(templates.FS, "base.html", "contact.html")
	r.Get("/contact", controllers.StaticHandler(t))

	t = views.MustParseFS(templates.FS, "base.html", "faq.html")
	r.Get("/faq", controllers.FAQ(t))

	cfg := models.DefaultPostgresCfg()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	us := models.UserService{
		DB: db,
	}

	users := controllers.Users{
		UserService: &us,
	}
	users.Templates.New = views.MustParseFS(templates.FS, "base.html", "signup.html")
	users.Templates.SignIn = views.MustParseFS(templates.FS, "base.html", "signin.html")
	r.Get("/signup", users.New)
	r.Post("/users", users.Create)
	r.Get("/signin", users.SignIn)
	r.Post("/signin", users.ProcessSignIn)
	r.Get("/users/me", users.CurrentUser)

	t = views.MustParseFS(templates.FS, "base.html", "forgot-password.html")
	r.Get("/forgot-password", controllers.StaticHandler(t))

	r.Get("/joke ", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusTeapot), http.StatusTeapot)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	csrfKey := []byte("ashlKfD8U8ui3xAfLk78Jh10AslKuHbH")
	csrfMw := csrf.Protect(csrfKey, csrf.Secure(false))

	fmt.Println("Starting server on :3000")
	err = http.ListenAndServe(":3000", csrfMw(r))
	if err != nil {
		panic(err)
	}
}

//  	defer db.Close()
// 	err = db.Ping()
// 	if err != nil {
// 		return nil, fmt.Errorf("open %w", err)
// 	}

// _, err = db.Exec(`
// CREATE TABLE IF NOT EXISTS users (
// id SERIAL PRIMARY KEY,
// name TEXT,
// email Text UNIQUE NOT NULL
// );

// CREATE TABLE IF NOT EXISTS orders (
// id SERIAL PRIMARY KEY,
// user_id INT NOT NULL,
// amount INT,
// description TEXT
// );
// `)

// 	if err != nil {
// 		panic(err)
// 	}
