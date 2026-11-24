package favorites

import (
	"fmt"
	"net/http"
	"testing"

	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"
)

// HttpPostFavorite добавляет объявление в избранное
func HttpPostFavorite(t *testing.T, token, adID string) *http.Response {
	runner := apiRunner.GetRunner()
	if token != "" {
		runner = runner.Auth(token)
	}
	return runner.Create().
		Post(path.FavoritesPath).
		JSON(map[string]string{"advertisement_id": adID}).
		Expect(t).
		End().Response
}

// HttpGetFavorites возвращает список избранных объявлений
func HttpGetFavorites(t *testing.T, token string) *http.Response {
	runner := apiRunner.GetRunner()
	if token != "" {
		runner = runner.Auth(token)
	}
	return runner.Create().
		Get(path.FavoritesPath).
		Expect(t).
		End().Response
}

// HttpDeleteFavorite удаляет объявление из избранного
func HttpDeleteFavorite(t *testing.T, token, adID string) *http.Response {
	runner := apiRunner.GetRunner()
	if token != "" {
		runner = runner.Auth(token)
	}
	return runner.Create().
		Delete(fmt.Sprintf("%s/%s", path.FavoritesPath, adID)).
		Expect(t).
		End().Response
}
