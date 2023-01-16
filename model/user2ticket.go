package model

import (
	"gorm.io/gorm"
)

type User2Ticket struct {
	gorm.Model
	UserID   uint
	TicketID uint
}

func (table *User2Ticket) TableName() string {
	return "user2ticket"
}
