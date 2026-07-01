// dto/article_dto.go
package dto

import (
	"encoding/json"
	"time"
)

// CreateArticleRequest DTO untuk membuat artikel baru
type CreateArticleRequest struct {
	Title      string  `form:"title" validate:"required,min=3,max=255"`
	Slug       string  `form:"slug" validate:"required,min=3,max=255"`
	Excerpt    *string `form:"excerpt" validate:"omitempty,max=500"`
	CoverImage *string
	Content    string `form:"content" validate:"required"`
}

// UpdateArticleRequest DTO untuk memperbarui artikel
type UpdateArticleRequest struct {
	Title   *string `form:"title" validate:"omitempty,min=3,max=255"`
	Slug    *string `form:"slug" validate:"omitempty,min=3,max=255"`
	Excerpt *string `form:"excerpt" validate:"omitempty,max=500"`
	Content *string `form:"content" validate:"omitempty"`
}

// ArticleResponse DTO untuk response artikel
type ArticleResponse struct {
	ID         string          `json:"id"`
	Title      string          `json:"title"`
	Slug       string          `json:"slug"`
	Excerpt    *string         `json:"excerpt,omitempty"`
	CoverImage *string         `json:"cover_image,omitempty"`
	Bucket     *string         `json:"bucket"`
	Key        *string         `json:"key"`
	Content    json.RawMessage `json:"content"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// ArticleListResponse DTO untuk response list artikel
type ArticleListResponse struct {
	Articles   []ArticleResponse `json:"articles"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

// ArticleFilterRequest DTO untuk filter pencarian
type ArticleFilterRequest struct {
	Search string `json:"search" query:"search"`
	Page   int    `json:"page" query:"page" validate:"min=1"`
	Limit  int    `json:"limit" query:"limit" validate:"min=1,max=100"`
}
