package model

import "gorm.io/gorm"

type TokenTransactionCollect struct {
	gorm.Model
	TaskName    string  `json:"task_name" gorm:"column:task_name"`
	ChainName   string  `json:"chain_name" gorm:"column:chain_name"`
	Address     string  `json:"address" gorm:"column:address"`
	BlockHeight int64   `json:"block_height" gorm:"column:block_height"`
	TxHash      string  `json:"tx_hash" gorm:"column:tx_hash"`
	TxTime      uint64  `json:"tx_time" gorm:"column:tx_time"`
	BuyAddress  string  `json:"buy_address" gorm:"column:buy_address"`
	BuySymbol   string  `json:"buy_symbol" gorm:"column:buy_symbol"`
	BuyAmount   float64 `json:"buy_amount" gorm:"column:buy_amount"`
	SellAddress string  `json:"sell_address" gorm:"column:sell_address"`
	SellAmount  float64 `json:"sell_amount" gorm:"column:sell_amount"`
	SellSymbol  string  `json:"sell_symbol" gorm:"column:sell_symbol"`
}

func (t *TokenTransactionCollect) TableName() string {
	return "token_transaction_collect"
}

func CreateTokenTransactionCollect(t *TokenTransactionCollect) error {
	return db.Create(t).Error
}

func init() {
	registerTable(&TokenTransactionCollect{})
}
