// utils/slug.go
package utility

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Regex untuk menghapus karakter non-alphanumeric
	nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]+`)

	// Regex untuk multiple hyphens
	multipleHyphenRegex = regexp.MustCompile(`-+`)
)

// GenerateSlug membuat slug dari judul
func GenerateSlug(title string) string {
	// Ubah ke lowercase
	slug := strings.ToLower(title)

	// Ganti spasi dan karakter khusus dengan hyphen
	slug = nonAlphanumericRegex.ReplaceAllString(slug, "-")

	// Hapus multiple hyphens
	slug = multipleHyphenRegex.ReplaceAllString(slug, "-")

	// Hapus leading dan trailing hyphens
	slug = strings.Trim(slug, "-")

	// Batasi panjang slug maksimal 255 karakter
	if len(slug) > 255 {
		slug = slug[:255]
	}

	return slug
}

// MakeUniqueSlug membuat slug unik dengan menambahkan suffix jika diperlukan
func MakeUniqueSlug(title string, existingSlugs map[string]bool) string {
	slug := GenerateSlug(title)

	if !existingSlugs[slug] {
		return slug
	}

	// Jika slug sudah ada, tambahkan suffix angka
	counter := 1
	for {
		newSlug := slug + "-" + fmt.Sprintf("%d", counter)
		if !existingSlugs[newSlug] {
			return newSlug
		}
		counter++
	}
}
