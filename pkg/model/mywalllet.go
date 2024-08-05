package model

import "gorm.io/gorm"

const (
	MyWalletStatusEnable  = 1
	MyWalletStatusDisable = 2
)

type MyWallet struct {
	gorm.Model
	Address        string  `json:"address" gorm:"column:address;type:varchar(255);not null;default:'';comment:地址"`
	ChainName      string  `json:"chain_name" gorm:"column:chain_name;type:varchar(255);not null;default:'';comment:链名称"`
	Name           string  `json:"name" gorm:"column:name;type:varchar(255);not null;default:'';comment:名称"`
	PrivateKey     string  `json:"private_key" gorm:"column:private_key;type:varchar(255);not null;default:'';comment:私钥"`
	EachSellAmount float64 `json:"each_sell_amount" gorm:"column:each_sell_amount;type:decimal(10,5);not null;default:0.00;comment:每次卖出数量"`
	Status         int     `json:"status" gorm:"column:status;type:int(11);not null;default:0;comment:状态"`
}

func (m *MyWallet) TableName() string {
	return "my_wallet"
}

func CreateMyWallet(data *MyWallet) error {
	return db.Create(data).Error
}

func SaveMyWallet(data *MyWallet) error {
	return db.Save(data).Error
}

func init() {
	registerTable(&MyWallet{})
}
