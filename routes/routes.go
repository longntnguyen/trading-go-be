package routes

import (
	"my-app/controller"
	"my-app/middleware"

	"github.com/gin-gonic/gin"
)

func UserAuthRoutes(router *gin.RouterGroup) {
    router.Use(middleware.Authenticate())
	router.GET("/user", controller.GetUserInfo())
}

func UsePublicRoutes(router *gin.Engine) { 
	router.POST("/register", controller.SignUp())
	router.POST("/login", controller.Login())
}