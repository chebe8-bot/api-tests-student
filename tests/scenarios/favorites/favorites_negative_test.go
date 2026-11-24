package favorites

import (
	"io"
	"net/http"
	"os"
	"testing"

	favClient "api-tests-template/internal/client/http/favorites"
	"api-tests-template/internal/managers/auth"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

type FavoritesNegativeSuite struct {
	suite.Suite
	buyerToken string
	myUserID   string
}

func TestFavoritesNegativeSuite(t *testing.T) {
	suite.Run(t, new(FavoritesNegativeSuite))
}

func (s *FavoritesNegativeSuite) SetupSuite() {
	_ = godotenv.Load("../../../.env")
	_ = godotenv.Overload("../../../.env.override")
	loginResp := auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD"))
	s.buyerToken = loginResp.Token
	s.myUserID = loginResp.User.Id
}

func (s *FavoritesNegativeSuite) TestFavoritesNegativeCases() {
	nonExistentID := "11111111-1111-4111-8111-111111111111"

	// 1. Несуществующий ID -> 404
	addMissingResp := favClient.HttpPostFavorite(s.T(), s.buyerToken, nonExistentID)
	defer addMissingResp.Body.Close()
	assert.Equal(s.T(), http.StatusNotFound, addMissingResp.StatusCode)
	s.assertErrorBody(addMissingResp)

	delMissingResp := favClient.HttpDeleteFavorite(s.T(), s.buyerToken, nonExistentID)
	defer delMissingResp.Body.Close()
	assert.Equal(s.T(), http.StatusNotFound, delMissingResp.StatusCode)
	s.assertErrorBody(delMissingResp)

	// 2. Без авторизации -> 401
	unauthGet := favClient.HttpGetFavorites(s.T(), "")
	defer unauthGet.Body.Close()
	assert.Equal(s.T(), http.StatusUnauthorized, unauthGet.StatusCode)
	s.assertErrorBody(unauthGet)

	unauthPost := favClient.HttpPostFavorite(s.T(), "", nonExistentID)
	defer unauthPost.Body.Close()
	assert.Equal(s.T(), http.StatusUnauthorized, unauthPost.StatusCode)
	s.assertErrorBody(unauthPost)

	unauthDel := favClient.HttpDeleteFavorite(s.T(), "", nonExistentID)
	defer unauthDel.Body.Close()
	assert.Equal(s.T(), http.StatusUnauthorized, unauthDel.StatusCode)
	s.assertErrorBody(unauthDel)

	// 3. Своё объявление в избранное -> 403
	ownAdResp := favClient.HttpPostAdvertisement(s.T(), s.buyerToken, map[string]string{
		"title":       "Своё объявление",
		"description": "Проверка 403",
	}, 1)
	defer ownAdResp.Body.Close()
	require.Equal(s.T(), http.StatusCreated, ownAdResp.StatusCode)
	ownBody, _ := io.ReadAll(ownAdResp.Body)
	ownAdID := gjson.GetBytes(ownBody, "id").String()
	defer func() {
		r := favClient.HttpDeleteAdvertisement(s.T(), s.buyerToken, ownAdID)
		r.Body.Close()
	}()

	addOwnResp := favClient.HttpPostFavorite(s.T(), s.buyerToken, ownAdID)
	defer addOwnResp.Body.Close()
	assert.Equal(s.T(), http.StatusForbidden, addOwnResp.StatusCode)
	s.assertErrorBody(addOwnResp)

	// 4. Повторное добавление чужого объявления -> 409
	foreignAdID := s.findForeignAdID()
	require.NotEmpty(s.T(), foreignAdID)

	cleanResp := favClient.HttpDeleteFavorite(s.T(), s.buyerToken, foreignAdID)
	cleanResp.Body.Close()

	firstAdd := favClient.HttpPostFavorite(s.T(), s.buyerToken, foreignAdID)
	defer firstAdd.Body.Close()
	require.Equal(s.T(), http.StatusCreated, firstAdd.StatusCode)
	defer func() {
		r := favClient.HttpDeleteFavorite(s.T(), s.buyerToken, foreignAdID)
		r.Body.Close()
	}()

	duplicateAdd := favClient.HttpPostFavorite(s.T(), s.buyerToken, foreignAdID)
	defer duplicateAdd.Body.Close()
	assert.Equal(s.T(), http.StatusConflict, duplicateAdd.StatusCode)
	s.assertErrorBody(duplicateAdd)
}

func (s *FavoritesNegativeSuite) findForeignAdID() string {
	resp := favClient.HttpGetAdvertisements(s.T(), s.buyerToken)
	defer resp.Body.Close()
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var foreignID string
	gjson.GetBytes(body, "items").ForEach(func(_, item gjson.Result) bool {
		if item.Get("user_id").String() != s.myUserID {
			foreignID = item.Get("id").String()
			return false
		}
		return true
	})
	if foreignID == "" {
		gjson.ParseBytes(body).ForEach(func(_, item gjson.Result) bool {
			if item.Get("user_id").String() != s.myUserID {
				foreignID = item.Get("id").String()
				return false
			}
			return true
		})
	}
	return foreignID
}

func (s *FavoritesNegativeSuite) assertErrorBody(resp *http.Response) {
	body, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), gjson.GetBytes(body, "error").String(), "Поле 'error' в теле ошибки не должно быть пустым: %s", string(body))
	assert.NotEmpty(s.T(), gjson.GetBytes(body, "message").String(), "Поле 'message' в теле ошибки не должно быть пустым: %s", string(body))
}
