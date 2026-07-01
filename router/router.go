package router

import (
	"database/sql"
	"log"
	"os"

	handler_auth "blog_server/auth/handler"
	service_auth "blog_server/auth/service"
	handlers "blog_server/blog/handler"
	"blog_server/blog/services"
	middleware "blog_server/share/middelware"
	storage "blog_server/share/s3_storage"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		allowedOrigins := map[string]bool{
			"http://127.0.0.1:5500": true,
			"http://localhost:5500": true,
			"http://127.0.0.1:8000": true, // <- TAMBAH INI
			"http://localhost:8000": true, // <- TAMBAH INI
		}

		if allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		// penting untuk preflight
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func SetupRouter(db *sql.DB) *gin.Engine {
	authService := service_auth.NewAuthService("", db)

	store, err := storage.NewStorage()
	if err != nil {
		log.Fatalf("storage has't initialize yet %v", err)
	}

	bucket := os.Getenv("B2_BUCKET")

	// Initialize service
	articleService := services.NewArticleService(db, *store, bucket)

	// Initialize hHandler
	articleHandler := handlers.NewArticleHandler(articleService)

	authHandler := handler_auth.NewAuthHandler(authService)

	// Setup router
	r := gin.Default()
	r.Use(CORSMiddleware())

	validator := middleware.NewUploadValidator()
	validator.UploadRule("cover_image", middleware.ImageRule(10, true))

	// Menggunakan Group untuk prefix "/api/v1"
	api := r.Group("/api")
	{
		// Article routes
		api.GET("/articles", articleHandler.ListArticles)
		api.POST("/articles", validator.Validate(), articleHandler.CreateArticle)

		// Di Gin, path parameter ditulis menggunakan tanda titik dua (:) bukan kurung kurawal
		api.GET("/articles/:id", articleHandler.GetArticle)
		api.PUT("/articles/:id", validator.Validate(), articleHandler.UpdateArticle)
		api.DELETE("/articles/:id", articleHandler.DeleteArticle)

		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.RefreshToken)

			// Ideally secure this with a JWT Authentication Middleware
			authGroup.POST("/change-password", authHandler.ChangePassword)
		}
	}
	return r
}
