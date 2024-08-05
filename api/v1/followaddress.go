package v1

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"smart-money/pkg/errcode"
	"smart-money/pkg/model"
	"smart-money/pkg/oklink"
	"smart-money/pkg/response"
	"smart-money/pkg/util"
)

type FollowAddressDetail struct {
	ChainName       string `json:"chain_name"`
	Address         string `json:"address"`
	LastErc20TxHash string `json:"last_erc20_tx_hash"`
	LastErc20TxTime int64  `json:"last_erc20_tx_time"`
	Status          int    `json:"status"`
}

type ListFollowAddressReq struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type ListFollowAddressResp []*FollowAddressDetail

func ListFollowAddress(c *gin.Context) {
	var req ListFollowAddressReq
	if err := c.Bind(&req); err != nil {
		response.BadRequest(c, errcode.ListFollowAddressParamsError, err)
		return
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
	var followAddresses []*model.FollowAddress
	err := model.GetDB().Offset(offset).Limit(pageSize).Find(&followAddresses).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		response.InternalServerError(c, err)
		return
	}

	var count int64
	err = model.GetDB().Model(&model.FollowAddress{}).Count(&count).Error
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	resp := make(ListFollowAddressResp, 0, len(followAddresses))
	for _, followAddress := range followAddresses {
		resp = append(resp, &FollowAddressDetail{
			ChainName:       followAddress.ChainName,
			Address:         followAddress.Address,
			Status:          followAddress.Status,
			LastErc20TxHash: followAddress.LastErc20TxHash,
			LastErc20TxTime: followAddress.LastErc20TxTime,
		})
	}

	response.OKList(c, count, resp)
}

type CreateFollowAddressReq struct {
	ChainName string `json:"chain_name"`
	Address   string `json:"address"`
	Status    int    `json:"status"`
}

type CreateFollowAddressResp struct {
}

func CreateFollowAddress(c *gin.Context) {
	var req CreateFollowAddressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, errcode.SaveFollowAddressParamsError, err)
		return
	}

	if !util.CheckChainName(req.ChainName) {
		response.BadRequest(c, errcode.SaveFollowAddressParamsError, fmt.Errorf("chain name error"))
		return
	}

	if req.Address == "" {
		response.BadRequest(c, errcode.SaveFollowAddressParamsError, fmt.Errorf("address error"))
		return
	}

	if req.Status != model.FollowAddressStatusNormal && req.Status != model.FollowAddressStatusStop {
		response.BadRequest(c, errcode.SaveFollowAddressParamsError, fmt.Errorf("status error"))
		return
	}

	// 查找最新一条
	latestTx, err := oklink.Api.GetToken20TransactionListByAddress(req.ChainName, req.Address, 1, 1)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}
	if len(latestTx.Data) == 0 || len(latestTx.Data[0].TransactionLists) == 0 {
		response.BadRequest(c, errcode.SaveFollowAddressParamsError, fmt.Errorf("can get follow address last erc20 tx hash"))
		return
	}
	latestTxTxHash := latestTx.Data[0].TransactionLists[0].TxId

	followAddress := &model.FollowAddress{
		ChainName:       req.ChainName,
		Address:         req.Address,
		Status:          req.Status,
		LastErc20TxHash: latestTxTxHash,
	}

	latestTxTime, err := strconv.Atoi(latestTx.Data[0].TransactionLists[0].TransactionTime)
	if err != nil {
		response.InternalServerError(c, err)
		return
	}
	followAddress.LastErc20TxTime = int64(latestTxTime)

	if err := model.CreateFollowAddress(followAddress); err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, &CreateFollowAddressResp{})
}

type UpdateFollowAddressReq struct {
	ID     int `json:"id"`
	Status int `json:"status"`
}

type UpdateFollowAddressResp struct {
}

func UpdateFollowAddress(c *gin.Context) {
	var req UpdateFollowAddressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, errcode.SaveFollowAddressParamsError, err)
		return
	}

	if req.ID <= 0 {
		response.BadRequest(c, errcode.SaveFollowAddressParamsError, fmt.Errorf("id error"))
		return
	}

	if req.Status != model.FollowAddressStatusNormal && req.Status != model.FollowAddressStatusStop {
		response.BadRequest(c, errcode.SaveFollowAddressParamsError, fmt.Errorf("status error"))
		return
	}

	followAddress := new(model.FollowAddress)
	err := model.GetDB().Where("id = ?", req.ID).First(followAddress).Error
	if err != nil {
		response.InternalServerError(c, err)
		return
	}
	followAddress.Status = req.Status

	if err := model.SaveFollowAddress(followAddress); err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, &UpdateFollowAddressResp{})
}

type DeleteFollowAddressReq struct {
	ID int `json:"id"`
}

type DeleteFollowAddressResp struct {
}

func DeleteFollowAddress(c *gin.Context) {
	var req DeleteFollowAddressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, errcode.SaveFollowAddressParamsError, err)
		return
	}

	if req.ID <= 0 {
		response.BadRequest(c, errcode.SaveFollowAddressParamsError, fmt.Errorf("id error"))
		return
	}

	if err := model.GetDB().Unscoped().Where("id=?", req.ID).Error; err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, &DeleteFollowAddressResp{})
}
