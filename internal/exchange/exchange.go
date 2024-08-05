package exchange

import (
	"github.com/ethereum/go-ethereum/common"
	inch "smart-money/pkg/1inch"
	"smart-money/pkg/eth"
)

type Exchange struct {
	chainName  string
	req        *inch.SwapRequest
	privateKey string
}

func NewExchange(chainName, privateKey string, req *inch.SwapRequest) *Exchange {
	return &Exchange{chainName: chainName, req: req, privateKey: privateKey}
}

func (e *Exchange) Swap() (common.Hash, error) {
	swapResp, err := inch.Swap(e.chainName, e.req)
	if err != nil {
		return [32]byte{}, err
	}
	return eth.Client.SendTransaction(e.privateKey, &swapResp.Tx)
}

func (e *Exchange) CheckAllowance() (string, error) {
	return inch.CheckAllowance(e.chainName, e.req.FromTokenAddress, e.req.FromAddress)
}

func (e *Exchange) ApproveTransaction(isInf bool) (common.Hash, error) {
	amount := e.req.Amount
	if isInf {
		amount = ""
	}
	approveResp, err := inch.ApproveTransaction(e.chainName, e.req.FromTokenAddress, amount)
	if err != nil {
		return [32]byte{}, err
	}

	return eth.Client.SendTransaction(e.privateKey, &inch.TxData{
		From:     e.req.FromAddress,
		To:       approveResp.To,
		Data:     approveResp.Data,
		Value:    approveResp.Value,
		GasPrice: approveResp.GasPrice,
	})
}
