package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Role     uint
	Username string
	Password string
	Balance  float64
	//Cart      []*Goods `gorm:"many2many:user_cart;"`
	Published []*Goods
	Favorites []*Goods `gorm:"many2many:user_favorites;"`
}

type Goods struct {
	gorm.Model
	Title       string
	Price       float64
	Description string
	UserID      uint
	//Quantity    uint
}
type Cart struct {
	gorm.Model
	UserID   uint
	GoodsID  uint
	Quantity uint
}
