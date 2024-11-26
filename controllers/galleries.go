package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
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
		log.Printf("DEBUG: gallery show: %v\n", err.Error())
		return
	}

	type Image struct {
		GalleryID       int
		Filename        string
		FilenameEscaped string
	}

	data := struct {
		ID     int
		Title  string
		Images []Image
	}{
		ID:    gallery.ID,
		Title: gallery.Title,
	}

	images, err := g.GalleryService.Images(gallery.ID)
	if err != nil {
		log.Printf("ERROR: gallery show: %v\n", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	for _, image := range images {
		data.Images = append(data.Images, Image{
			GalleryID:       image.GalleryID,
			Filename:        image.Filename,
			FilenameEscaped: url.PathEscape(image.Filename),
		})
	}

	g.Templates.Show.Execute(w, r, data)
}

func (g *Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(context.Background(), w, r, userMustOwnGallery)
	if err != nil {
		log.Printf("DEBUG: gallery edit: %v\n", err.Error())
		return
	}

	type Image struct {
		GalleryID       int
		Filename        string
		FilenameEscaped string
	}
	data := struct {
		ID     int
		Title  string
		Images []Image
	}{
		ID:    gallery.ID,
		Title: gallery.Title,
	}
	images, err := g.GalleryService.Images(gallery.ID)
	if err != nil {
		log.Printf("ERROR: gallery edit: %v\n", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	for _, image := range images {
		data.Images = append(data.Images, Image{
			GalleryID:       image.GalleryID,
			Filename:        image.Filename,
			FilenameEscaped: url.PathEscape(image.Filename),
		})
	}
	g.Templates.Edit.Execute(w, r, data)
}

func (g *Galleries) Update(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(context.Background(), w, r, userMustOwnGallery)
	if err != nil {
		log.Printf("DEBUG: gallery edit: %v\n", err.Error())
		return
	}

	gallery.Title = r.FormValue("title")
	err = g.GalleryService.Update(context.TODO(), gallery)
	if err != nil {
		log.Printf("ERROR: gallery update: %v\n", err.Error())
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
		log.Printf("ERROR: gallery index: %v\n", err.Error())
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
		log.Printf("DEBUG: gallery delete: %v\n", err.Error())
		return
	}

	err = g.GalleryService.Delete(context.TODO(), gallery.ID)
	if err != nil {
		log.Printf("ERROR: gallery delete: %v\n", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (g *Galleries) Image(w http.ResponseWriter, r *http.Request) {
	filename := g.filename(r)

	galleryID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("ERROR: image: %v\n", err.Error())
		http.Error(w, "Invalid ID", http.StatusNotFound)
		return
	}

	image, err := g.GalleryService.Image(galleryID, filename)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "Image not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
		log.Printf("ERROR: image: %v\n", err.Error())
		return
	}

	http.ServeFile(w, r, image.Path)
}

func (g *Galleries) UploadImage(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(context.TODO(), w, r, userMustOwnGallery)
	if err != nil {
		log.Printf("DEBUG: upload image: %v\n", err.Error())
		return
	}

	err = r.ParseMultipartForm(5 << 20) // 5 MB
	if err != nil {
		log.Printf("ERROR: upload image: %v\n", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	fileHeaders := r.MultipartForm.File["images"]
	for _, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("ERROR: upload image: %v\n", err.Error())
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		err = g.GalleryService.CreateImage(gallery.ID, fileHeader.Filename, file)
		if err != nil {
			log.Printf("ERROR: upload image: %v\n", err.Error())
			var fileErr models.FileError
			if errors.As(err, &fileErr) {
				msg := fmt.Sprintf("%v has an invalid content type or extension. Only png, gif ang jpeg files can be uploaded", fileHeader.Filename)
				http.Error(w, msg, http.StatusBadRequest)
				return
			}
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g *Galleries) DeleteImage(w http.ResponseWriter, r *http.Request) {
	filename := g.filename(r)
	gallery, err := g.galleryByID(context.TODO(), w, r, userMustOwnGallery)
	if err != nil {
		log.Printf("DEBUG: delete image: %v\n", err.Error())
		return
	}

	err = g.GalleryService.DeleteImage(gallery.ID, filename)
	if err != nil {
		log.Printf("ERROR: delete image: %v\n", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g *Galleries) filename(r *http.Request) string {
	filename := chi.URLParam(r, "filename")
	filename = filepath.Base(filename)
	return filename
}

type galleryOpt func(http.ResponseWriter, *http.Request, *models.Gallery) error

func (g *Galleries) galleryByID(ctx context.Context, w http.ResponseWriter, r *http.Request, opts ...galleryOpt) (*models.Gallery, error) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return nil, errors.Wrap(err, "gallery by ID")
	}

	gallery, err := g.GalleryService.ByID(ctx, id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "Gallery is not found", http.StatusFound)
			return nil, errors.Wrap(err, "gallery by ID")
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return nil, errors.Wrap(err, "gallery by ID")
	}

	for _, opt := range opts {
		err = opt(w, r, gallery)
		if err != nil {
			return nil, errors.Wrap(err, "gallery by ID")
		}
	}

	return gallery, nil
}

func userMustOwnGallery(w http.ResponseWriter, r *http.Request, gallery *models.Gallery) error {
	user := custctx.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You are not allowed to edit", http.StatusForbidden)
		return errors.New("user does not have access")
	}
	return nil
}
