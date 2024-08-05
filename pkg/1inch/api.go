package inch

import (
	"fmt"
	"net/http"
	"time"

	"github.com/imroc/req/v3"
	"github.com/mitchellh/mapstructure"
	"smart-money/pkg/util"
)

const (
	APIBASE       = "https://api.1inch.io/v5.0"
	BroadcastBASE = "https://tx-gateway.1inch.io/v1.1"
)

var reqC = req.C().DevMode().
	SetCommonRetryCount(3).
	SetCommonRetryBackoffInterval(3*time.Second, time.Minute).
	SetCommonRetryCondition(func(resp *req.Response, err error) bool {
		return resp.GetStatusCode() != http.StatusOK
	})

func Quote(chainName, fromTokenAddress, toTokenAddress string, amount int64) (*QuoteResp, error) {
	url := fmt.Sprintf("%s/%d/%s", APIBASE, util.ChainIDMap[chainName], "quote")
	resp := reqC.Get(url).SetQueryParamsAnyType(map[string]interface{}{
		"fromTokenAddress": fromTokenAddress,
		"toTokenAddress":   toTokenAddress,
		"amount":           amount,
	}).Do()

	if resp.Err != nil {
		return nil, resp.Err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("get url failed, status code:%d", resp.GetStatusCode())
	}

	response := new(QuoteResp)
	if err := resp.UnmarshalJson(&response); err != nil {
		return nil, err
	}

	return response, nil
}

type SwapRequest struct {
	FromTokenAddress  string   `mapstructure:"fromTokenAddress" form:"fromTokenAddress" validate:"required"`
	ToTokenAddress    string   `mapstructure:"toTokenAddress" form:"toTokenAddress" validate:"required"`
	Amount            string   `mapstructure:"amount" form:"amount" validate:"required"`
	FromAddress       string   `mapstructure:"fromAddress" form:"fromAddress" validate:"required"`
	Slippage          float64  `mapstructure:"slippage" form:"slippage" validate:"required"`
	Protocols         string   `mapstructure:"protocols,omitempty" form:"protocols"`
	DestReceiver      string   `mapstructure:"destReceiver,omitempty" form:"destReceiver"`
	ReferrerAddress   string   `mapstructure:"referrerAddress,omitempty" form:"referrerAddress"`
	Fee               string   `mapstructure:"fee,omitempty" form:"fee"`
	DisableEstimate   bool     `mapstructure:"disableEstimate,omitempty" form:"disableEstimate"`
	Permit            string   `mapstructure:"permit,omitempty" form:"permit"`
	CompatibilityMode bool     `mapstructure:"compatibilityMode,omitempty" form:"compatibilityMode"`
	BurnChi           bool     `mapstructure:"burnChi,omitempty" form:"burnChi"`
	AllowPartialFill  bool     `mapstructure:"allowPartialFill,omitempty" form:"allowPartialFill"`
	Parts             int      `mapstructure:"parts,omitempty" form:"parts"`
	MainRouteParts    int      `mapstructure:"mainRouteParts,omitempty" form:"mainRouteParts"`
	ConnectorTokens   []string `mapstructure:"connectorTokens,omitempty" form:"connectorTokens"`
	ComplexityLevel   int      `mapstructure:"complexityLevel,omitempty" form:"complexityLevel"`
	GasLimit          string   `mapstructure:"gasLimit,omitempty" form:"gasLimit"`
	GasPrice          string   `mapstructure:"gasPrice,omitempty" form:"gasPrice"`
}

func Swap(chainName string, req *SwapRequest) (*SwapResp, error) {
	reqm := make(map[string]any)
	if err := mapstructure.Decode(req, &reqm); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/%d/%s", APIBASE, util.ChainIDMap[chainName], "swap")
	resp := reqC.Get(url).SetQueryParamsAnyType(reqm).Do()

	if resp.Err != nil {
		return nil, resp.Err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("get url failed, status code:%d", resp.GetStatusCode())
	}

	response := new(SwapResp)
	if err := resp.UnmarshalJson(&response); err != nil {
		return nil, err
	}

	return response, nil
}

func CheckAllowance(chainName, tokenAddress, walletAddress string) (string, error) {
	url := fmt.Sprintf("%s/%d/%s", APIBASE, util.ChainIDMap[chainName], "approve/allowance")
	resp := reqC.Get(url).SetQueryParamsAnyType(map[string]interface{}{
		"tokenAddress":  tokenAddress,
		"walletAddress": walletAddress,
	}).Do()

	if resp.Err != nil {
		return "0", resp.Err
	}
	if resp.IsErrorState() {
		return "0", fmt.Errorf("get url failed, status code:%d", resp.GetStatusCode())
	}

	response := make(map[string]string)
	if err := resp.UnmarshalJson(&response); err != nil {
		return "0", err
	}

	return response["allowance"], nil
}

func ApproveTransaction(chainName, tokenAddress, amount string) (*ApproveTransactionResp, error) {
	url := fmt.Sprintf("%s/%d/%s", APIBASE, util.ChainIDMap[chainName], "approve/transaction")
	resp := reqC.Get(url).SetQueryParamsAnyType(map[string]interface{}{
		"tokenAddress": tokenAddress,
		"amount":       amount,
	}).Do()

	if resp.Err != nil {
		return nil, resp.Err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("get url failed, status code:%d", resp.GetStatusCode())
	}

	response := new(ApproveTransactionResp)
	if err := resp.UnmarshalJson(&response); err != nil {
		return nil, err
	}

	return response, nil
}

func GetSpender(chainName string) (string, error) {
	url := fmt.Sprintf("%s/%d/%s", APIBASE, util.ChainIDMap[chainName], "approve/spender")
	resp := reqC.Get(url).Do()

	if resp.Err != nil {
		return "", resp.Err
	}
	if resp.IsErrorState() {
		return "", fmt.Errorf("get url failed, status code:%d", resp.GetStatusCode())
	}

	response := make(map[string]string)
	if err := resp.UnmarshalJson(&response); err != nil {
		return "", err
	}

	return response["address"], nil
}

func Broadcast(chainID int64, body any) (string, error) {
	url := fmt.Sprintf("%s/%d/%s", BroadcastBASE, chainID, "broadcast")
	resp := reqC.Post(url).SetBodyJsonMarshal(body).Do()

	if resp.Err != nil {
		return "", resp.Err
	}
	if resp.IsErrorState() {
		return "", fmt.Errorf("get url failed, status code:%d", resp.GetStatusCode())
	}

	response := make(map[string]string)
	if err := resp.UnmarshalJson(&response); err != nil {
		return "", err
	}

	return response["transactionHash"], nil
}
