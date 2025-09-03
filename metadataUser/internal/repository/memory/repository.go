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
			"123": {ID: "123", FullName: "Lucía Gómez"},
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

func (r *Repository) Put(_ context.Context, id string, metadata *model.MetadataUser) error {
	r.Lock()
	defer r.Unlock()
	r.data[id] = metadata
	return nil
}