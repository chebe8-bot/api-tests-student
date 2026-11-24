package favorites

import (
	"fmt"
	"net/http"
	"testing"

	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"
)

// HttpPostAdvertisement отправляет запрос на создание объявления с указанными файлами фотографий
func HttpPostAdvertisement(t *testing.T, token string, fields map[string]string, photoPaths ...string) *http.Response {
	runner := apiRunner.GetRunner()
	if token != "" {
		runner = runner.Auth(token)
	}

	post := runner.Create().Post(path.AdvertisementPath)
	for k, v := range fields {
		post = post.MultipartFormData(k, v)
	}
	if len(photoPaths) > 0 {
		post = post.MultipartFile("photos", photoPaths...)
	}
	return post.Expect(t).End().Response
}

// HttpGetAdvertisementByID получает объявление по его ID
func HttpGetAdvertisementByID(t *testing.T, token, adID string) *http.Response {
	runner := apiRunner.GetRunner()
	if token != "" {
		runner = runner.Auth(token)
	}
	return runner.Create().
		Get(path.AdvertisementPath).
		Query("id", adID).
		Expect(t).
		End().Response
}

// HttpSearchAdvertisements выполняет поиск объявлений по подстроке search
func HttpSearchAdvertisements(t *testing.T, token, search string) *http.Response {
	runner := apiRunner.GetRunner()
	if token != "" {
		runner = runner.Auth(token)
	}
	return runner.Create().
		Get(path.AdvertisementsPath).
		Query("search", search).
		Expect(t).
		End().Response
}

// HttpGetAdvertisements возвращает общий список объявлений
func HttpGetAdvertisements(t *testing.T, token string) *http.Response {
	runner := apiRunner.GetRunner()
	if token != "" {
		runner = runner.Auth(token)
	}
	return runner.Create().
		Get(path.AdvertisementsPath).
		Expect(t).
		End().Response
}

// HttpGetAdvertisementPhotos получает список фотографий объявления
func HttpGetAdvertisementPhotos(t *testing.T, token, adID string) *http.Response {
	runner := apiRunner.GetRunner()
	if token != "" {
		runner = runner.Auth(token)
	}
	return runner.Create().
		Get(fmt.Sprintf("%s/%s/photos", path.AdvertisementsPath, adID)).
		Expect(t).
		End().Response
}

// HttpDeleteAdvertisement удаляет объявление по его ID
func HttpDeleteAdvertisement(t *testing.T, token, adID string) *http.Response {
	runner := apiRunner.GetRunner()
	if token != "" {
		runner = runner.Auth(token)
	}
	return runner.Create().
		Delete(path.AdvertisementPath).
		Query("id", adID).
		Expect(t).
		End().Response
}

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
