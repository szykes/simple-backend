package models

import (
	"context"
	"database/sql"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/szykes/simple-backend/errors"
	"golang.org/x/crypto/bcrypt"
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
		return nil, errors.Wrap(ErrPwMismatch, "create user")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, "create user")
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
				return nil, errors.Wrap(ErrEmailTaken, "create user")
			}
		}
		return nil, errors.Wrap(err, "create user")
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
		return nil, errors.Wrap(err, "authenticate user")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.Wrap(err, "authenticate user")
	}
	return &user, nil
}

func (u *UserService) UpdatePassword(ctx context.Context, userID int, password string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "update password")
	}
	passwordHash := string(hashedBytes)

	_, err = u.DB.ExecContext(ctx, `
    UPDATE users
    SET password_hash = $2
    WHERE id = $1;`, userID, passwordHash)
	if err != nil {
		return errors.Wrap(err, "update password")
	}

	return nil
}
