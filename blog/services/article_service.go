package services

import (
	"context"

	"blog_server/blog/dto"
	middleware "blog_server/share/middelware"
)

// ArticleService interface untuk business logic artikel
type ArticleService interface {
	// CreateArticle membuat artikel baru
	CreateArticle(ctx context.Context, req *dto.CreateArticleRequest, filePath *middleware.UploadedFile) (*dto.ArticleResponse, error)

	// GetArticleByID mengambil artikel berdasarkan ID
	GetArticleByID(ctx context.Context, id string) (*dto.ArticleResponse, error)

	// GetArticleBySlug mengambil artikel berdasarkan slug
	GetArticleBySlug(ctx context.Context, slug string) (*dto.ArticleResponse, error)

	// ListArticles mengambil daftar artikel dengan filter dan pagination
	ListArticles(ctx context.Context, filter *dto.ArticleFilterRequest) (*dto.ArticleListResponse, error)

	// UpdateArticle memperbarui artikel yang ada
	UpdateArticle(ctx context.Context, id string, req *dto.UpdateArticleRequest, filePath *middleware.UploadedFile) (*dto.ArticleResponse, error)

	// DeleteArticle menghapus artikel
	DeleteArticle(ctx context.Context, id string) error
}
