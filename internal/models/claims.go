package models

import "time"

type Claim struct {
	ID          int64      `json:"id" gorm:"primaryKey"`
	VIN         string     `json:"vin" gorm:"column:vin;not null;index"`
	RetailDate  time.Time  `json:"retail_date" gorm:"column:retail_date;type:date;not null"`
	RoOpenDate  time.Time  `json:"ro_open_date" gorm:"column:ro_open_date;type:date;not null"`
	RoCloseDate time.Time `json:"ro_close_date" gorm:"column:ro_close_date;type:date;not null"`
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (Claim) TableName() string { 
	return "claims" 
}