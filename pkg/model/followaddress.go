package model

import "gorm.io/gorm"

const (
	FollowAddressStatusNormal = 1
	FollowAddressStatusStop   = 2
)

type FollowAddress struct {
	gorm.Model
	Address         string `json:"address" gorm:"column:address;type:varchar(255);not null;default:'';comment:地址"`
	ChainName       string `json:"chain_name" gorm:"column:chain_name;type:varchar(255);not null;default:'';comment:链名称"`
	LastErc20TxHash string `json:"last_erc20_tx_hash" gorm:"column:last_erc20_tx_hash;type:varchar(255);not null;default:'';comment:最后一次erc20交易hash"`
	LastErc20TxTime int64  `json:"last_erc20_tx_time" gorm:"column:last_erc20_tx_time;type:bigint(20);not null;default:0;comment:最后一次erc20交易时间"`
	Status          int    `json:"status" gorm:"column:status;type:int(11);not null;default:0;comment:状态"`
}

func (f *FollowAddress) TableName() string {
	return "follow_address"
}

func CreateFollowAddress(f *FollowAddress) error {
	return db.Create(f).Error
}

func SaveFollowAddress(f *FollowAddress) error {
	return db.Save(f).Error
}

func init() {
	registerTable(&FollowAddress{})
}
