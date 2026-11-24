package favorites

import (
	"io"
	"net/http"
	"os"
	"testing"

	favClient "api-tests-template/internal/client/http/favorites"
	"api-tests-template/internal/managers/auth"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type FavoritesNegativeSuite struct {
	suite.Suite
	buyerToken  string
	sellerToken string
}

func TestFavoritesNegativeSuite(t *testing.T) {
	suite.Run(t, new(FavoritesNegativeSuite))
}

func (s *FavoritesNegativeSuite) SetupSuite() {
	_ = godotenv.Load("../../../.env")
	_ = godotenv.Overload("../../../.env.override")
	s.buyerToken = auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD")).Token
	s.sellerToken = auth.Login(s.T(), os.Getenv("TEST_LOGIN_2"), os.Getenv("TEST_PASSWORD_2")).Token
}

func (s *FavoritesNegativeSuite) TestFavoritesNegativeCases() {
	nonExistentID := "11111111-1111-4111-8111-111111111111"

	s.Run("Несуществующий ID при добавлении и удалении -> 404", func() {
		addMissingResp := favClient.HttpPostFavorite(s.T(), s.buyerToken, nonExistentID)
		defer addMissingResp.Body.Close()
		assertFavoriteError(s.T(), addMissingResp, http.StatusNotFound, "при добавлении несуществующего объявления")

		delMissingResp := favClient.HttpDeleteFavorite(s.T(), s.buyerToken, nonExistentID)
		defer delMissingResp.Body.Close()
		assertFavoriteError(s.T(), delMissingResp, http.StatusNotFound, "при удалении несуществующего объявления")
	})

	s.Run("Запросы к избранному без авторизации -> 401", func() {
		unauthGet := favClient.HttpGetFavorites(s.T(), "")
		defer unauthGet.Body.Close()
		assertFavoriteError(s.T(), unauthGet, http.StatusUnauthorized, "при GET /favorites без токена")

		unauthPost := favClient.HttpPostFavorite(s.T(), "", nonExistentID)
		defer unauthPost.Body.Close()
		assertFavoriteError(s.T(), unauthPost, http.StatusUnauthorized, "при POST /favorites без токена")

		unauthDel := favClient.HttpDeleteFavorite(s.T(), "", nonExistentID)
		defer unauthDel.Body.Close()
		assertFavoriteError(s.T(), unauthDel, http.StatusUnauthorized, "при DELETE /favorites без токена")
	})

	s.Run("Попытка добавить собственное объявление в избранное -> 403", func() {
		ownAdID := createTestAd(s.T(), s.buyerToken, "OwnAd403")
		defer deleteTestAd(s.T(), s.buyerToken, ownAdID)

		addOwnResp := favClient.HttpPostFavorite(s.T(), s.buyerToken, ownAdID)
		defer addOwnResp.Body.Close()
		assertFavoriteError(s.T(), addOwnResp, http.StatusForbidden, "при попытке добавить своё объявление в избранное")
	})

	s.Run("Повторное добавление объявления в избранное -> 409", func() {
		foreignAdID := createTestAd(s.T(), s.sellerToken, "ForeignAd409")
		defer deleteTestAd(s.T(), s.sellerToken, foreignAdID)

		firstAdd := favClient.HttpPostFavorite(s.T(), s.buyerToken, foreignAdID)
		firstBody, _ := io.ReadAll(firstAdd.Body)
		firstAdd.Body.Close()
		require.Equal(s.T(), http.StatusCreated, firstAdd.StatusCode, "Ожидался статус 201 Created при первом добавлении, тело: %s", string(firstBody))
		defer func() {
			r := favClient.HttpDeleteFavorite(s.T(), s.buyerToken, foreignAdID)
			r.Body.Close()
		}()

		duplicateAdd := favClient.HttpPostFavorite(s.T(), s.buyerToken, foreignAdID)
		defer duplicateAdd.Body.Close()
		assertFavoriteError(s.T(), duplicateAdd, http.StatusConflict, "при повторном добавлении в избранное")
	})
}
