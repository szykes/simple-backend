package models

import (
	"context"
	"database/sql"

	"fmt"

	"github.com/szykes/simple-backend/errors"
)

type Gallery struct {
	ID     int
	UserID int
	Title  string
}

type GalleryService struct {
	DB *sql.DB
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
		return nil, fmt.Errorf("create gallery: %w", err)
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
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query gallery by id: %w", err)
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
		return nil, fmt.Errorf("query galleries by user: %w", err)
	}

	// TODO: is it ok?
	galleries := make([]Gallery, 0, 5)
	for rows.Next() {
		gallery := Gallery{
			UserID: userID,
		}
		err = rows.Scan(&gallery.ID, &gallery.Title)
		if err != nil {
			return nil, fmt.Errorf("query galleries by user: %w", err)
		}
		galleries = append(galleries, gallery)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("query galleries by user: %w", err)
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
		return fmt.Errorf("update gallery: %w", err)
	}
	return nil
}

func (g *GalleryService) Delete(ctx context.Context, id int) error {
	_, err := g.DB.ExecContext(ctx, `
    DELETE FROM galleries
    WHERE id = $1;`, id)
	if err != nil {
		return fmt.Errorf("delete gallery: %w", err)
	}
	return nil
}
