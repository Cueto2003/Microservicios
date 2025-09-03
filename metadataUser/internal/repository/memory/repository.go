package memory

import (
	"context"
	"sync"

	"proyecto/metadataUser/internal/repository"
	model "proyecto/metadataUser/pkg"
)

type Repository struct {
	sync.RWMutex
	data map[string]*model.MetadataUser
}

func New() *Repository {
	return &Repository{
		data: map[string]*model.MetadataUser{
			"123": {
				Email:       "123",
				FullName:    "Lucía Gómez",
				AvatarURL:   "",
				PhoneNumber: "",
				BirthDate:   "",
				LastUpdated: "",
			},
		},
	}
}


func (r *Repository) Get(_ context.Context, id string) (*model.MetadataUser, error) {
	r.RLock()
	defer r.RUnlock()

	m, ok := r.data[id]
	if !ok {
		return nil, repository.ErrNotFound
	}

	return m, nil
}

func (r *Repository) Put(_ context.Context, metadata *model.MetadataUser) error {
	r.Lock()
	defer r.Unlock()
	r.data[metadata.Email] = metadata
	return nil
}