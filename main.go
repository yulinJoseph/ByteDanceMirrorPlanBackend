package main

import (
	"demo-concurrency/router"
	"demo-concurrency/service"
	"demo-concurrency/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
)

func main() {
	utils.InitMysql()
	utils.InitRedis()
	utils.InitRabbitmq()

	// temporary solution
	service.UploadTickets()
	fmt.Println("upload tickets")

	file, _ := os.Create("gin.log")
	gin.DefaultWriter = file
	if err := router.Router().Run(":7898"); err != nil {
		fmt.Println("Failed to start server")
		return
	}
}
