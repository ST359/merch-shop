package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/ST359/avito-trainee-backend-winter-2025/internal/config"
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	log.Info("starting service")
	_ = cfg
	http.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Success")
	})

	port := "8080"
	fmt.Printf("Server is running on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		slog.Error(fmt.Sprintf("Could not start server: %s\n", err))
	}
}
