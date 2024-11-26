package models

import (
	"context"
	"database/sql"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"fmt"

	"github.com/szykes/simple-backend/errors"
)

const (
	galleriesCountForOptimization = 5
	imagesCountForOptimization    = 20
)

type Image struct {
	GalleryID int
	Path      string
	Filename  string
}

type Gallery struct {
	ID     int
	UserID int
	Title  string
}

type GalleryService struct {
	DB *sql.DB

	ImagesDir string
}

func (g *GalleryService) Create(ctx context.Context, title string, userID int) (*Gallery, error) {
	gallery := Gallery{
		Title:  title,
		UserID: userID,
	}

	row := g.DB.QueryRowContext(ctx, `
    INSERT INTO galleries (title, user_id)
    VALUES ($1, $2) RETURNING id;`,
		gallery.Title, gallery.UserID)
	err := row.Scan(&gallery.ID)
	if err != nil {
		return nil, errors.Wrap(err, "create gallery", "title", title, "user ID", userID)
	}

	return &gallery, nil
}

func (g *GalleryService) ByID(ctx context.Context, id int) (*Gallery, error) {
	gallery := Gallery{
		ID: id,
	}

	row := g.DB.QueryRowContext(ctx, `
    SELECT title, user_id
    FROM galleries
    WHERE id = $1;`,
		gallery.ID)
	err := row.Scan(&gallery.Title, &gallery.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrNotFound
		}
		return nil, errors.Wrap(err, "query gallery by id", "ID", id)
	}
	return &gallery, nil
}

func (g *GalleryService) ByUserID(ctx context.Context, userID int) ([]Gallery, error) {
	rows, err := g.DB.QueryContext(ctx, `
    SELECT id, title
    FROM galleries
    WHERE user_id = $1;`,
		userID)
	if err != nil {
		return nil, errors.Wrap(err, "gallery by user ID", "user ID", userID)
	}

	galleries := make([]Gallery, 0, galleriesCountForOptimization)
	for rows.Next() {
		gallery := Gallery{
			UserID: userID,
		}
		err = rows.Scan(&gallery.ID, &gallery.Title)
		if err != nil {
			return nil, errors.Wrap(err, "gallery by user ID", "user ID", userID)
		}
		galleries = append(galleries, gallery)
	}
	if rows.Err() != nil {
		return nil, errors.Wrap(err, "gallery by user ID", "user ID", userID)
	}
	return galleries, nil
}

func (g *GalleryService) Update(ctx context.Context, gallery *Gallery) error {
	_, err := g.DB.ExecContext(ctx, `
    UPDATE galleries
    SET title = $2
    WHERE id = $1;`,
		gallery.ID, gallery.Title)
	if err != nil {
		return errors.Wrap(err, "update gallery", "title", gallery.Title)
	}
	return nil
}

func (g *GalleryService) Delete(ctx context.Context, id int) error {
	_, err := g.DB.ExecContext(ctx, `
    DELETE FROM galleries
    WHERE id = $1;`, id)
	if err != nil {
		return errors.Wrap(err, "delete gallery", "ID", id)
	}

	err = os.RemoveAll(g.galleryDir(id))
	if err != nil {
		return errors.Wrap(err, "delete gallery", "ID", id)
	}
	return nil
}

func (g *GalleryService) Images(galleryID int) ([]Image, error) {
	globPattern := filepath.Join(g.galleryDir(galleryID), "*")
	allFiles, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, errors.Wrap(err, "retrieve images", "gallery ID", galleryID)
	}

	images := make([]Image, 0, imagesCountForOptimization)
	for _, file := range allFiles {
		if hasExtension(file, g.extensions()) {
			images = append(images, Image{
				GalleryID: galleryID,
				Path:      file,
				Filename:  filepath.Base(file),
			})
		}
	}
	return images, nil
}

func (g *GalleryService) Image(galleryID int, filename string) (Image, error) {
	imagePath := filepath.Join(g.galleryDir(galleryID), filename)
	_, err := os.Stat(imagePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Image{}, fmt.Errorf("retrieve an image: %w gallery ID: %v, filename: %v", ErrNotFound, galleryID, filename)
		}
		return Image{}, fmt.Errorf("retrieve an image: %w gallery ID: %v, filename: %v", err, galleryID, filename)
	}
	return Image{
		Filename:  filename,
		GalleryID: galleryID,
		Path:      imagePath,
	}, nil
}

func (g *GalleryService) CreateImage(galleryID int, filename string, content io.ReadSeeker) error {
	err := checkContentType(content, g.imageContentTypes())
	if err != nil {
		return errors.Wrap(err, "create image", "gallery ID", galleryID, "filename", filename)
	}

	err = checkExtension(filename, g.extensions())
	if err != nil {
		return errors.Wrap(err, "create image", "gallery ID", galleryID, "filename", filename)
	}

	galleryDir := g.galleryDir(galleryID)
	err = os.MkdirAll(galleryDir, 0755)
	if err != nil {
		return errors.Wrap(err, "create image", "gallery ID", galleryID, "filename", filename)
	}

	imagePath := filepath.Join(galleryDir, filename)
	dst, err := os.Create(imagePath)
	if err != nil {
		return errors.Wrap(err, "create image", "gallery ID", galleryID, "filename", filename)
	}
	defer dst.Close()

	_, err = io.Copy(dst, content)
	if err != nil {
		return errors.Wrap(err, "create image", "gallery ID", galleryID, "filename", filename)
	}
	return nil
}

func (g *GalleryService) DeleteImage(galleryID int, filename string) error {
	image, err := g.Image(galleryID, filename)
	if err != nil {
		return errors.Wrap(err, "delete image", "gallery ID", galleryID, "filename", filename)
	}

	err = os.Remove(image.Path)
	if err != nil {
		return errors.Wrap(err, "delete image", "gallery ID", galleryID, "filename", filename)
	}
	return nil
}

func (g *GalleryService) imageContentTypes() []string {
	return []string{"image/png", "image/jpeg", "image/gif"}
}

func (g *GalleryService) extensions() []string {
	return []string{".png", ".jpg", ".jpeg", ".gif"}
}

func (g *GalleryService) galleryDir(id int) string {
	imagesDir := g.ImagesDir
	if imagesDir == "" {
		imagesDir = "images"
	}
	return filepath.Join(imagesDir, fmt.Sprintf("gallery-%d", id))
}

func hasExtension(file string, extensions []string) bool {
	for _, ext := range extensions {
		file = strings.ToLower(file)
		ext = strings.ToLower(ext)
		if filepath.Ext(file) == ext {
			return true
		}
	}
	return false
}
