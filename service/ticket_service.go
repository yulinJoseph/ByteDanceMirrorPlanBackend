package service

import (
	"bytes"
	"context"
	"demo-concurrency/model"
	"demo-concurrency/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"sync"
)

var rwmutex = sync.RWMutex{}

func UploadTickets() {
	var tickets *[]model.Ticket
	utils.DB.Where("status = ?", model.Selling).Find(&tickets)
	rwmutex.Lock()
	for _, ticket := range *tickets {
		utils.Redis.SAdd(context.Background(), "tickets", ticket.ID)
	}
	rwmutex.Unlock()
}

func GetTicketNum(c *gin.Context) {
	rwmutex.RLock()
	ticketNum := utils.Redis.SCard(context.Background(), "tickets")
	rwmutex.RUnlock()
	if ticketNum.Err() != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed when SCard tickets",
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"ticketNum": ticketNum.Val(),
		})
	}
}

func SellTicket(c *gin.Context) {
	ticketNum := utils.Redis.SCard(context.Background(), "tickets")
	if ticketNum.Err() != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed when SCard tickets",
		})
		return
	}
	if ticketNum.Val() == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "sold out",
		})
		return
	}

	rwmutex.Lock()
	ticketID := utils.Redis.SPop(context.Background(), "tickets")
	uID, err := strconv.ParseInt(c.Query("uid"), 10, 64)
	if err :=
		utils.Publish(utils.MQ4UserBuy,
			"user_buy_exchange",
			"user_buy_key",
			fmt.Sprintf("(now(), now(), %d, %s), ", uID, ticketID.Val())); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed when publish to rabbitmq",
		})
		return
	}
	if err :=
		utils.Publish(utils.MQ4SellTicket,
			"sell_ticket_exchange",
			"sell_ticket_key",
			fmt.Sprintf("%s, ", ticketID.Val())); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed when publish to rabbitmq",
		})
		return
	}
	rwmutex.Unlock()
	if ticketID.Err() != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed when SPop ticket",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "uid is not valid",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "success",
		"uID":      uID,
		"ticketID": ticketID.Val(),
	})
}

func UserBuyHandler() {
	msg, err := utils.Consume(utils.MQ4UserBuy)
	if err != nil {
		return
	}

	cnt := 0
	var buffer bytes.Buffer
	for d := range msg {
		//fmt.Printf("Received a message: %s\n", d.Body)
		if cnt == 0 {
			buffer.WriteString("insert into user2ticket (created_at, updated_at, user_id, ticket_id) values ")
			buffer.WriteString(string(d.Body))
			cnt++
		} else if cnt == 99 {
			buffer.WriteString(string(d.Body))
			buffer.Truncate(buffer.Len() - 2)
			buffer.WriteString(";")
			if err := utils.DB.Exec(buffer.String()).Error; err != nil {
				fmt.Println(err)
				return
			}
			cnt = 0
			buffer.Reset()
		} else {
			buffer.WriteString(string(d.Body))
			cnt++
		}
	}
}

func SellTicketHandler() {
	msg, err := utils.Consume(utils.MQ4SellTicket)
	if err != nil {
		return
	}

	cnt := 0
	var buffer bytes.Buffer
	for d := range msg {
		//fmt.Printf("Received a message: %s\n", d.Body)
		if cnt == 0 {
			buffer.WriteString("update ticket set status = 1 where id in (")
			buffer.WriteString(string(d.Body))
			cnt++
		} else if cnt == 99 {
			buffer.WriteString(string(d.Body))
			buffer.Truncate(buffer.Len() - 2)
			buffer.WriteString(");")
			if err := utils.DB.Exec(buffer.String()).Error; err != nil {
				fmt.Println(err)
				return
			}
			cnt = 0
			buffer.Reset()
		} else {
			buffer.WriteString(string(d.Body))
			cnt++
		}
	}
}
