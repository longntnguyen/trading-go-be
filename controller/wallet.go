package controller

import (
	"context"
	"encoding/json"
	"log"
	"my-app/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetOverView() gin.HandlerFunc {
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
	
		emp.UserID = uuid.New().String()
		createdUser, err := userCollection.InsertOne(context.Background(), &emp) 
	
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			log.Println("Error creating user: ", err)
			res.Error = err.Error() 
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	
		res.Data = createdUser
		c.Writer.WriteHeader(http.StatusOK) 
	}
}