package controller

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"my-app/constants"
	"my-app/database"
	"my-app/model"
	"my-app/services"
	"net/http"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.CollectionDB("user")

type UserService struct {
	MongoCollection *mongo.Collection
}

type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		res := &Response{}
		defer json.NewEncoder(c.Writer).Encode(res)

		var emp model.User
		err := json.NewDecoder(c.Request.Body).Decode(&emp)
		if err != nil {
			c.Writer.WriteHeader(http.StatusBadRequest)
			log.Println("Error decoding request body: ", err)
			res.Error = err.Error()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		if emp.Email == "" || emp.Password == "" {
			c.Writer.WriteHeader(http.StatusBadRequest)
			log.Println("Missing required fields")
			res.Error = "Missing required fields"
			return

		}
		var user model.User
		err = userCollection.FindOne(context.Background(), bson.D{{Key: "email", Value: emp.Email}}).Decode(&user)
		if err == nil {
			c.Writer.WriteHeader(http.StatusBadRequest)
			log.Println("Email already exists")
			res.Error = "Email already exists"
			return
		}

		hashedPassword, err := services.HashPassword(emp.Password)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			log.Println("Error hashing password: ", err)
			res.Error = err.Error()
			return
		}

		walletAddress, privateKey, err := services.CreateAccount(hashedPassword)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			log.Println("Error creating account: ", err)
			res.Error = err.Error()
			return
		}

		emp.UserID = uuid.New().String()
		emp.Password = hashedPassword
		emp.WalletAddress = walletAddress
		emp.PrivateKey = privateKey
		_, err = userCollection.InsertOne(context.Background(), &emp)
		var newUser model.User
		fmt.Println(emp, "emp")
		newUserErr := userCollection.FindOne(context.Background(), bson.D{{Key: "user_id", Value: emp.UserID}}).Decode(&newUser)
		if newUserErr != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			log.Println("Error finding user: ", newUserErr)
			res.Error = newUserErr.Error()
			c.JSON(http.StatusInternalServerError, gin.H{"error": newUserErr.Error()})

		}
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			log.Println("Error creating user: ", err)
			res.Error = err.Error()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		token, err := services.CreateToken(newUser.UserID)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			log.Println("Error creating token: ", err)
			res.Error = err.Error()
			return
		}

		res.Data = model.LoginResponse{
			Token: token,
			User: model.UserLoginResponse{
				Email:         newUser.Email,
				Name:          newUser.Name,
				UserID:        newUser.UserID,
				WalletAddress: newUser.WalletAddress,
			},
		}
		c.Writer.WriteHeader(http.StatusOK)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		res := &Response{}
		defer json.NewEncoder(c.Writer).Encode(res)

		var emp model.User
		err := json.NewDecoder(c.Request.Body).Decode(&emp)
		if err != nil {
			c.Writer.WriteHeader(http.StatusBadRequest)
			log.Println("Error decoding request body: ", err)
			res.Error = err.Error()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		var user model.User
		err = userCollection.FindOne(context.Background(), bson.D{{Key: "email", Value: emp.Email}}).Decode(&user)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			log.Println("Error finding user: ", err)
			res.Error = err.Error()
			return
		}
		validPassword := services.ComparePasswords(emp.Password, user.Password)

		if !validPassword {
			c.Writer.WriteHeader(http.StatusUnauthorized)
			log.Println("Incorrect password")
			res.Error = "Incorrect password"
			return
		}

		token, err := services.CreateToken(user.UserID)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			log.Println("Error creating token: ", err)
			res.Error = err.Error()
			return
		}

		res.Data = model.LoginResponse{
			Token: token,
			User: model.UserLoginResponse{
				Email:         user.Email,
				Name:          user.Name,
				UserID:        user.UserID,
				WalletAddress: user.WalletAddress,
			},
		}
		c.Writer.WriteHeader(http.StatusOK)
	}
}

func GetUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		res := &Response{}
		defer json.NewEncoder(c.Writer).Encode(res)

		token := c.Request.Header.Get("Authorization")
		userId, errToken := services.GetUserIdFromToken(token)
		if errToken != nil {
			c.Writer.WriteHeader(http.StatusUnauthorized)
			log.Println("Error decoding user: ", errToken)
			res.Error = errToken.Error()
			c.JSON(http.StatusUnauthorized, gin.H{"error": errToken.Error()})
		}
		var user model.User
		err := userCollection.FindOne(context.Background(), bson.D{{Key: "user_id", Value: userId}}).Decode(&user)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			log.Println("Error decoding user: ", err)
			res.Error = err.Error()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		userInfoResponse := model.GetUserInfoResponse{
			TokenBalance: []model.TokenBalance{},
			User: model.UserLoginResponse{
				Email:         user.Email,
				Name:          user.Name,
				UserID:        user.UserID,
				WalletAddress: user.WalletAddress,
			},
		}

		var symbols []string
		for _, token := range constants.TOKEN_LIST {
			symbols = append(symbols, token.Symbol)
		}
		listTokenInfo, tokenPricesErr := services.GetTokenPrice(symbols)
		if tokenPricesErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": tokenPricesErr.Error()})
			log.Fatal("Error getting token prices: ", tokenPricesErr)

		}

		for _, token := range constants.TOKEN_LIST {
			tokenBalance, errorBalance := services.TokenBalance(token.Address, user.WalletAddress)
			if errorBalance != nil {
				log.Fatal("Error getting token balance: ", errorBalance)
				c.JSON(http.StatusInternalServerError, gin.H{"error": errorBalance.Error()})
			}
			balanceInUSD := new(big.Float)
			for _, tokenInfo := range listTokenInfo {
				if token.Symbol == tokenInfo.Symbol {
					balanceFloat64, _ := tokenInfo.Balance.Float64()
					balanceInUSD.SetFloat64(balanceFloat64)
				}
			}
			balanceInUSDFloat, _ := balanceInUSD.Float64()
			balanceInUSD.SetFloat64(math.Round(balanceInUSDFloat*1e6) / 1e6)
			userInfoResponse.TokenBalance = append(userInfoResponse.TokenBalance, model.TokenBalance{
				TokenName:    token.Name,
				Balance:      *tokenBalance,
				BalanceInUSD: *balanceInUSD,
			})
		}

		res.Data = userInfoResponse
		c.Writer.WriteHeader(http.StatusOK)
	}
}

func GetUserBalanceByToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		res := &Response{}
		defer json.NewEncoder(c.Writer).Encode(res)

		token := c.Request.Header.Get("Authorization")
		tokenAddress := c.Query("tokenAddress")
		userId, errToken := services.GetUserIdFromToken(token)
		if errToken != nil {
			c.Writer.WriteHeader(http.StatusUnauthorized)
			log.Println("Error decoding user: ", errToken)
			res.Error = errToken.Error()
			c.JSON(http.StatusUnauthorized, gin.H{"error": errToken.Error()})
		}
		var user model.User
		err := userCollection.FindOne(context.Background(), bson.D{{Key: "user_id", Value: userId}}).Decode(&user)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			log.Println("Error decoding user: ", err)
			res.Error = err.Error()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		tokenInformation, err := services.GetTokenByAddress(tokenAddress)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return

		}
		if tokenInformation.Symbol == "BNB" {
			balance, err := services.GetBNBBalance(user.WalletAddress)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return

			}
			balanceRes, _ := balance.Float64()
			res.Data = model.GetUserBalanceByTokenResponse{
				Balance:   balanceRes,
				Symbol:    tokenInformation.Symbol,
				TokenName: tokenInformation.TokenName,
			}
			return
		}
		balance, err := services.TokenBalance(tokenAddress, user.WalletAddress)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return

		}
		balanceRes, _ := balance.Float64()

		res.Data = model.GetUserBalanceByTokenResponse{
			Balance:   balanceRes,
			Symbol:    tokenInformation.Symbol,
			TokenName: tokenInformation.TokenName,
		}

	}
}
