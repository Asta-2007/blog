package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"blog_server/database"
	"blog_server/router"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file %v", err)
	}

	db := database.ConnectDB(os.Getenv("DB_NAME"))
	// defer tetap dipanggil paling akhir untuk memastikan DB tertutup rapi
	defer func() {
		log.Println("Closing database connection...")
		db.Close()
	}()

	// Setup router Gin
	r := router.SetupRouter(db)

	// Konfigurasi HTTP Server secara eksplisit
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Channel untuk menangkap sinyal terminasi dari OS (Ctrl+C, kill, dll)
	quit := make(chan os.Signal, 1)
	// SIGINT = Ctrl+C di terminal, SIGTERM = Sinyal standard kill (Docker/Kubernetes)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Jalankan server di dalam Goroutine agar tidak memblokir penangkapan sinyal OS bawah
	go func() {
		log.Printf("Server starting on port %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v\n", err)
		}
	}()

	// Program akan tertahan di sini sampai ada sinyal masuk ke channel 'quit'
	<-quit
	log.Println("Shutdown signal received, shutting down server gracefully...")

	// Berikan toleransi waktu (timeout) bagi server untuk menyelesaikan request yang menggantung
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Mulai proses shutdown. Srv.Shutdown otomatis menolak request baru masuk,
	// dan menyelesaikan request lama yang masih berjalan hingga batasan waktu di 'ctx'.
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly.")
}
