package main

import (
	"log"
	"my-app/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    if err != nil {
        log.Fatal(err)
    } 
}

func main(){ 
    router := gin.Default()
    router.Use(cors.Default())
    authRoutes := router.Group("/api/auth/")
    routes.UserAuthRoutes(authRoutes) 
    routes.UsePublicRoutes(router)
    router.Run() 
}