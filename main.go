package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.String(200, "Hello World!")
	})

	router.POST("/accounts", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"id":          1,
			"name":        c.PostForm("name"),
			"placeholder": true,
		})
	})

	return router
}

func main() {
	router := NewRouter()
	log.Fatal(router.Run())
}
