package controllers

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/szykes/simple-backend/custctx"
	"github.com/szykes/simple-backend/errors"
	"github.com/szykes/simple-backend/models"
)

type Galleries struct {
	Templates struct {
		New   template
		Show  template
		Edit  template
		Index template
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

func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(context.Background(), w, r)
	if err != nil {
		return
	}

	data := struct {
		ID     int
		Title  string
		Images []string
	}{
		ID:    gallery.ID,
		Title: gallery.Title,
	}
	for i := 0; i < 20; i++ {
		w, h := rand.IntN(500)+200, rand.IntN(500)+200
		catImageURL := fmt.Sprintf("https://placecats.com/%d/%d", w, h)
		data.Images = append(data.Images, catImageURL)
	}
	g.Templates.Show.Execute(w, r, data)
}

func (g *Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(context.Background(), w, r, userMustOwnGallery)
	if err != nil {
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
	gallery, err := g.galleryByID(context.Background(), w, r, userMustOwnGallery)
	if err != nil {
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

func (g *Galleries) Index(w http.ResponseWriter, r *http.Request) {
	type Gallery struct {
		ID    int
		Title string
	}
	var data struct {
		Galleries []Gallery
	}

	user := custctx.User(r.Context())
	galleries, err := g.GalleryService.ByUserID(context.TODO(), user.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	for _, gallery := range galleries {
		data.Galleries = append(data.Galleries, Gallery{
			ID:    gallery.ID,
			Title: gallery.Title,
		})
	}

	g.Templates.Index.Execute(w, r, data)
}

func (g *Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(context.Background(), w, r, userMustOwnGallery)
	if err != nil {
		return
	}

	err = g.GalleryService.Delete(context.TODO(), gallery.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/galleries", http.StatusFound)
}

type galleryOpt func(http.ResponseWriter, *http.Request, *models.Gallery) error

func (g *Galleries) galleryByID(ctx context.Context, w http.ResponseWriter, r *http.Request, opts ...galleryOpt) (*models.Gallery, error) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return nil, fmt.Errorf("gallery by ID: %w", err)
	}

	gallery, err := g.GalleryService.ByID(ctx, id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "Gallery is not found", http.StatusFound)
			return nil, fmt.Errorf("gallery by ID: %w", err)
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return nil, fmt.Errorf("gallery by ID: %w", err)
	}

	for _, opt := range opts {
		err = opt(w, r, gallery)
		if err != nil {
			return nil, err
		}
	}

	return gallery, nil
}

func userMustOwnGallery(w http.ResponseWriter, r *http.Request, gallery *models.Gallery) error {
	user := custctx.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You are not allowed to edit", http.StatusForbidden)
		return fmt.Errorf("user does not have access")
	}
	return nil
}
