package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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
	Name         string
	Email        string
	Password     string
	PasswordConf string
}

func (u *UserService) Create(ctx context.Context, newUser NewUser) (*User, error) {
	var user User
	user.Name = newUser.Name
	user.Email = strings.ToLower(newUser.Email)

	if newUser.Password != newUser.PasswordConf {
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
    VALUES ($1, $2, $3) RETURNING id;`,
		user.Name, user.Email, user.PasswordHash)
	err = row.Scan(&user.ID)
	if err != nil {
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
    SELECT id, password_hash FROM users WHERE email=$1;`,
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
