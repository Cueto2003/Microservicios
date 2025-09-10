package memory

import (
	
	"sync"
	"time"
	"errors"
	"context"

    "proyecto/auth-server/pkg/model"
)

var ErrNotFound = errors.New("auth user not found")


type Repository struct {
	sync.RWMutex
	data map[string]*model.AuthUser
}



func New() *Repository {
	return &Repository{
		data: map[string]*model.AuthUser{
			"oscar@example.com": {
				Email:        "oscar@example.com",
				PasswordHash: "$2a$10$abcdefghijklmnopqrstuv", // hash de prueba
				Provider:     "local",
				Role:         "user",
				CreatedAt:    time.Now(),
			},
		},
	}
}

func (r* Repository) GetHashByEmail(ctx context.Context, email string) (*model.AuthUser, error) {
    r.RLock()
    defer r.RUnlock()

    user, ok := r.data[email]
    if !ok {
        return nil, ErrNotFound
    }
    return user, nil
}

func (r *Repository) Put(_ context.Context, AuthUser *model.AuthUser) error {
	r.Lock()
	defer r.Unlock()
	r.data[AuthUser.Email] = AuthUser
	return nil
}