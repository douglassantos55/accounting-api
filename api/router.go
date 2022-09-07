package api

import (
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func GetRouter() *gin.Engine {
	if router == nil {
		router = gin.Default()

		router.Use(func(c *gin.Context) {
			c.Set("CompanyID", uint(1))
		})

		registerRoutes(router)
	}

	return router
}

func registerRoutes(router *gin.Engine) {
	RegisterAccountsEndpoints(router)
	RegisterCustomerEndpoints(router)
	RegisterVendorEndpoints(router)
	RegisterProductEndpoints(router)
	RegisterPurchaseEndpoints(router)
	RegisterEntriesEndpoint(router)
	RegisterSalesEndpoints(router)
}
