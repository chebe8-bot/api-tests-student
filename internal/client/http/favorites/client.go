package favorites

import (
	"fmt"
	"net/http"
	"testing"
	"testing/fstest"

	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"
)

var minimalPNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
	0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
	0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
	0x42, 0x60, 0x82,
}

func buildTestFS(count int) (fstest.MapFS, []string) {
	fsMap := fstest.MapFS{}
	names := make([]string, 0, count)
	for i := 1; i <= count; i++ {
		name := fmt.Sprintf("photo_%d.png", i)
		fsMap[name] = &fstest.MapFile{Data: minimalPNG}
		names = append(names, name)
	}
	return fsMap, names
}

func HttpPostAdvertisement(t *testing.T, token string, fields map[string]string, photoCount int) *http.Response {
	runner := apiRunner.GetRunner()
	if token != "" {
		runner = runner.Auth(token)
	}

	req := runner.Create()
	if photoCount > 0 {
		fsMap, names := buildTestFS(photoCount)
		req = req.UseFS(fsMap)
		post := req.Post(path.AdvertisementPath)
		for k, v := range fields {
			post = post.MultipartFormData(k, v)
		}
		post = post.MultipartFile("photos", names...)
		return post.Expect(t).End().Response
	}

	post := req.Post(path.AdvertisementPath)
	for k, v := range fields {
		post = post.MultipartFormData(k, v)
	}
	return post.Expect(t).End().Response
}

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
