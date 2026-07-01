package model

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Article merepresentasikan struktur data artikel
type Article struct {
	ID         string          `json:"id"`
	Title      string          `json:"title"`
	Slug       string          `json:"slug"`
	Bucket     sql.NullString  `json:"bucket"`
	Path       sql.NullString  `json:"path"`
	Excerpt    sql.NullString  `json:"excerpt"`
	CoverImage sql.NullString  `json:"cover_image"`
	Content    json.RawMessage `json:"content"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// ArticleFilter untuk parameter filter pencarian
type ArticleFilter struct {
	Search string
	Limit  int
	Offset int
}
