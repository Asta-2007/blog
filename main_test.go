package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"blog_server/blog/dto"

	"github.com/stretchr/testify/require"
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

func TestArticle(t *testing.T) {
	file, err := os.Open("/home/mrizan/Downloads/articles.csv")
	if err != nil {
		t.Fatalf("Gagal membuka file CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Melewati baris pertama (header CSV)
	_, err = reader.Read()
	if err != nil {
		t.Fatalf("Gagal membaca header CSV: %v", err)
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
			t.Fatalf("Gagal membaca baris CSV: %v", err)
		}

		// Ekstraksi data kolom
		title := record[0]
		slug := record[1]
		rawHTMLContent := record[3]

		// Menggunakan t.Run untuk membuat sub-test terisolasi per judul artikel
		t.Run(title, func(t *testing.T) {
			// Transformasi data string konten ke json.RawMessage (Logic utama tetap sama)
			jsonContentBytes, err := json.Marshal(rawHTMLContent)
			if err != nil {
				t.Fatalf("Gagal melakukan json.Marshal konten: %v", err)
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
				t.Fatalf("Gagal melakukan marshal payload: %v", err)
			}

			// Membuat HTTP request baru mengarah ke mockServer.URL
			req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(payloadBytes))
			if err != nil {
				t.Fatalf("Gagal membuat HTTP request: %v", err)
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
		})
	}
}

func TestList(t *testing.T) {
	apiURL := "http://localhost:8080/api/articles"
	u, err := url.Parse(apiURL)
	if err != nil {
		t.Fatalf("parse faild with %v", err)
	}

	query := u.Query()
	query.Set("search", "")
	query.Set("page", strconv.Itoa(1))
	query.Set("limit", strconv.Itoa(1))
	u.RawQuery = query.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		t.Fatalf("request faild with err %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("response fatal error with %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}

	var result dto.ArticleListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("faild to decode json %v", err)
	}

	fmt.Println("result", result)
}

func TestCreateArticle(t *testing.T) {
	// Prepare multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Form fields
	require.NoError(t, writer.WriteField("title", "Hello Java"))
	require.NoError(t, writer.WriteField("content", `{"blocks":[]}`))
	require.NoError(t, writer.WriteField("slug", "JAVA"))
	require.NoError(t, writer.WriteField("excerpt", "Example excerpt"))

	// Open image
	// file, err := os.Open("/home/mrizan/Pictures/friend.jpg")
	// require.NoError(t, err)
	// defer file.Close()

	// // Create file field
	// part, err := writer.CreateFormFile("cover_image", filepath.Base(file.Name()))
	// require.NoError(t, err)

	// // Copy image into multipart
	// _, err = io.Copy(part, file)
	// require.NoError(t, err)

	// IMPORTANT: Close the writer before creating the request.
	// This writes the terminating multipart boundary.
	require.NoError(t, writer.Close())

	// Create request
	req, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080/api/articles",
		body,
	)
	require.NoError(t, err)

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Send request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Read response once
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	t.Logf("Status : %d", resp.StatusCode)
	t.Logf("Body   : %s", string(respBody))

	// Fail immediately if server returned an error
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf(
			"unexpected status: %d\nresponse:\n%s",
			resp.StatusCode,
			string(respBody),
		)
	}

	// Decode JSON from the bytes already read
	var article dto.ArticleResponse
	err = json.Unmarshal(respBody, &article)
	require.NoError(t, err)

	t.Logf("Created Article: %+v", article)
}

func TestUpdateArticle(t *testing.T) {
	// Prepare multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Form fields
	require.NoError(t, writer.WriteField("title", "Hello Go"))
	require.NoError(t, writer.WriteField("content", `{"test-11":[]}`))

	// Open image
	// file, err := os.Open("/home/mrizan/Pictures/friend.jpg")
	// require.NoError(t, err)
	// defer file.Close()

	// // Create file field
	// part, err := writer.CreateFormFile("cover_image", filepath.Base(file.Name()))
	// require.NoError(t, err)

	// // Copy image into multipart
	// _, err = io.Copy(part, file)
	// require.NoError(t, err)

	// IMPORTANT: Close the writer before creating the request.
	// This writes the terminating multipart boundary.
	require.NoError(t, writer.Close())

	// Create request
	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("http://localhost:8080/api/articles/%s", "16053507324032087655685281480206"),
		body,
	)
	require.NoError(t, err)

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Send request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Read response once
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	t.Logf("Status : %d", resp.StatusCode)
	t.Logf("Body   : %s", string(respBody))

	// Fail immediately if server returned an error
	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"unexpected status: %d\nresponse:\n%s",
			resp.StatusCode,
			string(respBody),
		)
	}

	// Decode JSON from the bytes already read
	var article dto.ArticleResponse
	err = json.Unmarshal(respBody, &article)
	require.NoError(t, err)

	t.Logf("Created Article: %+v", article)
}
