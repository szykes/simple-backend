package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/szykes/simple-backend/custctx"
	"github.com/szykes/simple-backend/errors"
	"github.com/szykes/simple-backend/models"
)

type Galleries struct {
	Templates struct {
		New  template
		Edit template
	}
	GalleryService *models.GalleryService
}

func (g *Galleries) New(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title string
	}{
		Title: r.FormValue("title"),
	}
	g.Templates.New.Execute(w, r, data)
}

func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	data := struct {
		UserID int
		Title  string
	}{
		UserID: custctx.User(r.Context()).ID,
		Title:  r.FormValue("title"),
	}

	gallery, err := g.GalleryService.Create(context.TODO(), data.Title, data.UserID)
	if err != nil {
		g.Templates.New.Execute(w, r, data, err)
		return
	}

	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g *Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	gallery, err := g.GalleryService.ByID(context.TODO(), id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "Gallery is not found", http.StatusFound)
			return
		}
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	user := custctx.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You are not allowed to edit", http.StatusFound)
		return
	}

	data := struct {
		ID    int
		Title string
	}{
		ID:    gallery.ID,
		Title: gallery.Title,
	}
	g.Templates.Edit.Execute(w, r, data)
}

func (g *Galleries) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	gallery, err := g.GalleryService.ByID(context.TODO(), id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "Gallery is not found", http.StatusFound)
			return
		}
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	user := custctx.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You are not allowed to edit", http.StatusFound)
		return
	}

	gallery.Title = r.FormValue("title")
	err = g.GalleryService.Update(context.TODO(), gallery)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}
