// handlers/article_handler.go
package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"blog_server/blog/dto"
	"blog_server/blog/services"
	errors_blog "blog_server/share/errors"
	middleware "blog_server/share/middelware"

	"github.com/gin-gonic/gin"
)

type ArticleHandler struct {
	articleService services.ArticleService
}

func NewArticleHandler(articleService services.ArticleService) *ArticleHandler {
	return &ArticleHandler{
		articleService: articleService,
	}
}

// CreateArticle handler untuk membuat artikel
func (h *ArticleHandler) CreateArticle(c *gin.Context) {
	var req dto.CreateArticleRequest

	// Menggunakan ShouldBindJSON untuk validasi otomatis request body
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request body %v", err)})
		return
	}

	path, err := middleware.GetUploadedFile(c, "cover_image")
	if err != nil {
		log.Printf("failed to get imge from context middleware %v", err)
	}

	article, err := h.articleService.CreateArticle(c.Request.Context(), &req, path)
	if err != nil {
		log.Println("error ", err)
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, article)
}

// GetArticle handler untuk mengambil artikel (Mendukung ID UUID atau Slug)
func (h *ArticleHandler) GetArticle(c *gin.Context) {
	idStr := c.Param("id") // Mengambil path parameter di Gin

	// Support untuk get by ID atau slug
	// id, err := uuid.Parse(idStr)
	// if err != nil {
	// 	// Jika bukan UUID, anggap sebagai slug
	// 	article, err := h.articleService.GetArticleBySlug(c.Request.Context(), idStr)
	// 	if err != nil {
	// 		h.handleError(c, err)
	// 		return
	// 	}
	// 	c.JSON(http.StatusOK, article)
	// 	return
	// }

	article, err := h.articleService.GetArticleByID(c.Request.Context(), idStr)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, article)
}

// ListArticles handler untuk list artikel dengan query params
func (h *ArticleHandler) ListArticles(c *gin.Context) {
	// Menggunakan DefaultQuery untuk fallback value bawaan Gin
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	filter := &dto.ArticleFilterRequest{
		Search: c.Query("search"),
		Page:   page,
		Limit:  limit,
	}

	result, err := h.articleService.ListArticles(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateArticle handler untuk update artikel
func (h *ArticleHandler) UpdateArticle(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateArticleRequest

	err := c.ShouldBind(&req)
	if err != nil {
		log.Printf("bind error type = %T", err)
		log.Printf("bind error = %#v", err)
		log.Printf("bind error = %v", err)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	path, err := middleware.GetUploadedFile(c, "cover_image")
	fmt.Println(path)
	if err != nil {
		log.Printf("failed to get imge from context middleware %v", err)
	}

	article, err := h.articleService.UpdateArticle(c.Request.Context(), id, &req, path)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, article)
}

// DeleteArticle handler untuk delete artikel
func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
	id := c.Param("id")

	if err := h.articleService.DeleteArticle(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Artikel berhasil dihapus",
	})
}

// handleError memetakan custom error ke response JSON Gin
func (h *ArticleHandler) handleError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *errors_blog.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"error": e.Error()})
	case *errors_blog.DuplicateError:
		c.JSON(http.StatusConflict, gin.H{"error": e.Error()})
	case errors_blog.ValidationErrors:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": e,
		})
	case *errors_blog.ValidationError:
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "message": err.Error()})
	}
}
