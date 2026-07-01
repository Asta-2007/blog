package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"blog_server/blog/model"
	"blog_server/blog/repository"
)

type articleRepository struct {
	db *sql.DB
}

// NewArticleRepository membuat instance baru ArticleRepository
func NewArticleRepository(db *sql.DB) repository.ArticleRepository {
	return &articleRepository{
		db: db,
	}
}

// Create membuat artikel baru
func (r *articleRepository) Create(ctx context.Context, article *model.Article) error {
	query := `
        INSERT INTO articles (id, title, slug, excerpt, cover_image, content, bucket, path, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	now := time.Now()
	if article.CreatedAt.IsZero() {
		article.CreatedAt = now
	}
	if article.UpdatedAt.IsZero() {
		article.UpdatedAt = now
	}

	// Pastikan content adalah JSON valid
	if len(article.Content) == 0 {
		article.Content = json.RawMessage("{}")
	}

	_, err := r.db.ExecContext(
		ctx, query,
		article.ID,
		article.Title,
		article.Slug,
		article.Excerpt,
		article.CoverImage,
		article.Content,
		article.Bucket,
		article.Path,
		article.CreatedAt,
		article.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal membuat artikel: %w", err)
	}

	return nil
}

// GetByID mengambil artikel berdasarkan ID
func (r *articleRepository) GetByID(ctx context.Context, id string) (*model.Article, error) {
	query := `
        SELECT id, title, slug, excerpt, cover_image, content, bucket, path, created_at, updated_at
        FROM articles
        WHERE id = ?
    `

	article := &model.Article{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&article.ID,
		&article.Title,
		&article.Slug,
		&article.Excerpt,
		&article.CoverImage,
		&article.Content,
		&article.Bucket,
		&article.Path,
		&article.CreatedAt,
		&article.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("artikel dengan ID %s tidak ditemukan", id)
	}

	if err != nil {
		return nil, fmt.Errorf("gagal mengambil artikel: %w", err)
	}

	return article, nil
}

// GetBySlug mengambil artikel berdasarkan slug
func (r *articleRepository) GetBySlug(ctx context.Context, slug string) (*model.Article, error) {
	query := `
         SELECT id, title, slug, excerpt, cover_image, content, bucket, path, created_at, updated_at
        FROM articles
        WHERE slug = ?
    `

	article := &model.Article{}
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&article.ID,
		&article.Title,
		&article.Slug,
		&article.Excerpt,
		&article.CoverImage,
		&article.Content,
		&article.Bucket,
		&article.Path,
		&article.CreatedAt,
		&article.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("artikel dengan slug '%s' tidak ditemukan", slug)
	}

	if err != nil {
		return nil, fmt.Errorf("gagal mengambil artikel: %w", err)
	}

	return article, nil
}

// List mengambil daftar artikel dengan filter
func (r *articleRepository) List(ctx context.Context, filter model.ArticleFilter) ([]*model.Article, error) {
	var conditions []string
	var args []interface{}

	if filter.Search != "" {
		conditions = append(conditions, "(title LIKE ? OR excerpt LIKE ?)")
		searchPattern := "%" + filter.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	query := `
        SELECT id, title, slug, excerpt, cover_image, content, created_at, updated_at
        FROM articles
    `

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)

		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar artikel: %w", err)
	}
	defer rows.Close()

	var articles []*model.Article
	for rows.Next() {
		article := &model.Article{}
		err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Slug,
			&article.Excerpt,
			&article.CoverImage,
			&article.Content,
			&article.CreatedAt,
			&article.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("gagal scan artikel: %w", err)
		}
		articles = append(articles, article)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterasi rows: %w", err)
	}

	return articles, nil
}

// Update memperbarui artikel yang ada
func (r *articleRepository) Update(
	ctx context.Context,
	id string,
	changes map[string]any,
) error {
	if len(changes) == 0 {
		return nil
	}

	changes["updated_at"] = time.Now()

	sets := make([]string, 0, len(changes))
	args := make([]any, 0, len(changes)+1)

	for column, value := range changes {
		sets = append(sets, column+" = ?")
		args = append(args, value)
	}

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE articles
		SET %s
		WHERE id = ?
	`, strings.Join(sets, ", "))

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("gagal memperbarui artikel: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal mendapatkan rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("artikel dengan ID %s tidak ditemukan", id)
	}

	return nil
}

// Delete menghapus artikel berdasarkan ID
func (r *articleRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM articles WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus artikel: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal mendapatkan rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("artikel dengan ID %s tidak ditemukan", id)
	}

	return nil
}

// Count menghitung total artikel berdasarkan filter
func (r *articleRepository) Count(ctx context.Context, filter model.ArticleFilter) (int64, error) {
	var conditions []string
	var args []interface{}

	if filter.Search != "" {
		conditions = append(conditions, "(title LIKE ? OR excerpt LIKE ?)")
		searchPattern := "%" + filter.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	query := `SELECT COUNT(*) FROM articles`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("gagal menghitung artikel: %w", err)
	}

	return count, nil
}
