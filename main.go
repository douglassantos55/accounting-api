package main

import (
	"log"

	"example.com/accounting/api"
)

func main() {
	router := api.GetRouter()
	log.Fatal(router.Run())
}
