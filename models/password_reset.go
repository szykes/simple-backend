package models

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/szykes/simple-backend/rand"
)

const (
	DefaultResetDuration = 1 * time.Hour
)

type PasswordReset struct {
	ID        int
	UserID    int
	Token     string // set only when creating a new session
	TokenHash string
	ExpiresAt time.Time
}

type PasswordResetService struct {
	DB            *sql.DB
	BytesPerToken int
	Duration      time.Duration
}

func (p *PasswordResetService) Create(ctx context.Context, email string) (*PasswordReset, error) {
	email = strings.ToLower(email)

	var userID int
	row := p.DB.QueryRowContext(ctx, `
    SELECT id
    FROM users
    WHERE email = $1;
`, email)
	err := row.Scan(&userID)
	if err != nil {
		// TODO: what if the user does not exist?
		return nil, fmt.Errorf("create: %w", err)
	}

	bytesPerToken := max(p.BytesPerToken, MinBytesPerToken)
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	duration := p.Duration
	if duration == 0 {
		duration = DefaultResetDuration
	}
	pwReset := PasswordReset{
		UserID:    userID,
		Token:     token,
		TokenHash: p.hash(token),
		ExpiresAt: time.Now().Add(duration),
	}

	row = p.DB.QueryRowContext(ctx, `
    INSERT INTO password_resets (user_id, token_hash, expires_at)
    VALUES ($1, $2, $3) ON CONFLICT (user_id)
    DO UPDATE SET token_hash = $2, expires_at = $3
    RETURNING id;`,
		pwReset.UserID, pwReset.TokenHash, pwReset.ExpiresAt)
	err = row.Scan(&pwReset.ID)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	return &pwReset, nil
}

func (p *PasswordResetService) Consume(ctx context.Context, token string) (*User, error) {
	tokenHash := p.hash(token)
	var user User
	var pwReset PasswordReset
	row := p.DB.QueryRowContext(ctx, `
    SELECT password_resets.id,
      password_resets.expires_at,
      users.id,
      users.email,
      users.password_hash
    FROM password_resets
      JOIN users ON users.id = password_resets.user_id
    WHERE password_resets.token_hash = $1;`,
		tokenHash)
	err := row.Scan(&pwReset.ID, &pwReset.ExpiresAt, &user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("consume: %w", err)
	}

	if time.Now().After(pwReset.ExpiresAt) {
		return nil, fmt.Errorf("token expired: %v", token)
	}

	_, err = p.DB.ExecContext(ctx, `
    DELETE FROM password_resets
    WHERE id = $1;`,
		pwReset.ID)
	if err != nil {
		return nil, fmt.Errorf("consume: %w", err)
	}

	return &user, nil
}

func (p *PasswordResetService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}
