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
		// Check login
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
		// get Token information
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
			TotalBalance:       0.0,
			TotalBalanceInUSD:  0.0,
			TokensOverViewList: []model.TokenOverView{},
		}
		totalBalanceInUSD := new(big.Float)
		balanceInUSCBTC := new(big.Float)
		for _, token := range constants.TOKEN_LIST {
			tokenBalance := new(big.Float)
			if token.Symbol == "BNB" {
				tokenBalance, err = services.GetBNBBalance(user.WalletAddress)
				if err != nil {
					log.Fatal("Error getting BNB balance: ", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				}
			} else {
				tokenBalance, err = services.TokenBalance(token.Address, user.WalletAddress)
				if err != nil {
					log.Fatal("Error getting token balance: ", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				}
			}

			balanceInUSD := new(big.Float)
			percentChange24H := new(big.Float)
			volume24H := new(big.Float)
			marketCap := new(big.Float)
			tokenId := new(big.Float)
			for _, tokenInfo := range listTokenInfo {
				if tokenInfo.Symbol == "BTC" {
					balanceFloat64, _ := tokenInfo.Balance.Float64()
					balanceInUSCBTC.SetFloat64(balanceFloat64)
				}
				if token.Symbol == tokenInfo.Symbol {
					balanceFloat64, _ := tokenInfo.Balance.Float64()
					balanceInUSD.SetFloat64(balanceFloat64)
					percentChange24HFloat64, _ := tokenInfo.PercentChange24h.Float64()
					percentChange24H.SetFloat64(percentChange24HFloat64)
					volume24HFloat64, _ := tokenInfo.Volume24H.Float64()
					volume24H.SetFloat64(volume24HFloat64)
					marketCapFloat64, _ := tokenInfo.MarketCap.Float64()
					marketCap.SetFloat64(marketCapFloat64)
					tokenIdFloat, _ := tokenInfo.TokenID.Float64()
					tokenId = new(big.Float).SetFloat64(tokenIdFloat)
				}
			}
			tokenBalanceInUSD := new(big.Float).Mul(tokenBalance, balanceInUSD)
			totalBalanceInUSD.Add(totalBalanceInUSD, tokenBalanceInUSD)
			overViewResponse.TokensOverViewList = append(overViewResponse.TokensOverViewList, model.TokenOverView{
				TokenName:        token.Name,
				Balance:          *tokenBalance,
				TokenID:          json.Number(tokenId.Text('f', -1)),
				BalanceInUSD:     *balanceInUSD,
				PercentChange24h: *percentChange24H,
				Volume24H:        *volume24H,
				MarketCap:        *marketCap,
				Symbol:           token.Symbol,
				ImageUrl:         fmt.Sprintf("https://s2.coinmarketcap.com/static/img/coins/64x64/%s.png", tokenId.Text('f', -1)),
			})
		}
		totalBalanceInUSDFloat64, _ := totalBalanceInUSD.Float64()
		overViewResponse.TotalBalanceInUSD = totalBalanceInUSDFloat64
		if balanceInUSCBTC.Cmp(big.NewFloat(0)) != 0 && totalBalanceInUSDFloat64 != 0 {
			balanceInUSCBTCFloat64, _ := balanceInUSCBTC.Float64()
			overViewResponse.TotalBalance = totalBalanceInUSDFloat64 / balanceInUSCBTCFloat64
		}
		res.Data = overViewResponse
		c.Writer.WriteHeader(http.StatusOK)
	}

}
