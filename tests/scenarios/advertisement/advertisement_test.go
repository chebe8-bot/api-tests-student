package advertisement

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
	title := fmt.Sprintf("TestProduct_%d", time.Now().UnixNano())
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

	// 3. Проверка появления созданного объявления в поиске GET /advertisements?search={title}
	var foundIDs []string
	var lastSearchStatus int
	for attempt := 0; attempt < 3; attempt++ {
		searchResp := favClient.HttpSearchAdvertisements(s.T(), s.token, title)
		lastSearchStatus = searchResp.StatusCode
		searchBody, _ := io.ReadAll(searchResp.Body)
		searchResp.Body.Close()

		if searchResp.StatusCode == http.StatusOK {
			gjson.GetBytes(searchBody, "items.#.id").ForEach(func(_, val gjson.Result) bool {
				foundIDs = append(foundIDs, val.String())
				return true
			})
			if len(foundIDs) == 0 {
				gjson.GetBytes(searchBody, "#.id").ForEach(func(_, val gjson.Result) bool {
					foundIDs = append(foundIDs, val.String())
					return true
				})
			}
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	require.Equal(s.T(), http.StatusOK, lastSearchStatus)
	assert.Contains(s.T(), foundIDs, adID, "Созданное объявление должно находиться через GET /advertisements?search={title}")

	// 4. Проверка фотографий
	photosResp := favClient.HttpGetAdvertisementPhotos(s.T(), s.token, adID)
	defer photosResp.Body.Close()
	require.Equal(s.T(), http.StatusOK, photosResp.StatusCode)

	// 5. Удаление объявления и проверка 404
	delResp := favClient.HttpDeleteAdvertisement(s.T(), s.token, adID)
	defer delResp.Body.Close()
	require.Equal(s.T(), http.StatusNoContent, delResp.StatusCode)

	getAfterDelResp := favClient.HttpGetAdvertisementByID(s.T(), s.token, adID)
	defer getAfterDelResp.Body.Close()
	assert.Equal(s.T(), http.StatusNotFound, getAfterDelResp.StatusCode)
	s.assertErrorBody(getAfterDelResp)
}

func (s *AdvertisementSuite) TestCreateAdvertisementNegative() {
	// 1. Без токена -> 401
	unauthResp := favClient.HttpPostAdvertisement(s.T(), "", map[string]string{
		"title":       "Без токена",
		"description": "Описание",
	}, 1)
	defer unauthResp.Body.Close()
	assert.Equal(s.T(), http.StatusUnauthorized, unauthResp.StatusCode)
	s.assertErrorBody(unauthResp)

	// 2. Без поля title -> 400
	noTitleResp := favClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
		"description": "Нет заголовка",
	}, 1)
	defer noTitleResp.Body.Close()
	assert.Equal(s.T(), http.StatusBadRequest, noTitleResp.StatusCode)
	s.assertErrorBody(noTitleResp)

	// 3. Пустой title -> 400
	emptyTitleResp := favClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
		"title":       "",
		"description": "Пустой заголовок",
	}, 1)
	defer emptyTitleResp.Body.Close()
	assert.Equal(s.T(), http.StatusBadRequest, emptyTitleResp.StatusCode)
	s.assertErrorBody(emptyTitleResp)

	// 4. Title длиннее лимита (300 символов) -> 400
	longTitleResp := favClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
		"title":       strings.Repeat("А", 300),
		"description": "Слишком длинный заголовок",
	}, 1)
	defer longTitleResp.Body.Close()
	assert.Equal(s.T(), http.StatusBadRequest, longTitleResp.StatusCode)
	s.assertErrorBody(longTitleResp)

	// 5. Строковая цена -> 400
	strPriceResp := favClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
		"title":       "Товар",
		"description": "Строковая цена",
		"price":       "not_a_number",
	}, 1)
	defer strPriceResp.Body.Close()
	assert.Equal(s.T(), http.StatusBadRequest, strPriceResp.StatusCode)
	s.assertErrorBody(strPriceResp)

	// 6. Отрицательная цена -> 400
	negPriceResp := favClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
		"title":       "Товар",
		"description": "Отрицательная цена",
		"price":       "-100",
	}, 1)
	defer negPriceResp.Body.Close()
	assert.Equal(s.T(), http.StatusBadRequest, negPriceResp.StatusCode)
	s.assertErrorBody(negPriceResp)

	// 7. Строковое количество -> 400
	strQtyResp := favClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
		"title":       "Товар",
		"description": "Строковое количество",
		"quantity":    "abc",
	}, 1)
	defer strQtyResp.Body.Close()
	assert.Equal(s.T(), http.StatusBadRequest, strQtyResp.StatusCode)
	s.assertErrorBody(strQtyResp)

	// 8. Отрицательное количество -> 400
	negQtyResp := favClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
		"title":       "Товар",
		"description": "Отрицательное количество",
		"quantity":    "-5",
	}, 1)
	defer negQtyResp.Body.Close()
	assert.Equal(s.T(), http.StatusBadRequest, negQtyResp.StatusCode)
	s.assertErrorBody(negQtyResp)

	// 9. Неподдерживаемый формат фото (.txt) -> 400
	badPhotoResp := favClient.HttpPostAdvertisementWithCustomFile(s.T(), s.token, map[string]string{
		"title":       "Товар с текстовым файлом",
		"description": "Неверный формат фото",
	}, "document.txt", []byte("это не картинка, а обычный текст"))
	defer badPhotoResp.Body.Close()
	assert.Equal(s.T(), http.StatusBadRequest, badPhotoResp.StatusCode)
	s.assertErrorBody(badPhotoResp)

	// 10. Более 3 фото (4 шт.) -> 400
	tooManyPhotosResp := favClient.HttpPostAdvertisement(s.T(), s.token, map[string]string{
		"title":       "Много фото",
		"description": "4 фото",
	}, 4)
	defer tooManyPhotosResp.Body.Close()
	assert.Equal(s.T(), http.StatusBadRequest, tooManyPhotosResp.StatusCode)
	s.assertErrorBody(tooManyPhotosResp)
}

func (s *AdvertisementSuite) assertErrorBody(resp *http.Response) {
	body, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), gjson.GetBytes(body, "error").String(), "Поле 'error' в теле ошибки не должно быть пустым: %s", string(body))
	assert.NotEmpty(s.T(), gjson.GetBytes(body, "message").String(), "Поле 'message' в теле ошибки не должно быть пустым: %s", string(body))
}
