package controller

import (
	"context"
	"encoding/json"
	"log"
	"my-app/model"
	"my-app/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetTokenInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		res := &Response{}
		defer json.NewEncoder(c.Writer).Encode(res)
		userId, err := services.GetUserIdFromToken(c.Request.Header.Get("Authorization"))

		page := c.Query("page")
		limit := c.Query("limit")
		pageInt, err := strconv.Atoi(page)
		if err != nil {
			log.Println(c.Writer, "Error parsing paging: %v\n", err)
		}
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			log.Println(c.Writer, "Error parsing limit: %v\n", err)
		}
		log.Println(c.Writer, "start: %d, limit: %d\n", pageInt, limitInt)
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
		tokens, err := services.GetListTokenInfo(pageInt, limitInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch token data"})
			return
		}

		listTokenInfoResponse := model.ListTokenInfoResponse{
			Page:   pageInt,
			Limit:  limitInt,
			Result: tokens,
		}
		res.Data = listTokenInfoResponse
		c.Writer.WriteHeader(http.StatusOK)
	}

}
