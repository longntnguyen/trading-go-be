package controller

import (
	"encoding/json"
	"math/big"

	"github.com/gin-gonic/gin"

	"context"
	"log"
	"my-app/model"
	"my-app/services"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

func TransferToAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		res := &Response{}
		defer json.NewEncoder(c.Writer).Encode(res)

		var transfer model.TransferToAddressRequest
		err := json.NewDecoder(c.Request.Body).Decode(&transfer)
		if err != nil {
			c.Writer.WriteHeader(http.StatusBadRequest)
			log.Println("Error decoding request body: ", err)
			res.Error = err.Error()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

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

		amount, err := transfer.Amount.Float64()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		bigAmount := big.NewFloat(amount)
		_, transferErr := services.SendToken(user.WalletAddress, transfer.ToAddress, user.PrivateKey, transfer.TokenAddress, bigAmount)
		if transferErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": transferErr.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Transfer success"})
	}
}

func GetTransferFee() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		res := &Response{}
		defer json.NewEncoder(c.Writer).Encode(res)

		var transfer model.TransferToAddressRequest
		err := json.NewDecoder(c.Request.Body).Decode(&transfer)
		if err != nil {
			c.Writer.WriteHeader(http.StatusBadRequest)
			log.Println("Error decoding request body: ", err)
			res.Error = err.Error()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

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

		amount, err := transfer.Amount.Float64()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		bigAmount := big.NewFloat(amount)
		transferFee, transferFeeErr := services.GetTransferFee(user.WalletAddress, transfer.ToAddress, user.PrivateKey, transfer.TokenAddress, bigAmount)
		if transferFeeErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": transferFeeErr.Error()})
			return
		}
		res.Data = transferFee
		c.Writer.WriteHeader(http.StatusOK)
	}
}