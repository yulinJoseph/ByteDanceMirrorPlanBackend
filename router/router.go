package router

import (
	"demo-concurrency/service"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	g := gin.Default()
	g.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	g.GET("/getTicketNum", service.GetTicketNum)
	g.GET("/sellTicket", service.SellTicket)
	go service.UserBuyHandler()
	go service.SellTicketHandler()

	return g
}
