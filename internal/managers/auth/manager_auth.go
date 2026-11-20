package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"api-tests-template/internal/client/http/auth"
	"api-tests-template/internal/managers/auth/models"
)

// Login выполняет авторизацию пользователя и возвращает токен с данными пользователя, ожидает 200
func Login(t *testing.T, username, password string) models.LoginOkResponse {
	require.NotEmpty(t, username, "Отсутствует логин пользователя")
	require.NotEmpty(t, password, "Отсутствует пароль пользователя")

	request := models.LoginRequest{
		Email:    username,
		Password: password,
	}

	loginRawResponse := auth.HttpPostAuthLogin(t, request)
	defer loginRawResponse.Body.Close()

	respArray, errResponse := io.ReadAll(loginRawResponse.Body)
	require.NoError(t, errResponse)

	loginResponse := models.LoginOkResponse{}
	err := json.Unmarshal(respArray, &loginResponse)
	require.NoError(t, err)

	return loginResponse
}

type RegisterRequest struct {
	BirthDate string `json:"birth_date"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

// RegisterUser регистрирует нового пользователя и возвращает его JWT-токен
func RegisterUser(t *testing.T, email, password string) string {
	req := RegisterRequest{
		BirthDate: "2000-01-01",
		Email:     email,
		FirstName: "Test",
		LastName:  "User",
		Password:  password,
	}

	resp := auth.HttpPostAuthRegister(t, req)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Ожидался статус 201 при регистрации пользователя")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var loginResp models.LoginOkResponse
	require.NoError(t, json.Unmarshal(body, &loginResp))
	require.NotEmpty(t, loginResp.Token)

	return loginResp.Token
}
