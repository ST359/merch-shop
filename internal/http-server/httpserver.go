package httpserver

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/ST359/avito-trainee-backend-winter-2025/internal/config"
	"github.com/ST359/avito-trainee-backend-winter-2025/internal/storage/postgres"
	"github.com/gin-gonic/gin"
)

type Storage interface {
	SendCoins(fromUser string, toUser string, amount int) error
	Buy(itemID string, user string) error
}
type APIServer struct {
	storage Storage
	log     *slog.Logger
}

func New(cfg config.Config) *APIServer {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	storage, err := postgres.New(cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBHost)
	if err != nil {
		log.Error(err.Error())
	}
	return &APIServer{storage: storage, log: log}
}
func (s *APIServer) PostApiSendCoin(c *gin.Context) {
	return
}
func (s *APIServer) PostApiAuth(c *gin.Context) {
	return
}
func (s *APIServer) PostApiBuyItem(c *gin.Context) {
	return
}
func (s *APIServer) PostApiInfo(c *gin.Context) {
	return
}

/* type ServerInterface interface {
	// Аутентификация и получение JWT-токена. При первой аутентификации пользователь создается автоматически.
	// (POST /api/auth)
	PostApiAuth(c *gin.Context)
	// Купить предмет за монеты.
	// (GET /api/buy/{item})
	GetApiBuyItem(c *gin.Context, item string)
	// Получить информацию о монетах, инвентаре и истории транзакций.
	// (GET /api/info)
	GetApiInfo(c *gin.Context)
	// Отправить монеты другому пользователю.
	// (POST /api/sendCoin)
	PostApiSendCoin(c *gin.Context)
} */

func Run() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	log.Info("starting service")
	log.Info(fmt.Sprintf("Config: %+v", cfg))
	//storage, err := postgres.New(cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBHost)
	/* 	if err != nil {
		log.Error(err.Error())
	} */
	http.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Success")
	})
	//http.HandleFunc("/api/buy/{item}", func(w http.ResponseWriter, r *http.Request) { buy.Buy(w, r, storage) })
	port := "8080"
	fmt.Printf("Server is running on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		slog.Error(fmt.Sprintf("Could not start server: %s\n", err))
	}
}
