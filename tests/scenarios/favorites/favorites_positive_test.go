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
)

type FavoritesPositiveSuite struct {
	suite.Suite
	buyerToken  string
	sellerToken string
}

func TestFavoritesPositiveSuite(t *testing.T) {
	suite.Run(t, new(FavoritesPositiveSuite))
}

func (s *FavoritesPositiveSuite) SetupSuite() {
	_ = godotenv.Load("../../../.env")
	_ = godotenv.Overload("../../../.env.override")
	s.buyerToken = auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD")).Token
	s.sellerToken = auth.Login(s.T(), os.Getenv("TEST_LOGIN_2"), os.Getenv("TEST_PASSWORD_2")).Token
}

func (s *FavoritesPositiveSuite) TestFavoritesFullLifecycle() {
	adID := createTestAd(s.T(), s.sellerToken, "FavPosAd")
	defer func() {
		r := favClient.HttpDeleteFavorite(s.T(), s.buyerToken, adID)
		r.Body.Close()
	}()

	// 1. Добавить в избранное -> есть в GET /favorites
	addResp := favClient.HttpPostFavorite(s.T(), s.buyerToken, adID)
	addBody, _ := io.ReadAll(addResp.Body)
	addResp.Body.Close()
	require.Equal(s.T(), http.StatusCreated, addResp.StatusCode, "Ожидался статус 201 Created при добавлении в избранное, тело: %s", string(addBody))

	idsAfterAdd, favBody1 := getFavoriteAdIDs(s.T(), s.buyerToken)
	assert.Contains(s.T(), idsAfterAdd, adID, "Добавленное объявление должно присутствовать в GET /favorites, тело: %s", favBody1)

	// 2. Удалить из избранного -> нет в GET /favorites
	delFavResp := favClient.HttpDeleteFavorite(s.T(), s.buyerToken, adID)
	delFavBody, _ := io.ReadAll(delFavResp.Body)
	delFavResp.Body.Close()
	require.Equal(s.T(), http.StatusNoContent, delFavResp.StatusCode, "Ожидался статус 204 No Content при удалении из избранного, тело: %s", string(delFavBody))

	idsAfterDel, favBody2 := getFavoriteAdIDs(s.T(), s.buyerToken)
	assert.NotContains(s.T(), idsAfterDel, adID, "Удалённое объявление должно отсутствовать в GET /favorites, тело: %s", favBody2)

	// 3. Каскадный сценарий: снова добавляем -> продавец удаляет объявление -> оно исчезает из избранного
	reAddResp := favClient.HttpPostFavorite(s.T(), s.buyerToken, adID)
	reAddBody, _ := io.ReadAll(reAddResp.Body)
	reAddResp.Body.Close()
	require.Equal(s.T(), http.StatusCreated, reAddResp.StatusCode, "Ожидался статус 201 Created при повторном добавлении в избранное, тело: %s", string(reAddBody))

	deleteTestAd(s.T(), s.sellerToken, adID)

	idsAfterAdDelete, favBody3 := getFavoriteAdIDs(s.T(), s.buyerToken)
	assert.NotContains(s.T(), idsAfterAdDelete, adID, "После удаления объявления владельцем оно должно пропасть из избранного, тело: %s", favBody3)
}
