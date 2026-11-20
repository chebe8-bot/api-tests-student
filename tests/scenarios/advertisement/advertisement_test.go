package advertisement

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	favClient "api-tests-template/internal/client/http/favorites"
	"api-tests-template/internal/managers/auth"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

type AdvertisementSuite struct {
	suite.Suite
	token string
}

func TestAdvertisementSuite(t *testing.T) {
	suite.Run(t, new(AdvertisementSuite))
}

func (s *AdvertisementSuite) SetupSuite() {
	_ = godotenv.Load("../../../.env")
	_ = godotenv.Overload("../../../.env.override")
	loginResp := auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD"))
	s.token = loginResp.Token
}

func (s *AdvertisementSuite) TestCreateGetAndDeleteAdvertisementPositive() {
	title := fmt.Sprintf("Тестовый товар %d", time.Now().UnixNano())
	fields := map[string]string{
		"title":       title,
		"description": "Описание тестового объявления",
		"price":       "1500",
		"quantity":    "2",
	}

	// 1. Создание объявления
	createResp := favClient.HttpPostAdvertisement(s.T(), s.token, fields, 1)
	defer createResp.Body.Close()
	require.Equal(s.T(), http.StatusCreated, createResp.StatusCode)

	createBody, _ := io.ReadAll(createResp.Body)
	adID := gjson.GetBytes(createBody, "id").String()
	require.NotEmpty(s.T(), adID)

	// 2. Получение объявления по ID
	getResp := favClient.HttpGetAdvertisementByID(s.T(), s.token, adID)
	defer getResp.Body.Close()
	require.Equal(s.T(), http.StatusOK, getResp.StatusCode)

	getBody, _ := io.ReadAll(getResp.Body)
	assert.Equal(s.T(), title, gjson.GetBytes(getBody, "title").String())

	// 3. Проверка фотографий
	photosResp := favClient.HttpGetAdvertisementPhotos(s.T(), s.token, adID)
	defer photosResp.Body.Close()
	require.Equal(s.T(), http.StatusOK, photosResp.StatusCode)

	// 4. Удаление объявления и проверка 404
	delResp := favClient.HttpDeleteAdvertisement(s.T(), s.token, adID)
	defer delResp.Body.Close()
	require.Equal(s.T(), http.StatusNoContent, delResp.StatusCode)

	getAfterDelResp := favClient.HttpGetAdvertisementByID(s.T(), s.token, adID)
	defer getAfterDelResp.Body.Close()
	assert.Equal(s.T(), http.StatusNotFound, getAfterDelResp.StatusCode)
}

func (s *AdvertisementSuite) TestCreateAdvertisementNegative() {
	// Без токена -> 401
	unauthResp := favClient.HttpPostAdvertisement(s.T(), "", map[string]string{
		"title":       "Без токена",
		"description": "Описание",
	}, 1)
	defer unauthResp.Body.Close()
	assert.Equal(s.T(), http.StatusUnauthorized, unauthResp.StatusCode)

	// Без title -> 400
	noTitleResp := favClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
		"description": "Нет заголовка",
	}, 1)
	defer noTitleResp.Body.Close()
	assert.Equal(s.T(), http.StatusBadRequest, noTitleResp.StatusCode)

	// 4 фото (больше 3) -> 400
	tooManyPhotosResp := favClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
		"title":       "Много фото",
		"description": "4 фото",
	}, 4)
	defer tooManyPhotosResp.Body.Close()
	assert.Equal(s.T(), http.StatusBadRequest, tooManyPhotosResp.StatusCode)
}
