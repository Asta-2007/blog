// services/article_service_impl.go
package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"

	"blog_server/blog/dto"
	"blog_server/blog/model"
	"blog_server/blog/repository"
	"blog_server/blog/repository/mysql"
	errors_blog "blog_server/share/errors"
	middleware "blog_server/share/middelware"
	storage "blog_server/share/s3_storage"
	utility "blog_server/share/utils"
)

type articleService struct {
	bucket      string
	store       storage.Storage
	articleRepo repository.ArticleRepository
}

// NewArticleService membuat instance baru ArticleService
func NewArticleService(db *sql.DB, store storage.Storage, bucket string) ArticleService {
	return &articleService{
		bucket:      bucket,
		store:       store,
		articleRepo: mysql.NewArticleRepository(db),
	}
}

func (s *articleService) CreateArticle(
	ctx context.Context,
	req *dto.CreateArticleRequest,
	filePath *middleware.UploadedFile,
) (*dto.ArticleResponse, error) {
	// ---------- VALIDATION ----------
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// ---------- SLUG ----------
	slug := req.Slug
	if slug == "" {
		slug = utility.GenerateSlug(req.Title)
	}

	// optional: ensure uniqueness only if needed
	// (avoid DB call if not required for your business logic)

	// ---------- BUILD ARTICLE ----------
	article := &model.Article{
		ID:    utility.UniqueID(32),
		Title: req.Title,
		Slug:  slug,
	}

	var content json.RawMessage

	// ---------- EXCERPT ----------
	if req.Excerpt != nil {
		article.Excerpt = sql.NullString{
			String: *req.Excerpt,
			Valid:  true,
		}
	}

	// ---------- CONTENT VALIDATION (early fail is OK here too) ----------
	if req.Content != "" {
		if !json.Valid([]byte(req.Content)) {
			return nil, &errors_blog.ValidationError{
				Field:   "content",
				Message: "content harus JSON valid",
			}
		}
		content = json.RawMessage(req.Content)
	} else {
		content = json.RawMessage(`{"blocks":[],"tags":[]}`)
	}

	article.Content = content

	// ---------- FILE UPLOAD ----------
	if filePath != nil && filePath.FileIncluded {

		bucket := s.bucket // inject via struct (avoid os.Getenv inside function)

		key := fmt.Sprintf("%s%s", utility.UniqueID(10), getFileExtension(filePath.MimeType))

		upload, err := s.store.UploadFile(
			ctx,
			bucket,
			key,
			filePath.FilePath,
			filePath.MimeType,
		)
		if err != nil {
			return nil, fmt.Errorf("upload cover image failed: %w", err)
		}

		article.Bucket = sql.NullString{String: upload.Bucket, Valid: true}
		article.Path = sql.NullString{String: upload.Key, Valid: true}

	}

	// ---------- SAVE ----------
	if err := s.articleRepo.Create(ctx, article); err != nil {
		return nil, fmt.Errorf("create article failed: %w", err)
	}

	return s.toArticleResponse(article), nil
}

// GetArticleByID mengambil artikel berdasarkan ID
func (s *articleService) GetArticleByID(ctx context.Context, id string) (*dto.ArticleResponse, error) {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// fmt.Println(article)

	return s.toArticleResponse(article), nil
}

// GetArticleBySlug mengambil artikel berdasarkan slug
func (s *articleService) GetArticleBySlug(ctx context.Context, slug string) (*dto.ArticleResponse, error) {
	article, err := s.articleRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return s.toArticleResponse(article), nil
}

// ListArticles mengambil daftar artikel dengan filter dan pagination
func (s *articleService) ListArticles(ctx context.Context, filter *dto.ArticleFilterRequest) (*dto.ArticleListResponse, error) {
	// Set default values
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	// Hitung offset
	offset := (filter.Page - 1) * filter.Limit

	// Siapkan filter untuk repository
	repoFilter := model.ArticleFilter{
		Search: filter.Search,
		Limit:  filter.Limit,
		Offset: offset,
	}

	// Ambil total count
	totalCount, err := s.articleRepo.Count(ctx, repoFilter)
	if err != nil {
		return nil, fmt.Errorf("gagal menghitung total artikel: %w", err)
	}

	// Ambil data artikel
	articles, err := s.articleRepo.List(ctx, repoFilter)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar artikel: %w", err)
	}

	// Konversi ke response
	var articleResponses []dto.ArticleResponse
	for _, article := range articles {
		if article.Path.String != "" {
			url, _ := s.store.GetUrl(ctx, article.Path.String, article.Bucket.String)
			article.CoverImage = sql.NullString{String: url, Valid: true}
		}
		articleResponses = append(articleResponses, *s.toArticleResponse(article))
	}

	// Hitung total halaman
	totalPages := int(math.Ceil(float64(totalCount) / float64(filter.Limit)))

	return &dto.ArticleListResponse{
		Articles:   articleResponses,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

// UpdateArticle memperbarui artikel yang ada
func (s *articleService) UpdateArticle(
	ctx context.Context,
	id string,
	req *dto.UpdateArticleRequest,
	filePath *middleware.UploadedFile,
) (*dto.ArticleResponse, error) {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	changes := map[string]any{}

	// ---------- TITLE ----------
	if req.Title != nil {
		article.Title = *req.Title
		changes["title"] = article.Title
	}

	// ---------- SLUG ----------
	if req.Slug != nil && *req.Slug != article.Slug {

		existing, err := s.articleRepo.GetBySlug(ctx, *req.Slug)
		if err == nil && existing != nil && existing.ID != id {
			return nil, &errors_blog.DuplicateError{
				Resource: "artikel",
				Field:    "slug",
				Value:    *req.Slug,
			}
		}

		article.Slug = *req.Slug
		changes["slug"] = article.Slug
	}

	// ---------- EXCERPT ----------
	if req.Excerpt != nil {
		article.Excerpt = sql.NullString{
			String: *req.Excerpt,
			Valid:  true,
		}
		changes["excerpt"] = article.Excerpt
	}

	// ---------- CONTENT ----------

	if req.Content != nil {
		var incoming map[string]any
		_ = json.Unmarshal([]byte(*req.Content), &incoming)

		var existing map[string]any
		_ = json.Unmarshal(article.Content, &existing)

		// merge
		for k, v := range incoming {
			existing[k] = v
		}

		article.Content, _ = json.Marshal(existing)
		changes["content"] = article.Content
	}

	// ---------- IMAGE ----------
	if filePath != nil && filePath.FileIncluded {

		key := fmt.Sprintf("%s%s", utility.UniqueID(10), filePath.Ext)

		uploadResult, err := s.store.UploadFile(
			ctx,
			s.bucket,
			key,
			filePath.FilePath,
			filePath.MimeType,
		)
		if err != nil {
			return nil, err
		}

		url, err := s.store.GetUrl(ctx, uploadResult.Key, uploadResult.Bucket)
		if err != nil {
			return nil, err
		}

		article.Bucket = sql.NullString{
			String: uploadResult.Bucket,
			Valid:  true,
		}

		article.Path = sql.NullString{
			String: uploadResult.Key,
			Valid:  true,
		}

		article.CoverImage = sql.NullString{
			String: url,
			Valid:  true,
		}

		changes["bucket"] = article.Bucket
		changes["path"] = article.Path
		changes["cover_image"] = article.CoverImage
	}

	if len(changes) == 0 {
		return s.toArticleResponse(article), nil
	}

	if err := s.articleRepo.Update(ctx, id, changes); err != nil {
		return nil, err
	}

	return s.toArticleResponse(article), nil
}

// DeleteArticle menghapus artikel
func (s *articleService) DeleteArticle(ctx context.Context, id string) error {
	// Cek apakah artikel ada
	user, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if user.Bucket.Valid && user.Path.Valid {
		err := s.store.DeleteFile(ctx, user.Bucket.String, user.Path.String)
		if err != nil {
			log.Printf("failed to remove imge %v", err)
		}
	}

	// Hapus artikel
	if err := s.articleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("gagal menghapus artikel: %w", err)
	}

	return nil
}

// validateCreateRequest memvalidasi request create artikel
func (s *articleService) validateCreateRequest(req *dto.CreateArticleRequest) error {
	var errors errors_blog.ValidationErrors

	// Validasi title
	if strings.TrimSpace(req.Title) == "" {
		errors = append(errors, errors_blog.ValidationError{
			Field:   "title",
			Message: "title tidak boleh kosong",
		})
	}

	if len(req.Title) > 255 {
		errors = append(errors, errors_blog.ValidationError{
			Field:   "title",
			Message: "title maksimal 255 karakter",
		})
	}

	// Validasi content
	if len(req.Content) == 0 {
		errors = append(errors, errors_blog.ValidationError{
			Field:   "content",
			Message: "content tidak boleh kosong",
		})
	}

	if !json.Valid([]byte(req.Content)) {
		errors = append(errors, errors_blog.ValidationError{
			Field:   "content",
			Message: "content harus berupa JSON yang valid",
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// toArticleResponse mengkonversi model ke response DTO
func (s *articleService) toArticleResponse(article *model.Article) *dto.ArticleResponse {
	response := &dto.ArticleResponse{
		ID:        article.ID,
		Title:     article.Title,
		Slug:      article.Slug,
		Content:   article.Content,
		CreatedAt: article.CreatedAt,
		UpdatedAt: article.UpdatedAt,
	}

	// Handle optional fields
	if article.Excerpt.Valid {
		response.Excerpt = &article.Excerpt.String
	}

	if article.CoverImage.Valid {
		response.CoverImage = &article.CoverImage.String
	}

	if article.Bucket.Valid {
		response.Bucket = &article.Bucket.String
	}

	if article.Path.Valid {
		response.Key = &article.Path.String
	}

	return response
}

func getFileExtension(mime string) string {
	switch mime {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}
