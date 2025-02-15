package httpserver

import (
	"errors"
	"log/slog"
	"os"

	"github.com/ST359/avito-trainee-backend-winter-2025/internal/config"
	"github.com/ST359/avito-trainee-backend-winter-2025/internal/storage"
	"github.com/ST359/avito-trainee-backend-winter-2025/internal/storage/postgres"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
	internalServerErrorMsg     string = "Internal server error"
	wrongPassOrUsernameErrMsg  string = "Wrong password or username"
	insufficientBalanceErrMsg  string = "Insufficient balance"
	recieverDoesNotExistErrMsg string = "User to send coins to does not exist"
	unauthorizedErrMsg         string = "Unauthorized"
	noSuchItemErrMsg           string = "Requested merch not found"
)

var (
	usernameKey   string = "username"
	authorizedKey string = "authorized"
)

type Storage interface {
	SendCoins(fromUser string, toUser string, amount int) error
	Buy(item string, user string) error
	AddUser(name, passHash string) error
	UserPassHash(name string) (string, error)
	UserExist(name string) (bool, error)
	UserInfo(user string) (*postgres.UserInfo, error)
	ItemExist(name string) (bool, error)
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
func (s *APIServer) PostApiSendCoin(ctx *gin.Context, request PostApiSendCoinRequestObject) (PostApiSendCoinResponseObject, error) {
	authorized := ctx.GetBool(authorizedKey)
	if !authorized {
		errResp := ErrorResponse{Errors: &unauthorizedErrMsg}
		return PostApiSendCoin401JSONResponse(errResp), nil
	}
	fromUser := ctx.GetString("username")
	amount, toUser := request.Body.Amount, request.Body.ToUser
	exists, err := s.storage.UserExist(toUser)
	if err != nil {
		s.log.Error(err.Error())
		errResp := ErrorResponse{Errors: &internalServerErrorMsg}
		return PostApiSendCoin500JSONResponse(errResp), err
	}
	if !exists {
		errResp := ErrorResponse{Errors: &recieverDoesNotExistErrMsg}
		return PostApiSendCoin400JSONResponse(errResp), nil
	}
	err = s.storage.SendCoins(fromUser, toUser, amount)
	if err != nil {
		if errors.Is(err, storage.ErrUnsufficientBalance) {
			errResp := ErrorResponse{Errors: &insufficientBalanceErrMsg}
			return PostApiSendCoin400JSONResponse(errResp), nil
		}
		s.log.Error(err.Error())
		errResp := ErrorResponse{Errors: &internalServerErrorMsg}
		return PostApiSendCoin500JSONResponse(errResp), err

	}
	return PostApiSendCoin200Response{}, nil
}
func (s *APIServer) PostApiAuth(ctx *gin.Context, req PostApiAuthRequestObject) (PostApiAuthResponseObject, error) {
	name, pass := req.Body.Username, req.Body.Password
	var authResp AuthResponse
	//check if user exists
	exists, err := s.storage.UserExist(name)
	if err != nil {
		s.log.Error(err.Error())
		errResp := ErrorResponse{Errors: &internalServerErrorMsg}
		return PostApiAuth500JSONResponse(errResp), err
	}
	if exists {
		passHash, err := s.storage.UserPassHash(name)
		if err != nil {
			s.log.Error(err.Error())
			errResp := ErrorResponse{Errors: &internalServerErrorMsg}
			return PostApiAuth500JSONResponse(errResp), err
		}
		err = bcrypt.CompareHashAndPassword([]byte(passHash), []byte(pass))
		if err != nil {
			errResp := ErrorResponse{Errors: &wrongPassOrUsernameErrMsg}
			return PostApiAuth401JSONResponse(errResp), nil
		}
		//return jwt token here
		token, err := createToken(name, s.jwtSecret)
		if err != nil {
			s.log.Error(err.Error())
			errResp := ErrorResponse{Errors: &internalServerErrorMsg}
			return PostApiAuth500JSONResponse(errResp), err
		}
		authResp.Token = &token
		return PostApiAuth200JSONResponse(authResp), nil
	} else {
		bPas, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			s.log.Error(err.Error())
			errResp := ErrorResponse{Errors: &internalServerErrorMsg}
			return PostApiAuth500JSONResponse(errResp), err
		}
		err = s.storage.AddUser(name, string(bPas))
		if err != nil {
			s.log.Error(err.Error())
			errResp := ErrorResponse{Errors: &internalServerErrorMsg}
			return PostApiAuth500JSONResponse(errResp), err
		}
		//return jwt token here
		token, err := createToken(name, s.jwtSecret)
		if err != nil {
			s.log.Error(err.Error())
			errResp := ErrorResponse{Errors: &internalServerErrorMsg}
			return PostApiAuth500JSONResponse(errResp), err
		}
		authResp.Token = &token
		return PostApiAuth200JSONResponse(authResp), nil
	}
}
func (s *APIServer) GetApiBuyItem(ctx *gin.Context, req GetApiBuyItemRequestObject) (GetApiBuyItemResponseObject, error) {
	authorized := ctx.GetBool(authorizedKey)
	if !authorized {
		errResp := ErrorResponse{Errors: &unauthorizedErrMsg}
		return GetApiBuyItem401JSONResponse(errResp), nil
	}

	exists, err := s.storage.ItemExist(req.Item)
	if err != nil || !exists {
		errResp := ErrorResponse{Errors: &noSuchItemErrMsg}
		return GetApiBuyItem400JSONResponse(errResp), nil
	}

	buyer := ctx.GetString(usernameKey)
	exists, err = s.storage.UserExist(buyer)
	if !exists || err != nil {
		errResp := ErrorResponse{Errors: &unauthorizedErrMsg}
		return GetApiBuyItem401JSONResponse(errResp), nil
	}
	err = s.storage.Buy(req.Item, buyer)
	if err != nil {
		if errors.Is(err, storage.ErrUnsufficientBalance) {
			errResp := ErrorResponse{Errors: &insufficientBalanceErrMsg}
			return GetApiBuyItem400JSONResponse(errResp), nil
		}
		s.log.Error(err.Error())
		errResp := ErrorResponse{Errors: &internalServerErrorMsg}
		return GetApiBuyItem500JSONResponse(errResp), nil
	}
	return GetApiBuyItem200Response{}, nil
}
func (s *APIServer) GetApiInfo(ctx *gin.Context, request GetApiInfoRequestObject) (GetApiInfoResponseObject, error) {
	authorized := ctx.GetBool(authorizedKey)
	if !authorized {
		errResp := ErrorResponse{Errors: &unauthorizedErrMsg}
		return GetApiInfo401JSONResponse(errResp), nil
	}
	name, _ := ctx.Get(usernameKey)
	exists, err := s.storage.UserExist(name.(string))
	if !exists || err != nil {
		errResp := ErrorResponse{Errors: &unauthorizedErrMsg}
		return GetApiInfo401JSONResponse(errResp), nil
	}
	dbUserInfo, err := s.storage.UserInfo(name.(string))
	if err != nil {
		s.log.Error(err.Error())
		errResp := ErrorResponse{Errors: &internalServerErrorMsg}
		return GetApiInfo500JSONResponse(errResp), nil
	}
	var respInfo InfoResponse
	respInfo.Coins = &dbUserInfo.Coins
	var inv []struct {
		Quantity *int    `json:"quantity,omitempty"`
		Type     *string `json:"type,omitempty"`
	}
	for _, entry := range dbUserInfo.Inventory {
		q := entry.Quantity
		t := entry.Type
		inv = append(inv, struct {
			Quantity *int    `json:"quantity,omitempty"`
			Type     *string `json:"type,omitempty"`
		}{
			Quantity: &q,
			Type:     &t,
		})
	}
	respInfo.Inventory = &inv
	respInfo.CoinHistory = convertCoinHistory(dbUserInfo.CoinHistory)
	return GetApiInfo200JSONResponse(respInfo), nil
}
func (s *APIServer) AuthMiddleware(f StrictHandlerFunc, operationID string) StrictHandlerFunc {
	if operationID == "PostApiAuth" {
		return f
	}

	return func(ctx *gin.Context, request interface{}) (interface{}, error) {
		token, err := GetTokenFromContext(ctx)
		if err != nil {
			s.log.Error(err.Error())
			ctx.Set(authorizedKey, false)
			return f(ctx, request)
		}
		name, err := GetUserFromToken(token, s.jwtSecret)
		if err != nil {
			s.log.Error(err.Error())
			ctx.Set(authorizedKey, false)
			return f(ctx, request)
		}
		ctx.Set(authorizedKey, true)
		ctx.Set(usernameKey, name)
		return f(ctx, request)
	}
}
func convertCoinHistory(coinHistory postgres.CoinHistory) *struct {
	Received *[]struct {
		Amount   *int    `json:"amount,omitempty"`
		FromUser *string `json:"fromUser,omitempty"`
	} `json:"received,omitempty"`
	Sent *[]struct {
		Amount *int    `json:"amount,omitempty"`
		ToUser *string `json:"toUser,omitempty"`
	} `json:"sent,omitempty"`
} {
	received := make([]struct {
		Amount   *int    `json:"amount,omitempty"`
		FromUser *string `json:"fromUser,omitempty"`
	}, len(coinHistory.Received))

	for i, transaction := range coinHistory.Received {
		amount := transaction.Amount
		fromUser := transaction.FromUser
		received[i] = struct {
			Amount   *int    `json:"amount,omitempty"`
			FromUser *string `json:"fromUser,omitempty"`
		}{
			Amount:   &amount,
			FromUser: &fromUser,
		}
	}

	sent := make([]struct {
		Amount *int    `json:"amount,omitempty"`
		ToUser *string `json:"toUser,omitempty"`
	}, len(coinHistory.Sent))

	for i, transaction := range coinHistory.Sent {
		amount := transaction.Amount
		toUser := transaction.ToUser
		sent[i] = struct {
			Amount *int    `json:"amount,omitempty"`
			ToUser *string `json:"toUser,omitempty"`
		}{
			Amount: &amount,
			ToUser: &toUser,
		}
	}

	return &struct {
		Received *[]struct {
			Amount   *int    `json:"amount,omitempty"`
			FromUser *string `json:"fromUser,omitempty"`
		} `json:"received,omitempty"`
		Sent *[]struct {
			Amount *int    `json:"amount,omitempty"`
			ToUser *string `json:"toUser,omitempty"`
		} `json:"sent,omitempty"`
	}{
		Received: &received,
		Sent:     &sent,
	}
}

func Run() {
	cfg := config.MustLoad()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	log.Info("starting service")

	s := New(cfg)
	r := gin.Default()

	handler := NewStrictHandler(s, []StrictMiddlewareFunc{s.AuthMiddleware})
	RegisterHandlers(r, handler)

	r.Run(":" + cfg.ServicePort)
}
