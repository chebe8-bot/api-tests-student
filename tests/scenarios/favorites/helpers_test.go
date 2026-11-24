package favorites

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	adClient "api-tests-template/internal/client/http/advertisement"
	favClient "api-tests-template/internal/client/http/favorites"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const validPhotoForFavorites = "../advertisement/testdata/valid.png"

// createTestAd создаёт тестовое объявление от имени переданного токена и возвращает его ID
func createTestAd(t *testing.T, ownerToken, titlePrefix string) string {
	t.Helper()
	resp := adClient.HttpPostAdvertisement(t, ownerToken, map[string]string{
		"title":       fmt.Sprintf("%s_%d", titlePrefix, time.Now().UnixNano()),
		"description": "Тестовое объявление для проверки избранного",
		"price":       "1000",
	}, validPhotoForFavorites)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Ошибка чтения ответа при создании тестового объявления")
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Ожидался статус 201 Created при создании объявления, тело: %s", string(body))

	adID := gjson.GetBytes(body, "id").String()
	require.NotEmpty(t, adID, "ID созданного объявления не должен быть пустым, тело: %s", string(body))
	return adID
}

// deleteTestAd удаляет тестовое объявление от имени владельца
func deleteTestAd(t *testing.T, ownerToken, adID string) {
	t.Helper()
	resp := adClient.HttpDeleteAdvertisement(t, ownerToken, adID)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusNoContent, resp.StatusCode, "Ожидался статус 204 No Content при удалении объявления, тело: %s", string(body))
}

// getFavoriteAdIDs возвращает список ID объявлений из избранного пользователя
func getFavoriteAdIDs(t *testing.T, token string) ([]string, string) {
	t.Helper()
	resp := favClient.HttpGetFavorites(t, token)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Ошибка чтения списка избранного")
	require.Equal(t, http.StatusOK, resp.StatusCode, "Ожидался статус 200 OK при получении избранного, тело: %s", string(body))

	var ids []string
	gjson.GetBytes(body, "items.#.advertisement.id").ForEach(func(_, val gjson.Result) bool {
		ids = append(ids, val.String())
		return true
	})
	return ids, string(body)
}

// assertFavoriteError проверяет ожидаемый статус-код и непустые поля error и message в теле ошибки
func assertFavoriteError(t *testing.T, resp *http.Response, expectedStatus int, contextMsg string) {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Ошибка чтения тела ответа %s", contextMsg)
	assert.Equal(t, expectedStatus, resp.StatusCode, "Ожидался статус %d %s, получено тело: %s", expectedStatus, contextMsg, string(body))
	assert.NotEmpty(t, gjson.GetBytes(body, "error").String(), "Поле 'error' не должно быть пустым %s, тело: %s", contextMsg, string(body))
	assert.NotEmpty(t, gjson.GetBytes(body, "message").String(), "Поле 'message' не должно быть пустым %s, тело: %s", contextMsg, string(body))
}
