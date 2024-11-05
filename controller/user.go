package controller

import (
	"context"
	"log"
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
	Data	interface{}	`json:"data,omitempty"`
	Error	string		`json:"error,omitempty"`
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
			log.Println("Invalid password")
			res.Error = "Invalid password" 
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
				Email: user.Email,
				Name: user.Name,
				UserID: user.UserID,
				WalletAddress: user.WalletAddress,
			},
		}
		c.Writer.WriteHeader(http.StatusOK)
	}
}

func GetUserInfo() gin.HandlerFunc{
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
				Email: user.Email,
				Name: user.Name,
				UserID: user.UserID,
				WalletAddress: user.WalletAddress,
			},
		}

		for _, token := range constants.TOKEN_LIST {
			tokenBalance, errorBalance := services.TokenBalance(token.Address, user.WalletAddress)
			if errorBalance != nil {
				log.Fatal("Error getting token balance: ", errorBalance)
			}
			userInfoResponse.TokenBalance = append(userInfoResponse.TokenBalance, model.TokenBalance{
				Name: token.Name,
				Balance: tokenBalance.String(),
			})
		}
	
		res.Data = userInfoResponse
		c.Writer.WriteHeader(http.StatusOK) 
	}
}