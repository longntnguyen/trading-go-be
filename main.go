package main

import (
	"context"
	"fmt"
	"log"
	"my-app/controller"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type todo struct {
    ID      string  `json:"id"`
    Name    string  `json:"name"`
    Active  bool    `json:"active"`
}

var mongoClient *mongo.Client

var todoList = []todo{
    {ID: "1", Name: "Task 1", Active: true},
    {ID: "2", Name: "Task 2", Active: false},
}

func init() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    mongoClient,err= mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGO_URI")))

    if err != nil {
        log.Fatal(err)
    }

    err = mongoClient.Ping(context.Background(),  readpref.Primary())

    if err != nil {
        log.Fatal("Ping fail: ", err)
    }
}

func main(){ 
    defer mongoClient.Disconnect(context.Background())
    coll := mongoClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("COLLECTION_NAME"))

    userService := controller.UserService{MongoCollection: coll}
    router := gin.Default()
    router.GET("/todo", func(c *gin.Context){
        c.IndentedJSON(http.StatusOK, todoList)
    })
    router.GET(("users"), func(c *gin.Context){
        fmt.Println("ccccc",c.Request.Body)
        userService.GetAllUsers(c.Writer, c.Request)
    })
    router.GET(("user"), func(c *gin.Context){ 
        userService.GetUserByEmail(c.Writer, c.Request)
    })
    router.GET(("login"), func(c *gin.Context){
        fmt.Println("AAAAAAAAAA")
    })
    router.POST(("login"), func(c *gin.Context){
        userService.Login(c.Writer, c.Request)
    })
    router.POST(("register"), func(c *gin.Context){
        userService.CreateUser(c.Writer, c.Request)
    })
    router.Run() 
}