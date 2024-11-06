package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"my-app/constants"
	"my-app/model"
	"my-app/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetOverView() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		res := &Response{}
		defer json.NewEncoder(c.Writer).Encode(res)

		userId, err := services.GetUserIdFromToken(c.Request.Header.Get("Authorization"))

		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			log.Println("Error decoding user: ", err)
			res.Error = err.Error()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		var user model.User
		err = userCollection.FindOne(context.Background(), bson.D{{Key: "user_id", Value: userId}}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var symbols []string
		for _, token := range constants.TOKEN_LIST {
			symbols = append(symbols, token.Symbol)
		}
		listTokenInfo, tokenPricesErr := services.GetTokenPrice(symbols)

		if tokenPricesErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": tokenPricesErr.Error()})
			return
		}

		overViewResponse := model.OverViewResponse{
			TotalBalance:       *big.NewFloat(0),
			TotalBalanceInUSD:  *big.NewFloat(0),
			TokensOverViewList: []model.TokenOverView{},
		}
		for _, token := range constants.TOKEN_LIST {
			tokenBalance, errorBalance := services.TokenBalance(token.Address, user.WalletAddress)
			if errorBalance != nil {
				log.Fatal("Error getting token balance: ", errorBalance)
				c.JSON(http.StatusInternalServerError, gin.H{"error": errorBalance.Error()})
			}
			balanceInUSD := new(big.Float)
			percentChange24H := new(big.Float)
			volume24H := new(big.Float)
			marketCap := new(big.Float)
			tokenId := new(json.Number)
			for _, tokenInfo := range listTokenInfo {
				if token.Symbol == tokenInfo.Symbol {
					balanceFloat64, _ := tokenInfo.Balance.Float64()
					balanceInUSD.SetFloat64(balanceFloat64)
					percentChange24HFloat64, _ := tokenInfo.PercentChange24h.Float64()
					percentChange24H.SetFloat64(percentChange24HFloat64)
					volume24HFloat64, _ := tokenInfo.Volume24H.Float64()
					volume24H.SetFloat64(volume24HFloat64)
					marketCapFloat64, _ := tokenInfo.MarketCap.Float64()
					marketCap.SetFloat64(marketCapFloat64)
					tokenId = &tokenInfo.TokenID
				}
			}

			overViewResponse.TokensOverViewList = append(overViewResponse.TokensOverViewList, model.TokenOverView{
				TokenName:        token.Name,
				Balance:          *tokenBalance,
				TokenID:          *tokenId,
				BalanceInUSD:     *balanceInUSD,
				PercentChange24h: *percentChange24H,
				Volume24H:        *volume24H,
				MarketCap:        *marketCap,
				Symbol:           token.Symbol,
				ImageUrl:         fmt.Sprintf("https://s2.coinmarketcap.com/static/img/coins/64x64/%s.png", tokenId),
			})
		}
		res.Data = overViewResponse
		c.Writer.WriteHeader(http.StatusOK)
	}

}
