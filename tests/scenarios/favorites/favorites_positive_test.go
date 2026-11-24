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

type FavoritesPositiveSuite struct {
	suite.Suite
	buyerToken string
	myUserID   string
}

func TestFavoritesPositiveSuite(t *testing.T) {
	suite.Run(t, new(FavoritesPositiveSuite))
}

func (s *FavoritesPositiveSuite) SetupSuite() {
	_ = godotenv.Load("../../../.env")
	_ = godotenv.Overload("../../../.env.override")
	loginResp := auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD"))
	s.buyerToken = loginResp.Token
	s.myUserID = loginResp.User.Id
}

func (s *FavoritesPositiveSuite) TestFavoritesFullLifecycle() {
	adID := s.findForeignAdID()
	require.NotEmpty(s.T(), adID, "На сайте должно быть хотя бы одно чужое объявление")

	// На всякий случай удаляем из избранного перед стартом теста
	cleanResp := favClient.HttpDeleteFavorite(s.T(), s.buyerToken, adID)
	cleanResp.Body.Close()

	// 1. Добавить в избранное -> есть в списке
	addResp := favClient.HttpPostFavorite(s.T(), s.buyerToken, adID)
	defer addResp.Body.Close()
	require.Equal(s.T(), http.StatusCreated, addResp.StatusCode)
	assert.Contains(s.T(), s.getFavoriteAdIDs(), adID)

	// 2. Удалить из избранного -> нет в списке
	delFavResp := favClient.HttpDeleteFavorite(s.T(), s.buyerToken, adID)
	defer delFavResp.Body.Close()
	require.Equal(s.T(), http.StatusNoContent, delFavResp.StatusCode)
	assert.NotContains(s.T(), s.getFavoriteAdIDs(), adID)
}

func (s *FavoritesPositiveSuite) findForeignAdID() string {
	resp := favClient.HttpGetAdvertisements(s.T(), s.buyerToken)
	defer resp.Body.Close()
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var foreignIDs []string
	gjson.GetBytes(body, "items").ForEach(func(_, item gjson.Result) bool {
		if item.Get("user_id").String() != s.myUserID {
			foreignIDs = append(foreignIDs, item.Get("id").String())
		}
		return true
	})
	if len(foreignIDs) == 0 {
		gjson.ParseBytes(body).ForEach(func(_, item gjson.Result) bool {
			if item.Get("user_id").String() != s.myUserID {
				foreignIDs = append(foreignIDs, item.Get("id").String())
			}
			return true
		})
	}
	if len(foreignIDs) > 1 {
		return foreignIDs[1]
	}
	if len(foreignIDs) == 1 {
		return foreignIDs[0]
	}
	return ""
}

func (s *FavoritesPositiveSuite) getFavoriteAdIDs() []string {
	resp := favClient.HttpGetFavorites(s.T(), s.buyerToken)
	defer resp.Body.Close()
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var ids []string
	gjson.GetBytes(body, "items.#.advertisement.id").ForEach(func(_, val gjson.Result) bool {
		ids = append(ids, val.String())
		return true
	})
	return ids
}
