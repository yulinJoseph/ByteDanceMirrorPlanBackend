package model

import "gorm.io/gorm"

const (
	Selling = iota
	Sold
)

type Ticket struct {
	gorm.Model
	Status  int
	Details string
}

func (table *Ticket) TableName() string {
	return "ticket"
}

func Sell(t *Ticket) {
	t.Status++
}
