package metadataUser

import (
	"context"
	"errors"

	model "proyecto/metadataUser/pkg"
)

//error personal 
var ErrNotFound = errors.New("Not found")

//Tipos de request (Solo Get)
type metadataUserRepository interface {
	Get(ctx context.Context, id string) (*model.MetadataUser, error)
	Put(_ context.Context, metadata *model.MetadataUser) (error)
}

type Controller struct {
	repo metadataUserRepository
}

//Constructor
func New(repo metadataUserRepository) *Controller {
	return &Controller{repo}
}


func (c *Controller) Get(ctx context.Context, id string) (*model.MetadataUser, error) {
	res, err := c.repo.Get(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return res, err
}


func (c *Controller) Put(ctx context.Context, metadata *model.MetadataUser) (*model.MetadataUser, error) {
	err := c.repo.Put(ctx, metadata)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}