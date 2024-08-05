package model

import "gorm.io/gorm"

const (
	FollowTradeStatusSuccess = 1
	FollowTradeStatusFail    = 2
	FollowTradeStatusFinish  = 3
)

type FollowTrade struct {
	gorm.Model
	ChainName               string  `json:"chain_name" gorm:"column:chain_name;type:varchar(255);not null;default:'';comment:链名称"`
	WalletAddress           string  `json:"wallet_address" gorm:"column:wallet_addreess;type:varchar(255);not null;default:'';comment:地址"`
	FollowAddress           string  `json:"follow_address" gorm:"column:follow_address;type:varchar(255);not null;default:'';comment:关注地址"`
	FollowAddressBuyTxHash  string  `json:"follow_address_buy_tx_hash" gorm:"column:follow_address_buy_tx_hash;type:varchar(255);not null;default:'';comment:关注地址购买交易哈希"`
	WalletAddressBuyTxHash  string  `json:"wallet_address_buy_tx_hash" gorm:"column:wallet_address_buy_tx_hash;type:varchar(255);not null;default:'';comment:钱包地址购买交易哈希"`
	WalletAddressBuyGas     float64 `json:"wallet_address_buy_gas" gorm:"column:wallet_address_buy_gas;type:decimal(20,8);not null;default:0;comment:钱包地址购买手续费"`
	BuyTokenAddress         string  `json:"buy_token_address" gorm:"column:buy_token_address;type:varchar(255);not null;default:'';comment:购买币种地址"`
	BuySymbol               string  `json:"buy_symbol" gorm:"column:buy_symbol;type:varchar(255);not null;default:'';comment:购买币种符号"`
	BuyTokenDecimal         int     `json:"buy_token_decimal" gorm:"column:buy_token_decimal;type:int(11);not null;default:0;comment:购买币种精度"`
	WalletAddressBuyTime    int64   `json:"wallet_address_buy_time" gorm:"column:wallet_address_buy_time;type:bigint(20);not null;default:0;comment:钱包地址购买时间"`
	FollowAddressBuyTime    int64   `json:"follow_address_buy_time" gorm:"column:follow_address_buy_time;type:bigint(20);not null;default:0;comment:关注地址购买时间"`
	WalletAddressBuyAmount  float64 `json:"wallet_address_buy_amount" gorm:"column:wallet_address_buy_amount;type:decimal(20,8);not null;default:0;comment:钱包地址购买数量"`
	FollowAddressBuyAmount  float64 `json:"follow_address_buy_amount" gorm:"column:follow_address_buy_amount;type:decimal(20,8);not null;default:0;comment:关注地址购买数量"`
	WalletAddressSellAmount float64 `json:"wallet_address_sell_amount" gorm:"column:wallet_address_sell_amount;type:decimal(20,8);not null;default:0;comment:钱包地址卖出数量"`
	IsSellPrincipal         int     `json:"is_sell_principal" gorm:"column:is_sell_principal;type:tinyint(1);not null;default:0;comment:是否卖出本金"`
	SellPrincipalTxHash     string  `json:"sell_principal_tx_hash" gorm:"column:sell_principal_tx_hash;type:varchar(255);not null;default:'';comment:卖出本金交易哈希"`
	Status                  int     `json:"status" gorm:"column:status;type:tinyint(1);not null;default:0;comment:状态"`
	FailReason              string  `json:"fail_reason" gorm:"column:fail_reason;type:text;comment:失败原因"`
}

func (f *FollowTrade) TableName() string {
	return "follow_trade"
}

func CreateFollowTrade(followTrade *FollowTrade) error {
	return db.Create(followTrade).Error
}

func SaveFollowTrade(followTrade *FollowTrade) error {
	return db.Save(followTrade).Error
}

func init() {
	registerTable(&FollowTrade{})
}
