package models

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"

	"github.com/szykes/simple-backend/errors"
	"github.com/szykes/simple-backend/rand"
)

const MinBytesPerToken = 32

type Session struct {
	ID        int
	UserID    int
	Token     string // set only when creating a new session
	TokenHash string
}

type SessionService struct {
	DB            *sql.DB
	BytesPerToken int
}

func (s *SessionService) Create(ctx context.Context, userID int) (*Session, error) {
	bytesPerToken := max(s.BytesPerToken, MinBytesPerToken)
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, errors.Wrap(err, "create session", "user ID", userID)
	}
	session := Session{
		UserID:    userID,
		Token:     token,
		TokenHash: s.hash(token),
	}

	row := s.DB.QueryRowContext(ctx, `
    INSERT INTO sessions (user_id, token_hash)
    VALUES ($1, $2) ON CONFLICT (user_id)
    DO UPDATE SET token_hash = $2
    RETURNING id;`,
		session.UserID, session.TokenHash)
	err = row.Scan(&session.ID)
	if err != nil {
		return nil, errors.Wrap(err, "create session", "user ID", userID)
	}
	return &session, nil
}

func (s *SessionService) User(ctx context.Context, token string) (*User, error) {
	tokenHash := s.hash(token)

	var user User
	row := s.DB.QueryRowContext(ctx, `
    SELECT users.id, users.name, users.email, users.password_hash
    FROM sessions
    JOIN users ON users.id = sessions.user_id
    WHERE sessions.token_hash = $1;`,
		tokenHash)
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, errors.Wrap(err, "create user session")
	}
	return &user, nil
}

func (s *SessionService) Delete(ctx context.Context, token string) error {
	tokenHash := s.hash(token)

	_, err := s.DB.ExecContext(ctx, `
    DELETE FROM sessions
    WHERE token_hash = $1`,
		tokenHash)
	if err != nil {
		return errors.Wrap(err, "delete session")
	}
	return nil
}

func (s *SessionService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}
