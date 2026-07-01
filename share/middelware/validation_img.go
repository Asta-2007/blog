package middleware

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	_ "golang.org/x/image/webp"

	"github.com/gin-gonic/gin"
)

type uploadRule struct {
	// Field name in form
	FieldName string

	// Allowed extensions (e.g., [".jpg", ".png"])
	AllowedExtensions []string

	// Max file size in bytes
	MaxSize int64

	// Required MIME types (optional)
	AllowedMimeTypes []string

	// Opsional mark
	Optional bool

	// Custom validation function
	CustomValidator func(*multipart.FileHeader) error

	// Error messages
	Messages struct {
		NoFile      string
		InvalidType string
		TooLarge    string
		InvalidMime string
	}
}

type UploadedFile struct {
	FileIncluded bool
	FilePath     string
	Ext          string
	Size         int64
	MimeType     string
}

type UploadValidator struct {
	rules map[string]uploadRule
}

func NewUploadValidator() *UploadValidator {
	return &UploadValidator{
		rules: make(map[string]uploadRule),
	}
}

func (v *UploadValidator) UploadRule(fileName string, rule uploadRule) {
	// set defalult message
	if rule.Messages.NoFile == "" {
		rule.Messages.NoFile = "No file uploaded"
	}
	if rule.Messages.InvalidType == "" {
		rule.Messages.InvalidType = "Invalid file type. Allowed: %s"
	}
	if rule.Messages.TooLarge == "" {
		rule.Messages.TooLarge = "File too large (max %d MB)"
	}
	v.rules[fileName] = rule
}

func (v *UploadValidator) Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("invalid multipart form err %v", err),
			})
			return
		}

		for fileName, rule := range v.rules {

			fhs := form.File[fileName]
			if len(fhs) == 0 {
				if rule.Optional {
					// skip validation, but still set nil so handler knows

					c.Set(fileName, &UploadedFile{
						FileIncluded: false,
						FilePath:     "",
						Ext:          "",
						Size:         0,
						MimeType:     "",
					})
					continue
				}

				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": fileName + " file not found",
					"field": fileName,
				})
				return
			}

			fh := fhs[0]

			if fh.Size > rule.MaxSize {
				c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
					"error": fmt.Sprintf(rule.Messages.TooLarge, rule.MaxSize/(1024*1024)),
					"field": fileName,
				})
				return
			}

			// Extension
			ext := strings.ToLower(filepath.Ext(fh.Filename))

			if len(rule.AllowedExtensions) > 0 &&
				!slices.Contains(rule.AllowedExtensions, ext) {

				c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
					"error":   "invalid file extension",
					"field":   fileName,
					"got":     ext,
					"allowed": rule.AllowedExtensions,
				})
				return
			}

			// MIME validation
			var mimeType string
			if len(rule.AllowedMimeTypes) > 0 {
				src, err := fh.Open()
				if err != nil {
					log.Println("failed to open file", err)
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
						"error": "failed to read file",
					})
					return
				}

				buffer := make([]byte, 512)
				n, err := src.Read(buffer)
				src.Close()

				if err != nil && err != io.EOF {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"error": "failed to read file content",
					})
					return
				}

				mimeType = http.DetectContentType(buffer[:n])

				// strict check
				allowed := slices.Contains(rule.AllowedMimeTypes, mimeType)
				if !allowed {
					c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
						"error":   "invalid file content type",
						"field":   fileName,
						"got":     mimeType,
						"allowed": rule.AllowedMimeTypes,
					})
					return
				}
			}

			// Custom validation
			if rule.CustomValidator != nil {
				if err := rule.CustomValidator(fh); err != nil {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"error": err.Error(),
						"field": fileName,
					})
					return
				}
			}

			path, err := SaveToTmp(fh)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
					"field": fileName,
				})
				return
			}
			// Store ke context
			c.Set(fileName, &UploadedFile{
				FileIncluded: true,
				FilePath:     path,
				Ext:          ext,
				Size:         fh.Size,
				MimeType:     mimeType,
			})

			log.Println(
				"file validated",
				path,
				mimeType,
			)
		}

		c.Next()
	}
}

func ImageRule(maxSizeMB int64, opsional bool) uploadRule {
	return uploadRule{
		AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".gif", ".webp"},
		AllowedMimeTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
		},
		MaxSize:  maxSizeMB * 1024 * 1024,
		Optional: opsional,
		CustomValidator: func(file *multipart.FileHeader) error {
			src, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer src.Close()

			// DecodeConfig langsung dari reader
			config, format, err := image.DecodeConfig(src)
			if err != nil {
				return fmt.Errorf("invalid image file: %w", err)
			}

			switch format {
			case "jpeg", "png", "gif", "webp":
				// ok
			default:
				return fmt.Errorf("unsupported image format: %s", format)
			}

			if config.Width > 5000 || config.Height > 5000 {
				return fmt.Errorf("image dimensions too large (max 5000x5000)")
			}

			return nil
		},
	}
}

func DocumentRule(maxSizeMB int64) uploadRule {
	return uploadRule{
		AllowedExtensions: []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx"},
		AllowedMimeTypes: []string{
			"application/pdf",
			"application/msword",
			"application/vnd.openxmlformats-officedocument",
			"application/vnd.ms-excel",
			"application/vnd.ms-powerpoint",
		},
		MaxSize: maxSizeMB * 1024 * 1024,
	}
}

func VideoRule(maxSizeMB int64) uploadRule {
	return uploadRule{
		AllowedExtensions: []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv"},
		AllowedMimeTypes:  []string{"video/"},
		MaxSize:           maxSizeMB * 1024 * 1024,
	}
}

func ArchiveRule(maxSizeMB int64) uploadRule {
	return uploadRule{
		AllowedExtensions: []string{".zip", ".rar", ".7z", ".tar", ".gz"},
		AllowedMimeTypes: []string{
			"application/zip",
			"application/x-rar",
			"application/x-7z",
			"application/x-tar",
			"application/gzip",
		},
		MaxSize: maxSizeMB * 1024 * 1024,
	}
}

func GetUploadedFile(c *gin.Context, key string) (*UploadedFile, error) {
	val, exists := c.Get(key)
	if !exists {
		return nil, fmt.Errorf("file %s not found in context", key)
	}

	file, ok := val.(*UploadedFile)
	if !ok {
		return nil, fmt.Errorf("invalid file type for key %s", key)
	}

	return file, nil
}

func SaveToTmp(fileHeader *multipart.FileHeader) (string, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}

	defer src.Close()

	tmpDir := "./tmp/uploads"
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", err
	}

	tmpFile, err := os.CreateTemp(tmpDir, "upload-*")
	if err != nil {
		return "", err
	}

	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, src); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func CleanupTempFile(filePath string) {
	if filePath == "" {
		return
	}
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		log.Printf("failed to remove file %v", err)
	}
}
