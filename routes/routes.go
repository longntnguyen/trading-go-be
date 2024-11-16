package routes

import (
	"my-app/controller"
	"my-app/middleware"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func UserAuthRoutes(router *gin.RouterGroup) {
	router.Use(middleware.Authenticate())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{http.MethodGet, http.MethodPatch, http.MethodPost, http.MethodHead, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Content-Type", "X-XSRF-TOKEN", "Accept", "Origin", "X-Requested-With", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	router.GET("/user", controller.GetUserInfo())
	router.GET("/overview", controller.GetOverView())
	router.POST("/transfer", controller.TransferToAddress())
	router.POST("/transfer-fee", controller.GetTransferFee())
	router.GET("/token-info", controller.GetUserBalanceByToken())
}

func UsePublicRoutes(router *gin.Engine) {
	router.POST("/register", controller.SignUp())
	router.POST("/login", controller.Login())
	router.GET("/tokens-list", controller.GetTokenInfo())
}
