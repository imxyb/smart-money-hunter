package oklink

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

var Api *API

type API struct {
	c      *req.Client
	apiKey string
	host   string
}

func InitAPI(apiKey, host string) {
	Api = &API{
		c: req.C().
			SetCommonRetryCount(3).
			SetCommonRetryBackoffInterval(5*time.Second, time.Minute).
			SetCommonRetryCondition(func(resp *req.Response, err error) bool {
				return resp.GetStatusCode() != http.StatusOK
			}),
		apiKey: apiKey,
		host:   host,
	}
}

func (a *API) GetNormalTransactionListByAddressAndToken(chainName, address, pt string, page, limit int) (*TransactionListResp, error) {
	url := fmt.Sprintf("%s/api/v5/explorer/address/transaction-list", a.host)
	resp := a.c.Get(url).SetHeader("Ok-Access-Key", a.apiKey).SetQueryParamsAnyType(map[string]interface{}{
		"address":        address,
		"chainShortName": chainName,
		"protocolType":   pt,
		"limit":          limit,
		"page":           page,
	}).Do()

	if resp.Err != nil {
		return nil, resp.Err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("get url failed, status code:%d", resp.GetStatusCode())
	}

	response := new(TransactionListResp)
	if err := resp.UnmarshalJson(&response); err != nil {
		return nil, err
	}
	return response, nil
}

func (a *API) GetToken20TransactionListByAddressAndToken(chainName, address, tokenAddress string, page, limit int) (*TransactionListResp, error) {
	url := fmt.Sprintf("%s/api/v5/explorer/address/transaction-list", a.host)
	resp := a.c.Get(url).SetHeader("Ok-Access-Key", a.apiKey).SetQueryParamsAnyType(map[string]interface{}{
		"address":              address,
		"chainShortName":       chainName,
		"protocolType":         "token_20",
		"limit":                limit,
		"page":                 page,
		"tokenContractAddress": strings.ToLower(tokenAddress),
	}).Do()

	if resp.Err != nil {
		return nil, resp.Err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("get url failed, status code:%d", resp.GetStatusCode())
	}

	response := new(TransactionListResp)
	if err := resp.UnmarshalJson(&response); err != nil {
		return nil, err
	}
	return response, nil
}

func (a *API) GetToken20TransactionListByAddress(chainName, address string, page, limit int) (*TransactionListResp, error) {
	url := fmt.Sprintf("%s/api/v5/explorer/address/transaction-list", a.host)
	resp := a.c.Get(url).SetHeader("Ok-Access-Key", a.apiKey).SetQueryParamsAnyType(map[string]interface{}{
		"address":        address,
		"chainShortName": chainName,
		"protocolType":   "token_20",
		"limit":          limit,
		"page":           page,
	}).Do()

	if resp.Err != nil {
		return nil, resp.Err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("get url failed, status code:%d", resp.GetStatusCode())
	}

	response := new(TransactionListResp)
	if err := resp.UnmarshalJson(&response); err != nil {
		return nil, err
	}
	return response, nil
}

func (a *API) GetTransactionDetail(chainName, txID string) (*TransactionDetailResp, error) {
	url := fmt.Sprintf("%s/api/v5/explorer/transaction/transaction-fills", a.host)

	resp := a.c.Get(url).SetHeader("Ok-Access-Key", a.apiKey).SetQueryParamsAnyType(map[string]interface{}{
		"txid":           txID,
		"chainShortName": chainName,
	}).Do()

	if resp.Err != nil {
		return nil, resp.Err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("get url failed, status code:%d", resp.GetStatusCode())
	}

	response := new(TransactionDetailResp)
	if err := resp.UnmarshalJson(&response); err != nil {
		return nil, err
	}

	return response, nil
}

func (a *API) GetTokenHolderList(chainName, address string, page, limit int) (*TokenHolderListResp, error) {
	url := fmt.Sprintf("%s/api/v5/explorer/token/position-list", a.host)

	resp := a.c.Get(url).SetHeader("Ok-Access-Key", a.apiKey).SetQueryParamsAnyType(map[string]interface{}{
		"chainShortName":       chainName,
		"tokenContractAddress": strings.ToLower(address),
		"limit":                limit,
		"page":                 page,
	}).Do()

	if resp.Err != nil {
		return nil, resp.Err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("get url failed, status code:%d", resp.GetStatusCode())
	}

	response := new(TokenHolderListResp)
	if err := resp.UnmarshalJson(&response); err != nil {
		return nil, err
	}

	return response, nil
}

func (a *API) GetAddressDetail(chainName, address string) (*AddressDetailResp, error) {
	url := fmt.Sprintf("%s/api/v5/explorer/address/address-summary", a.host)

	resp := a.c.Get(url).SetHeader("Ok-Access-Key", a.apiKey).SetQueryParamsAnyType(map[string]interface{}{
		"address":        address,
		"chainShortName": chainName,
	}).Do()

	if resp.Err != nil {
		return nil, resp.Err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("get url failed, status code:%d", resp.GetStatusCode())
	}

	response := new(AddressDetailResp)
	if err := resp.UnmarshalJson(&response); err != nil {
		return nil, err
	}

	return response, nil
}

func (a *API) GetAddressBalance(chainName, address, tokenAddress string, page, limit int) (*AddressBalanceResp, error) {
	url := fmt.Sprintf("%s/api/v5/explorer/address/address-balance-fills", a.host)

	resp := a.c.Get(url).SetHeader("Ok-Access-Key", a.apiKey).SetQueryParamsAnyType(map[string]interface{}{
		"address":              strings.ToLower(address),
		"chainShortName":       chainName,
		"protocolType":         "token_20",
		"tokenContractAddress": strings.ToLower(tokenAddress),
		"limit":                limit,
		"page":                 page,
	}).Do()

	if resp.Err != nil {
		return nil, resp.Err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("get url failed, status code:%d", resp.GetStatusCode())
	}

	response := new(AddressBalanceResp)
	if err := resp.UnmarshalJson(&response); err != nil {
		return nil, err
	}

	return response, nil
}

func (a *API) GetGasFee(chainName string) (*GasFeeResp, error) {
	url := fmt.Sprintf("%s/api/v5/explorer/blockchain/fee", a.host)

	resp := a.c.Get(url).SetHeader("Ok-Access-Key", a.apiKey).SetQueryParamsAnyType(map[string]interface{}{
		"chainShortName": chainName,
	}).Do()

	if resp.Err != nil {
		return nil, resp.Err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("get url failed, status code:%d", resp.GetStatusCode())
	}

	response := new(GasFeeResp)
	if err := resp.UnmarshalJson(&response); err != nil {
		return nil, err
	}

	return response, nil
}
