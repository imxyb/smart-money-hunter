package v1

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"smart-money/pkg/errcode"
	"smart-money/pkg/model"
	"smart-money/pkg/response"
)

type FollowTradeReq struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type FollowTradeDetail struct {
	ChainName               string  `json:"chain_name"`
	WalletAddress           string  `json:"wallet_address"`
	FollowAddress           string  `json:"follow_address"`
	FollowAddressBuyTxHash  string  `json:"follow_address_buy_tx_hash"`
	WalletAddressBuyTxHash  string  `json:"wallet_address_buy_tx_hash"`
	WalletAddressBuyGas     float64 `json:"wallet_address_buy_gas"`
	BuyTokenAddress         string  `json:"buy_token_address"`
	BuySymbol               string  `json:"buy_symbol"`
	BuyTokenDecimal         int     `json:"buy_token_decimal"`
	WalletAddressBuyTime    int64   `json:"wallet_address_buy_time"`
	FollowAddressBuyTime    int64   `json:"follow_address_buy_time"`
	WalletAddressBuyAmount  float64 `json:"wallet_address_buy_amount"`
	FollowAddressBuyAmount  float64 `json:"follow_address_buy_amount"`
	WalletAddressSellAmount float64 `json:"wallet_address_sell_amount"`
	IsSellPrincipal         int     `json:"is_sell_principal"`
	SellPrincipalTxHash     string  `json:"sell_principal_tx_hash"`
	Status                  int     `json:"status"`
	FailReason              string  `json:"fail_reason"`
}

type FollowTradeResp []*FollowTradeDetail

func ListFollowTrade(c *gin.Context) {
	var req FollowTradeReq
	if err := c.Bind(&req); err != nil {
		response.BadRequest(c, errcode.ListFollowTradeParamsError, err)
	}

	page := req.Page
	pageSize := req.PageSize
	if page == 0 {
		page = defaultPage
	}
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	offset := (page - 1) * pageSize
	var followTrades []*model.FollowTrade
	err := model.GetDB().Offset(offset).Limit(pageSize).Find(&followTrades).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		response.InternalServerError(c, err)
		return
	}

	var count int64
	err = model.GetDB().Model(&model.FollowTrade{}).Count(&count).Error
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	resp := make([]*FollowTradeDetail, 0, len(followTrades))
	for _, followTrade := range followTrades {
		resp = append(resp, &FollowTradeDetail{
			ChainName:               followTrade.ChainName,
			WalletAddress:           followTrade.WalletAddress,
			FollowAddress:           followTrade.FollowAddress,
			FollowAddressBuyTxHash:  followTrade.FollowAddressBuyTxHash,
			WalletAddressBuyTxHash:  followTrade.WalletAddressBuyTxHash,
			WalletAddressBuyGas:     followTrade.WalletAddressBuyGas,
			BuyTokenAddress:         followTrade.BuyTokenAddress,
			BuySymbol:               followTrade.BuySymbol,
			BuyTokenDecimal:         followTrade.BuyTokenDecimal,
			WalletAddressBuyTime:    followTrade.WalletAddressBuyTime,
			FollowAddressBuyTime:    followTrade.FollowAddressBuyTime,
			WalletAddressBuyAmount:  followTrade.WalletAddressBuyAmount,
			FollowAddressBuyAmount:  followTrade.FollowAddressBuyAmount,
			WalletAddressSellAmount: followTrade.WalletAddressSellAmount,
			IsSellPrincipal:         followTrade.IsSellPrincipal,
			SellPrincipalTxHash:     followTrade.SellPrincipalTxHash,
			Status:                  followTrade.Status,
			FailReason:              followTrade.FailReason,
		})
	}

	response.OKList(c, count, &resp)
}
