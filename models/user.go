package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/szykes/simple-backend/errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken = errors.New("models: email address is already in use")
)

type User struct {
	ID           int
	Name         string
	Email        string
	PasswordHash string
}

type UserService struct {
	DB *sql.DB
}

type NewUser struct {
	Name            string
	Email           string
	Password        string
	ConfirmPassword string
}

func (u *UserService) Create(ctx context.Context, newUser NewUser) (*User, error) {
	var user User
	user.Name = newUser.Name
	user.Email = strings.ToLower(newUser.Email)

	if newUser.Password != newUser.ConfirmPassword {
		// TODO: handling error correctly
		return nil, fmt.Errorf("create user: mismatching password")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	user.PasswordHash = string(hashedBytes)

	row := u.DB.QueryRowContext(ctx, `
    INSERT INTO users (name, email, password_hash)
    VALUES ($1, $2, $3)
    RETURNING id;`,
		user.Name, user.Email, user.PasswordHash)
	err = row.Scan(&user.ID)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == pgerrcode.UniqueViolation {
				return nil, ErrEmailTaken
			}
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

func (u *UserService) Authenticate(ctx context.Context, email, password string) (*User, error) {
	email = strings.ToLower(email)
	user := User{
		Email: email,
	}
	row := u.DB.QueryRowContext(ctx, `
    SELECT id, password_hash
    FROM users
    WHERE email=$1;`,
		email)
	err := row.Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// TODO: implement a proper error propagating solution
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	return &user, nil
}

func (u *UserService) UpdatePassword(ctx context.Context, userID int, password string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	passwordHash := string(hashedBytes)

	_, err = u.DB.ExecContext(ctx, `
    UPDATE users
    SET password_hash = $2
    WHERE id = $1;`, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}
