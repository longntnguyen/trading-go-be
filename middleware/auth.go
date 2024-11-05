package middleware

import (
	"context"
	"my-app/database"
	"my-app/model"
	"my-app/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.CollectionDB("user")

func Authenticate() gin.HandlerFunc{
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("Authorization")
		if clientToken == "" {
			c.JSON(403, gin.H{"error": "No Token"})
			c.Abort()
			return
		}
		userId, errToken := services.GetUserIdFromToken(clientToken)
		if errToken != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": errToken.Error()})
			c.Abort()
			return
		} 
		var user model.User
		err := userCollection.FindOne(context.Background(), bson.D{{Key: "user_id", Value: userId}}).Decode(&user)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error decoding user"})
			c.Abort()
			return
		}
		c.Next()
	}
}