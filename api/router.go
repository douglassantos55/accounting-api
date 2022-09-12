package api

import (
	"net/http"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func GetRouter() *gin.Engine {
	if router == nil {
		router = gin.Default()

		config := cors.DefaultConfig()
		config.AllowAllOrigins = true
		config.AddAllowHeaders("CompanyID")

		router.Use(cors.New(config))

		router.Use(func(c *gin.Context) {
			companyID, err := strconv.ParseUint(c.Request.Header.Get("CompanyID"), 10, 64)
			if err != nil {
				c.Status(http.StatusBadRequest)
				return
			}
			c.Set("CompanyID", uint(companyID))
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
	RegisterServicesEndpoints(router)
}
