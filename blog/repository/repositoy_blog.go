package repository

import (
	"context"

	"blog_server/blog/model"
)

// ArticleRepository interface untuk operasi CRUD artikel
type ArticleRepository interface {
	// Create membuat artikel baru
	Create(ctx context.Context, article *model.Article) error

	// GetByID mengambil artikel berdasarkan ID
	GetByID(ctx context.Context, id string) (*model.Article, error)

	// GetBySlug mengambil artikel berdasarkan slug
	GetBySlug(ctx context.Context, slug string) (*model.Article, error)

	// List mengambil daftar artikel dengan filter
	List(ctx context.Context, filter model.ArticleFilter) ([]*model.Article, error)

	// Update memperbarui artikel yang ada
	Update(ctx context.Context, id string, changes map[string]any) error

	// Delete menghapus artikel berdasarkan ID
	Delete(ctx context.Context, id string) error

	// Count menghitung total artikel berdasarkan filter
	Count(ctx context.Context, filter model.ArticleFilter) (int64, error)
}
