package model

import "gorm.io/gorm"

type AddressTrade struct {
	gorm.Model
	Address      string  `json:"address" gorm:"column:address;type:varchar(255);not null;default:'';comment:地址"`
	ChainName    string  `json:"chain_name" gorm:"column:chain_name"`
	FirstTxTime  uint64  `json:"first_tx_time" gorm:"column:first_tx_time"`
	LastTxTime   uint64  `json:"last_tx_time" gorm:"column:last_tx_time"`
	BuyAddress   string  `json:"buy_address" gorm:"column:buy_address"`
	BuySymbol    string  `json:"buy_symbol" gorm:"column:buy_symbol"`
	SellAddress  string  `json:"sell_address" gorm:"column:sell_address"`
	SellSymbol   string  `json:"sell_symbol" gorm:"column:sell_symbol"`
	BuyTotalUsd  float64 `json:"buy_total_usd" gorm:"column:buy_total_usd"`
	SellTotalUsd float64 `json:"sell_total_usd" gorm:"column:sell_total_usd"`
	Profit       float64 `json:"profit" gorm:"column:profit"`
}

func (a *AddressTrade) TableName() string {
	return "address_trade"
}

func CreateAddressTrade(addressTrade *AddressTrade) error {
	return db.Create(addressTrade).Error
}

func init() {
	registerTable(&AddressTrade{})
}
