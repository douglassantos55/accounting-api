package api

import "github.com/gin-gonic/gin"

var router *gin.Engine

func GetRouter() *gin.Engine {
	if router == nil {
		router = gin.Default()
		registerRoutes()
	}

	router.Use(func(c *gin.Context) {
		c.Set("CompanyID", uint(1))
	})

	return router
}

func registerRoutes() {
	RegisterAccountsEndpoints()
}
