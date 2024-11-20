package custctx

import (
	"context"

	"github.com/szykes/simple-backend/models"
)

type key uint

const (
	userKey key = iota
)

func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func User(ctx context.Context) *models.User {
	user, ok := ctx.Value(userKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}
