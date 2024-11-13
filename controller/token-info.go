package controller

import (
	"encoding/json"
	"log"
	"my-app/model"
	"my-app/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetTokenInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		res := &Response{}
		defer json.NewEncoder(c.Writer).Encode(res)

		page := c.Query("page")
		limit := c.Query("limit")
		pageInt, err := strconv.Atoi(func() string {
			if page == "" {
				return "1"
			}
			return page
		}())
		if err != nil {
			log.Println(c.Writer, "Error parsing paging: %v\n", err)
		}
		limitInt, err := strconv.Atoi(func() string {
			if limit == "" {
				return "10"
			}
			return limit
		}())
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
