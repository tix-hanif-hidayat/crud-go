package model

import "time"

type Hotel struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Room struct {
	Id      int    `json:"id" gorm:"primaryKey;autoIncrement"`
	HotelId int    `json:"hotel_id"`
	Name    string `json:"name"`
	Price   int    `json:"price"`
}

type ExchangeRate struct {
	BaseCurrency  string    `json:"base_currency"`
	QuoteCurrency string    `json:"quote_currency"`
	Rate          float64   `json:"rate"`
	LastUpdated   time.Time `json:"last_updated"`
}
