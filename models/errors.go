package models

import "github.com/szykes/simple-backend/errors"

var (
	ErrNotFound   = errors.New("models: no resource is found")
	ErrEmailTaken = errors.New("models: email address is already in use")
)
