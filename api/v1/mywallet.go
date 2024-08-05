package v1

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"smart-money/pkg/errcode"
	"smart-money/pkg/model"
	"smart-money/pkg/response"
	"smart-money/pkg/util"
)

type WalletDetail struct {
	Address        string    `json:"address"`
	Name           string    `json:"name"`
	ChainName      string    `json:"chain_name"`
	EachSellAmount float64   `json:"each_sell_amount"`
	Status         int       `json:"status"`
	CreateAt       time.Time `json:"create_at"`
}

type ListWalletReq struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type ListWalletResp []*WalletDetail

func ListWallet(c *gin.Context) {
	var req ListWalletReq
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
	var myWallets []*model.MyWallet
	err := model.GetDB().Offset(offset).Limit(pageSize).Find(&myWallets).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		response.InternalServerError(c, err)
		return
	}

	var count int64
	err = model.GetDB().Model(&model.MyWallet{}).Count(&count).Error

	resp := make(ListWalletResp, 0, len(myWallets))
	for _, wallet := range myWallets {
		resp = append(resp, &WalletDetail{
			Address:        wallet.Address,
			Name:           wallet.Name,
			ChainName:      wallet.ChainName,
			EachSellAmount: wallet.EachSellAmount,
			Status:         wallet.Status,
			CreateAt:       wallet.CreatedAt,
		})
	}

	response.OKList(c, count, resp)
}

type CreateWalletReq struct {
	Address        string  `json:"address"`
	PrivateKey     string  `json:"private_key"`
	ChainName      string  `json:"chain_name"`
	Name           string  `json:"name"`
	EachSellAmount float64 `json:"each_sell_amount"`
	Status         int     `json:"status"`
}

type CreateWalletResp struct {
}

func CreateWallet(c *gin.Context) {
	var req CreateWalletReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, errcode.SaveWalletParamsError, err)
		return
	}

	if req.Address == "" {
		response.BadRequest(c, errcode.SaveWalletParamsError, fmt.Errorf("address is empty"))
		return
	}
	if !util.CheckChainName(req.ChainName) {
		response.BadRequest(c, errcode.SaveWalletParamsError, fmt.Errorf("chain name is invalid"))
		return
	}
	if req.Name == "" {
		response.BadRequest(c, errcode.SaveWalletParamsError, fmt.Errorf("name is empty"))
		return
	}
	if req.EachSellAmount <= 0 {
		response.BadRequest(c, errcode.SaveWalletParamsError, fmt.Errorf("each sell amount is invalid"))
		return
	}
	if req.Status != model.MyWalletStatusEnable && req.Status != model.MyWalletStatusDisable {
		response.BadRequest(c, errcode.SaveWalletParamsError, fmt.Errorf("status is invalid"))
		return
	}

	var count int64
	err := model.GetDB().Model(&model.MyWallet{}).Where("address = ? and chain_name=?", req.Address, req.ChainName).Count(&count).Error
	if err != nil {
		response.InternalServerError(c, err)
		return
	}
	if count > 0 {
		response.BadRequest(c, errcode.SaveWalletExistError, fmt.Errorf("address is exist"))
		return
	}

	myWallet := &model.MyWallet{
		Address:        req.Address,
		PrivateKey:     req.PrivateKey,
		ChainName:      req.ChainName,
		Name:           req.Name,
		EachSellAmount: req.EachSellAmount,
		Status:         model.MyWalletStatusEnable,
	}

	err = model.GetDB().Create(myWallet).Error
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, &CreateWalletResp{})
}

type UpdateWalletReq struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	EachSellAmount float64 `json:"each_sell_amount"`
	Status         int     `json:"status"`
}

type UpdateWalletResp struct {
}

func UpdateWallet(c *gin.Context) {
	var req UpdateWalletReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, errcode.SaveWalletParamsError, err)
		return
	}

	if req.Name == "" {
		response.BadRequest(c, errcode.SaveWalletParamsError, fmt.Errorf("name is empty"))
		return
	}
	if req.EachSellAmount <= 0 {
		response.BadRequest(c, errcode.SaveWalletParamsError, fmt.Errorf("each sell amount is invalid"))
		return
	}
	if req.Status != model.MyWalletStatusEnable && req.Status != model.MyWalletStatusDisable {
		response.BadRequest(c, errcode.SaveWalletParamsError, fmt.Errorf("status is invalid"))
		return
	}

	var myWallet model.MyWallet
	err := model.GetDB().Where("id = ?", req.ID).First(&myWallet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.BadRequest(c, errcode.SaveWalletNotExistError, err)
			return
		}
		response.InternalServerError(c, err)
		return
	}

	myWallet.Name = req.Name
	myWallet.EachSellAmount = req.EachSellAmount
	myWallet.Status = req.Status

	err = model.GetDB().Save(&myWallet).Error
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, &UpdateWalletResp{})
}

type DeleteWalletReq struct {
	ID int `json:"id"`
}

type DeleteWalletResp struct {
}

func DeleteWallet(c *gin.Context) {
	var req DeleteWalletReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, errcode.SaveWalletParamsError, err)
		return
	}

	err := model.GetDB().Where("id = ?", req.ID).Unscoped().Delete(&model.MyWallet{}).Error
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, &DeleteWalletResp{})
}
