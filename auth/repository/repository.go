package repository_auth

import (
	"context"

	model_auth "blog_server/auth/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuthEntity interface {
	Create(ctx context.Context, entity model_auth.Entity) error

	GetByEmail(ctx context.Context, email string) (*model_auth.LoginEntity, error)
	GetByID(ctx context.Context, id string) (*model_auth.Entity, error)
	GetAll(ctx context.Context, filter bson.M) ([]*model_auth.Entity, error)

	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, entity bson.M) error
}
