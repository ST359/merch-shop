package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateToken(t *testing.T) {
	secretKey := []byte("mysecret")

	// Сценарий 1: Корректное создание токена
	token, err := createToken("testuser", secretKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Сценарий 2: Ошибка при создании токена (например, если секретный ключ пуст)
	emptySecretKey := []byte("")
	token, err = createToken("testuser", emptySecretKey)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestGetUserFromToken(t *testing.T) {
	secretKey := []byte("mysecret")

	// Сценарий 1: Корректный токен
	token, _ := createToken("testuser", secretKey)
	name, err := GetUserFromToken(token, secretKey)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", name)

	// Сценарий 2: Некорректный токен
	invalidToken := "invalid.token.string"
	name, err = GetUserFromToken(invalidToken, secretKey)
	assert.Empty(t, name)
	assert.Error(t, err)
}
func TestGetTokenFromContext(t *testing.T) {
	// Создаем тестовый контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Сценарий 1: Заголовок Authorization установлен корректно
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "Bearer mytoken")

	token, err := GetTokenFromContext(c)
	assert.NoError(t, err)
	assert.Equal(t, "mytoken", token)

	// Сценарий 2: Заголовок Authorization отсутствует
	c.Request.Header.Del("Authorization")
	token, err = GetTokenFromContext(c)
	assert.Error(t, err)
	assert.Empty(t, token)

	// Сценарий 3: Неправильный формат заголовка Authorization
	c.Request.Header.Set("Authorization", "InvalidFormat")
	token, err = GetTokenFromContext(c)
	assert.Error(t, err)
	assert.Empty(t, token)
}
func TestValidateToken(t *testing.T) {
	secretKey := []byte("mysecret")

	// Сценарий 1: Корректный токен
	token, _ := createToken("testuser", secretKey)
	claims, err := validateToken(token, secretKey)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", claims["name"])

	// Сценарий 2: Некорректный токен
	invalidToken := "invalid.token.string"
	claims, err = validateToken(invalidToken, secretKey)
	assert.Empty(t, claims)
	assert.Error(t, err)
}
