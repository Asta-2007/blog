package main

import (
	"blog_server/blog/dto"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type CreateArticleRequest struct {
	Title      string          `json:"title" validate:"required,min=3,max=255"`
	Slug       string          `json:"slug" validate:"required,min=3,max=255"`
	Excerpt    *string         `json:"excerpt" validate:"omitempty,max=500"`
	CoverImage *string         `json:"cover_image" validate:"omitempty,url"`
	Content    json.RawMessage `json:"content" validate:"required"`
}

func stringPtr(s string) *string {
	return &s
}
func ListArticle() {
	apiURL := "http://localhost:8080/api/articles"
	u, err := url.Parse(apiURL)

	if err != nil {
		log.Fatalf("parse faild with %v", err)
	}

	query := u.Query()
	query.Set("search", "")
	query.Set("page", strconv.Itoa(1))
	query.Set("limit", strconv.Itoa(10))
	u.RawQuery = query.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		log.Fatalf("request faild with err %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("response fatal error with %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("unexpected status code: %d", resp.StatusCode)
	}

	var result dto.ArticleListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("faild to decode json %v", err)
	}

	// BENAR (pretty print)
	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(b))
}

func CreatedArticle() {
	file, err := os.Open("/home/mrizan/Downloads/articles.csv")
	if err != nil {
		log.Fatalf("Gagal membuka file CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Melewati baris pertama (header CSV)
	_, err = reader.Read()
	if err != nil {
		log.Fatalf("Gagal membaca header CSV: %v", err)
	}

	// 3. Gunakan HTTP Client bawaan dari mockServer yang sudah dikonfigurasi otomatis
	// 3. Konfigurasi HTTP Client dengan Timeout yang aman
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	apiURL := "http://localhost:8080/api/articles"
	// 4. Iterasi/loop setiap baris data di dalam CSV untuk dijadikan sub-test (Table-Driven Style)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Gagal membaca baris CSV: %v", err)
		}

		// Ekstraksi data kolom
		title := record[0]
		slug := record[1]
		rawHTMLContent := record[3]

		// Menggunakan t.Run untuk membuat sub-test terisolasi per judul artikel
		// Transformasi data string konten ke json.RawMessage (Logic utama tetap sama)
		jsonContentBytes, err := json.Marshal(rawHTMLContent)
		if err != nil {
			log.Fatalf("Gagal melakukan json.Marshal konten: %v", err)
		}

		// Membuat excerpt otomatis
		var excerpt *string
		if len(rawHTMLContent) > 150 {
			excerpt = stringPtr(rawHTMLContent[:150] + "...")
		} else {
			excerpt = stringPtr(rawHTMLContent)
		}

		// Konstruksi struct request payload
		requestPayload := CreateArticleRequest{
			Title:      title,
			Slug:       slug,
			Excerpt:    excerpt,
			CoverImage: nil,
			Content:    json.RawMessage(jsonContentBytes),
		}

		// Ubah struct menjadi JSON Bytes
		payloadBytes, err := json.Marshal(requestPayload)
		if err != nil {
			log.Fatalf("Gagal melakukan marshal payload: %v", err)
		}

		// Membuat HTTP request baru mengarah ke mockServer.URL
		req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			log.Fatalf("Gagal membuat HTTP request: %v", err)
		}

		// Mengisi Headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer MOCK_SECRET_API_TOKEN")

		// Eksekusi HTTP Request
		fmt.Printf("Mengirim: [%s] ... ", title)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("ERROR: Gagal mengirim HTTP request: %v\n", err)
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close() //

		// Assertions (Pemeriksaan hasil) apakah status codenya sesuai ekspektasi unit test
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			fmt.Printf("SUKSES (Status: %s)\n", resp.Status)
		} else {
			fmt.Printf("GAGAL (Status: %s) -> Response: %s\n", resp.Status, string(respBody))
		}

	}
}

func main() {
	CreatedArticle()
}
