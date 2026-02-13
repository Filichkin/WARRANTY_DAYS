package models

import "time"

type Claims struct {
	ID           int64         `json:"id"`
	VIN          string        `json:"vin"`
	RetailDate   time.Time     `json:"retail_date"`
	RoOpenDate   time.Time     `json:"ro_open_date"`
	RoCloseDate  time.Time     `json:"ro_close_date"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}