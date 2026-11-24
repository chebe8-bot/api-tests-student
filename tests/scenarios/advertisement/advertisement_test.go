package advertisement

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	adClient "api-tests-template/internal/client/http/advertisement"
	"api-tests-template/internal/managers/auth"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

const (
	validPhotoPath   = "testdata/valid.png"
	invalidPhotoPath = "testdata/invalid.txt"
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
	title := fmt.Sprintf("TestProduct_%d", time.Now().UnixNano())
	description := "Описание тестового объявления для проверки API"
	fields := map[string]string{
		"title":       title,
		"description": description,
		"price":       "1500",
		"quantity":    "2",
	}

	// 1. Создание объявления
	createResp := adClient.HttpPostAdvertisement(s.T(), s.token, fields, validPhotoPath)
	defer createResp.Body.Close()
	createBody, err := io.ReadAll(createResp.Body)
	require.NoError(s.T(), err, "Ошибка чтения тела ответа при создании объявления")
	require.Equal(s.T(), http.StatusCreated, createResp.StatusCode, "Ожидался 201 Created, получено: %s", string(createBody))

	adID := gjson.GetBytes(createBody, "id").String()
	require.NotEmpty(s.T(), adID, "ID созданного объявления не должен быть пустым, ответ: %s", string(createBody))

	// 2. Получение объявления по ID и полная сверка полей
	getResp := adClient.HttpGetAdvertisementByID(s.T(), s.token, adID)
	defer getResp.Body.Close()
	getBody, err := io.ReadAll(getResp.Body)
	require.NoError(s.T(), err, "Ошибка чтения тела ответа при получении объявления")
	require.Equal(s.T(), http.StatusOK, getResp.StatusCode, "Ожидался 200 OK при GET /advertisement, ответ: %s", string(getBody))

	assert.Equal(s.T(), adID, gjson.GetBytes(getBody, "id").String(), "ID объявления не совпадает, ответ: %s", string(getBody))
	assert.Equal(s.T(), title, gjson.GetBytes(getBody, "title").String(), "Title объявления не совпадает, ответ: %s", string(getBody))
	assert.Equal(s.T(), description, gjson.GetBytes(getBody, "description").String(), "Description объявления не совпадает, ответ: %s", string(getBody))
	assert.Equal(s.T(), int64(1500), gjson.GetBytes(getBody, "price").Int(), "Price объявления не совпадает, ответ: %s", string(getBody))
	assert.Equal(s.T(), int64(2), gjson.GetBytes(getBody, "quantity").Int(), "Quantity объявления не совпадает, ответ: %s", string(getBody))

	// 3. Проверка появления созданного объявления в поиске GET /advertisements?search={title}
	var foundIDs []string
	var lastSearchStatus int
	var lastSearchBody []byte
	for attempt := 0; attempt < 3; attempt++ {
		searchResp := adClient.HttpSearchAdvertisements(s.T(), s.token, title)
		lastSearchStatus = searchResp.StatusCode
		lastSearchBody, _ = io.ReadAll(searchResp.Body)
		searchResp.Body.Close()

		if searchResp.StatusCode == http.StatusOK {
			gjson.GetBytes(lastSearchBody, "items.#.id").ForEach(func(_, val gjson.Result) bool {
				foundIDs = append(foundIDs, val.String())
				return true
			})
			if len(foundIDs) == 0 {
				gjson.GetBytes(lastSearchBody, "#.id").ForEach(func(_, val gjson.Result) bool {
					foundIDs = append(foundIDs, val.String())
					return true
				})
			}
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	require.Equal(s.T(), http.StatusOK, lastSearchStatus, "Ожидался 200 OK при поиске объявления, ответ: %s", string(lastSearchBody))
	assert.Contains(s.T(), foundIDs, adID, "Созданное объявление должно находиться через GET /advertisements?search={title}, ответ: %s", string(lastSearchBody))

	// 4. Проверка фотографий (id и url)
	photosResp := adClient.HttpGetAdvertisementPhotos(s.T(), s.token, adID)
	defer photosResp.Body.Close()
	photosBody, err := io.ReadAll(photosResp.Body)
	require.NoError(s.T(), err, "Ошибка чтения ответа при получении фото")
	require.Equal(s.T(), http.StatusOK, photosResp.StatusCode, "Ожидался 200 OK при получении фото, ответ: %s", string(photosBody))

	photoID := gjson.GetBytes(photosBody, "0.id").String()
	photoURL := gjson.GetBytes(photosBody, "0.url").String()
	if photoID == "" {
		photoID = gjson.GetBytes(photosBody, "photos.0.id").String()
		photoURL = gjson.GetBytes(photosBody, "photos.0.url").String()
	}
	assert.NotEmpty(s.T(), photoID, "ID фотографии не должен быть пустым, ответ: %s", string(photosBody))
	assert.NotEmpty(s.T(), photoURL, "URL фотографии не должен быть пустым, ответ: %s", string(photosBody))

	// 5. Удаление объявления и проверка 404
	delResp := adClient.HttpDeleteAdvertisement(s.T(), s.token, adID)
	defer delResp.Body.Close()
	require.Equal(s.T(), http.StatusNoContent, delResp.StatusCode, "Ожидался 204 No Content при удалении объявления")

	getAfterDelResp := adClient.HttpGetAdvertisementByID(s.T(), s.token, adID)
	defer getAfterDelResp.Body.Close()
	s.assertErrorStatusAndBody(getAfterDelResp, http.StatusNotFound, "при запросе удалённого объявления")
}

func (s *AdvertisementSuite) TestCreateAdvertisementNegative() {
	s.Run("Без токена авторизации -> 401", func() {
		resp := adClient.HttpPostAdvertisement(s.T(), "", map[string]string{
			"title":       "Без токена",
			"description": "Описание",
		}, validPhotoPath)
		defer resp.Body.Close()
		s.assertErrorStatusAndBody(resp, http.StatusUnauthorized, "без токена")
	})

	s.Run("Без обязательного поля title -> 400", func() {
		resp := adClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
			"description": "Нет заголовка",
		}, validPhotoPath)
		defer resp.Body.Close()
		s.assertErrorStatusAndBody(resp, http.StatusBadRequest, "без поля title")
	})

	s.Run("Пустой title -> 400", func() {
		resp := adClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
			"title":       "",
			"description": "Пустой заголовок",
		}, validPhotoPath)
		defer resp.Body.Close()
		s.assertErrorStatusAndBody(resp, http.StatusBadRequest, "при пустом title")
	})

	s.Run("Title длиннее лимита (300 символов) -> 400", func() {
		resp := adClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
			"title":       strings.Repeat("А", 300),
			"description": "Слишком длинный заголовок",
		}, validPhotoPath)
		defer resp.Body.Close()
		s.assertErrorStatusAndBody(resp, http.StatusBadRequest, "при превышении длины title")
	})

	s.Run("Строковая цена -> 400", func() {
		resp := adClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
			"title":       "Товар",
			"description": "Строковая цена",
			"price":       "not_a_number",
		}, validPhotoPath)
		defer resp.Body.Close()
		s.assertErrorStatusAndBody(resp, http.StatusBadRequest, "при строковой цене")
	})

	s.Run("Отрицательная цена -> 400", func() {
		resp := adClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
			"title":       "Товар",
			"description": "Отрицательная цена",
			"price":       "-100",
		}, validPhotoPath)
		defer resp.Body.Close()
		s.assertErrorStatusAndBody(resp, http.StatusBadRequest, "при отрицательной цене")
	})

	s.Run("Строковое количество -> 400", func() {
		resp := adClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
			"title":       "Товар",
			"description": "Строковое количество",
			"quantity":    "abc",
		}, validPhotoPath)
		defer resp.Body.Close()
		s.assertErrorStatusAndBody(resp, http.StatusBadRequest, "при строковом количестве")
	})

	s.Run("Отрицательное количество -> 400", func() {
		resp := adClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
			"title":       "Товар",
			"description": "Отрицательное количество",
			"quantity":    "-5",
		}, validPhotoPath)
		defer resp.Body.Close()
		s.assertErrorStatusAndBody(resp, http.StatusBadRequest, "при отрицательном количестве")
	})

	s.Run("Неподдерживаемый формат фото (.txt) -> 400", func() {
		resp := adClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
			"title":       "Товар с текстовым файлом",
			"description": "Неверный формат фото",
		}, invalidPhotoPath)
		defer resp.Body.Close()
		s.assertErrorStatusAndBody(resp, http.StatusBadRequest, "при загрузке неподдерживаемого формата файла")
	})

	s.Run("Более 3 фотографий (4 шт.) -> 400", func() {
		resp := adClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
			"title":       "Много фото",
			"description": "4 фото",
		}, validPhotoPath, validPhotoPath, validPhotoPath, validPhotoPath)
		defer resp.Body.Close()
		s.assertErrorStatusAndBody(resp, http.StatusBadRequest, "при загрузке более 3 фотографий")
	})
}

func (s *AdvertisementSuite) assertErrorStatusAndBody(resp *http.Response, expectedStatus int, contextMsg string) {
	body, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err, "Ошибка чтения тела ответа %s", contextMsg)
	assert.Equal(s.T(), expectedStatus, resp.StatusCode, "Ожидался статус %d %s, получено тело: %s", expectedStatus, contextMsg, string(body))
	assert.NotEmpty(s.T(), gjson.GetBytes(body, "error").String(), "Поле 'error' не должно быть пустым %s, тело: %s", contextMsg, string(body))
	assert.NotEmpty(s.T(), gjson.GetBytes(body, "message").String(), "Поле 'message' не должно быть пустым %s, тело: %s", contextMsg, string(body))
}
