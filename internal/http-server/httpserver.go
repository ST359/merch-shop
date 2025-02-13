package httpserver

import (
	"context"
	"log/slog"
	"os"

	"github.com/ST359/avito-trainee-backend-winter-2025/internal/config"
	"github.com/ST359/avito-trainee-backend-winter-2025/internal/storage/postgres"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
	internalServerErrorMsg    string = "Internal server error"
	wrongPassOrUsernameErrMsg string = "Wrong password or username"
)

type Storage interface {
	SendCoins(fromUser string, toUser string, amount int) error
	Buy(itemID string, user string) error
	AddUser(name, passHash string) error
	UserPassHash(name string) (string, error)
	UserExist(name string) (bool, error)
	UserInfo(user string) (*postgres.UserInfo, error)
}
type APIServer struct {
	jwtSecret []byte
	storage   Storage
	log       *slog.Logger
}

func New(cfg *config.Config) *APIServer {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	storage, err := postgres.New(cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBHost)
	if err != nil {
		log.Error(err.Error())
	}
	return &APIServer{jwtSecret: []byte("jwtSecretKey"), storage: storage, log: log}
}
func (s *APIServer) PostApiSendCoin(ctx context.Context, request PostApiSendCoinRequestObject) (PostApiSendCoinResponseObject, error) {
	return nil, nil
}
func (s *APIServer) PostApiAuth(ctx context.Context, req PostApiAuthRequestObject) (PostApiAuthResponseObject, error) {
	name, pass := req.Body.Username, req.Body.Password
	var authResp AuthResponse
	//check if user exists
	exists, err := s.storage.UserExist(name)
	if err != nil {
		errResp := ErrorResponse{Errors: &internalServerErrorMsg}
		return PostApiAuth500JSONResponse(errResp), err
	}
	if exists {
		passHash, err := s.storage.UserPassHash(name)
		if err != nil {
			errResp := ErrorResponse{Errors: &internalServerErrorMsg}
			return PostApiAuth500JSONResponse(errResp), err
		}
		err = bcrypt.CompareHashAndPassword([]byte(passHash), []byte(pass))
		if err != nil {
			errResp := ErrorResponse{Errors: &wrongPassOrUsernameErrMsg}
			return PostApiAuth500JSONResponse(errResp), nil
		}
		//return jwt token here
		token, err := createToken(name, s.jwtSecret)
		if err != nil {
			errResp := ErrorResponse{Errors: &internalServerErrorMsg}
			return PostApiAuth500JSONResponse(errResp), err
		}
		authResp.Token = &token
		return PostApiAuth200JSONResponse(authResp), nil
	} else {
		bPas, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			errResp := ErrorResponse{Errors: &internalServerErrorMsg}
			return PostApiAuth500JSONResponse(errResp), err
		}
		err = s.storage.AddUser(name, string(bPas))
		if err != nil {
			errResp := ErrorResponse{Errors: &internalServerErrorMsg}
			return PostApiAuth500JSONResponse(errResp), err
		}
		token, err := createToken(name, s.jwtSecret)
		if err != nil {
			errResp := ErrorResponse{Errors: &internalServerErrorMsg}
			return PostApiAuth500JSONResponse(errResp), err
		}
		authResp.Token = &token
		return PostApiAuth200JSONResponse(authResp), nil
	}
}
func (s *APIServer) GetApiBuyItem(ctx context.Context, req GetApiBuyItemRequestObject) (GetApiBuyItemResponseObject, error) {
	return nil, nil
}
func (s *APIServer) GetApiInfo(ctx context.Context, request GetApiInfoRequestObject) (GetApiInfoResponseObject, error) {
	return nil, nil
}
func (s *APIServer) PostApiInfo(c *gin.Context) {
	return
}

func Run() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	log.Info("starting service")
	s := New(cfg)
	r := gin.Default()
	handler := NewStrictHandler(s, nil)
	RegisterHandlers(r, handler)
	r.Run(":" + cfg.ServicePort)
}
