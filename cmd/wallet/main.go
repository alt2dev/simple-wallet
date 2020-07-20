package main

import (
	"log"
	"os"

	mw "github.com/alt2dev/simple-wallet/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	var err error

	err = mw.ParseDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	err = mw.InitDB()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer mw.CloseDB()

	router := gin.Default()

	router.POST("/wallet/create", mw.CreatePOST)
	router.POST("/wallet/topup", mw.TopupPOST)
	router.POST("/wallet/send", mw.SendPOST)
	router.GET("/wallet/:id/history", mw.HistoryGET)

	router.Run(":8080")
}
