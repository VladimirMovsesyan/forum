package graphql

import (
	"context"
	"github.com/VladimirMovsesyan/forum/internal/domain/model"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type repository interface {
	CreatePost(ctx context.Context, post model.Post) (*model.Post, error)
	Post(ctx context.Context, id int) (*model.Post, error)
	Posts(ctx context.Context) ([]*model.Post, error)

	CreateComment(ctx context.Context, comment model.Comment) (*model.Comment, error)
	Comments(ctx context.Context, postID int) ([]*model.Comment, error)
}
type Resolver struct {
	storage repository
}

func NewResolver(storage repository) Resolver {
	return Resolver{
		storage: storage,
	}
}
