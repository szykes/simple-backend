package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/szykes/simple-backend/controllers"
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

	users := controllers.Users{}
	users.Templates.New = views.MustParseFS(templates.FS, "base.html", "signup.html")
	r.Get("/signup", users.New)
	r.Post("/users", users.Create)

	r.Get("/joke ", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusTeapot), http.StatusTeapot)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
	fmt.Println("Starting server on :3000")
	err := http.ListenAndServe(":3000", r)
	if err != nil {
		panic(err)
	}
}
