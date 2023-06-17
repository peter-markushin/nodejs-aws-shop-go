package models

type Stock struct {
	ProductID string `json:"-" gorm:"primaryKey"`
	Count     uint8  `json:"count"`
}
