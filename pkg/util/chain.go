package util

import (
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"smart-money/pkg/log"
)

var reqC = req.C()

type USDT struct {
	ContractAddress string
	Decimal         uint8
}

var USDTContractMap = map[string]*USDT{
	"eth": {
		ContractAddress: "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
		Decimal:         6,
	},
	"bsc": {
		ContractAddress: "0x55d398326f99059ff775485246999027b3197955",
		Decimal:         18,
	},
}

var MinSellUsd = map[string]float64{
	"eth": 30,
}

type MainToken struct {
	ContractAddress string
	Decimal         uint8
}

var MainTokenInfo = map[string]*MainToken{
	"eth": {
		ContractAddress: "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		Decimal:         18,
	},
	"bsc": {
		ContractAddress: "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
		Decimal:         18,
	},
}

var ChainIDMap = map[string]int64{
	"eth": 1,
	"bsc": 56,
}

var MainTokens = []string{
	"BNB", "WBNB", "ETH", "WETH", "USDT", "USDC", "DAI",
}

var ChainMainToken = map[string]string{
	"eth": "ETH",
	"bsc": "BNB",
}

var (
	ethPriceCache    = make(map[string]float64)
	ethPriceCacheMu  sync.Mutex
	ethPriceInitOnce sync.Once
)

func updateEthPrice(chainName string) {
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=DIGITAL_CURRENCY_DAILY&symbol=%s&market=USD&apikey=E1UXVNPVNF795OR3", ChainMainToken[chainName])
	resp := reqC.Get(url).Do()

	if resp.Err != nil {
		log.Errorf("get eth price failed, err:%v", resp.Err)
		return
	}
	if resp.IsErrorState() {
		log.Errorf("get eth price failed, status code:%d", resp.GetStatusCode())
		return
	}

	s, err := resp.ToString()
	if err != nil {
		log.Errorf("get eth price failed, err:%v", err)
		return
	}
	gjson.Get(s, "Time Series (Digital Currency Daily)").ForEach(func(key, value gjson.Result) bool {
		date := key.String()
		vm := value.Map()
		price := vm["4a. close (USD)"].Float()
		ethPriceCacheMu.Lock()
		ethPriceCache[date] = price
		ethPriceCacheMu.Unlock()
		return true
	})
}

func GetMainTokenPriceInDate(chainName string, ts int64) float64 {
	ethPriceInitOnce.Do(func() {
		updateEthPrice(chainName)
		go func() {
			for {
				time.Sleep(time.Hour * 12)
				updateEthPrice(chainName)
			}
		}()
	})

	t := time.Unix(ts, 0)
	date := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	ethPriceCacheMu.Lock()
	defer ethPriceCacheMu.Unlock()
	if price, ok := ethPriceCache[date.Format("2006-01-02")]; ok {
		return price
	}
	return 0
}

func CalcGasFee(chainName string, gasPrice, gasUsed int64) float64 {
	b := &big.Int{}
	gb := big.NewInt(gasPrice)
	ub := big.NewInt(gasUsed)
	total := b.Mul(gb, ub)
	dl := math.Pow(10, float64(MainTokenInfo[chainName].Decimal))
	totalDf := decimal.NewFromInt(total.Int64()).Div(decimal.NewFromFloat(dl))
	totalGasFee, _ := totalDf.Float64()
	return totalGasFee
}

func CheckChainName(chainName string) bool {
	support := []string{"eth", "bsc"}
	return ArrayContains(support, chainName)
}
