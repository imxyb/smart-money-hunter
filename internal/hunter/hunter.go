package hunter

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	ccllector "smart-money/internal/collector"
	inch "smart-money/pkg/1inch"
	"smart-money/pkg/eth"
	"smart-money/pkg/log"
	"smart-money/pkg/model"
	"smart-money/pkg/oklink"
	"smart-money/pkg/util"
)

var (
	EOF          = fmt.Errorf("EOF")
	ErrNotInDate = fmt.Errorf("not in date range")
)

type Hunter struct {
	chainName       string
	taskName        string
	db              *gorm.DB
	collector       ccllector.Collector
	collectorParams map[string]any
	collectDuration int64
}

func NewHunter(chainName, taskName string, collector ccllector.Collector, collectorParams map[string]any, collectorSeconds int64) *Hunter {
	return &Hunter{
		chainName:       chainName,
		taskName:        taskName,
		collector:       collector,
		db:              model.GetDB(),
		collectorParams: collectorParams,
		collectDuration: collectorSeconds,
	}
}

func (h *Hunter) Work() error {
	addresses, err := h.Collect()
	if err != nil {
		log.Errorf("collect %s %s error: %s", h.chainName, h.taskName, err.Error())
		return err
	}

	if err := h.Analyze(addresses); err != nil {
		log.Errorf("analyze %s %s error: %s", h.chainName, h.taskName, err.Error())
		return err
	}
	log.Infof("analyze %s %s success", h.chainName, h.taskName)
	return nil
}

func (h *Hunter) Collect() ([]string, error) {
	addresses, err := h.collector.Collect(h.chainName, h.collectorParams)
	if err != nil {
		return nil, err
	}

	for _, address := range addresses {
		tokens, err := h.getBuyTokens(address)
		if err != nil {
			return nil, err
		}

		for _, tokenAddress := range tokens {
			if err = h.recordTokenTransactionOfAddress(address, tokenAddress); err != nil {
				return nil, err
			}
		}

	}
	return addresses, nil
}

func (h *Hunter) getBuyTokens(address string) ([]string, error) {
	now := time.Now()
	cacheInvalid := make(map[string]bool)
	loop := func(tokens map[string]bool, page int) error {
		tlResp, err := oklink.Api.GetToken20TransactionListByAddress(h.chainName, address, page, 20)
		if err != nil {
			return err
		}

		if len(tlResp.Data[0].TransactionLists) == 0 {
			return EOF
		}

		if err != nil {
			return err
		}
		for _, data := range tlResp.Data {
			for _, tx := range data.TransactionLists {
				if tx.State != "success" {
					continue
				}

				if _, exist := cacheInvalid[tx.TransactionSymbol]; exist {
					continue
				}

				txTime, _ := strconv.Atoi(tx.TransactionTime)
				// 只拿一个月内的记录
				if now.Unix()-int64(txTime) > h.collectDuration {
					return EOF
				}
				if util.IsMainToken(tx.TransactionSymbol) {
					continue
				}
				detail, err := oklink.Api.GetTransactionDetail(h.chainName, tx.TxId)
				if err != nil {
					return err
				}
				// 检查是否还有余额，有的话说明没卖完，不计算在交易内
				txDetail := detail.Data[0]
				tokenAddress := txDetail.TokenTransferDetails[0].TokenContractAddress
				balanceResp, err := oklink.Api.GetAddressBalance(h.chainName, address, tokenAddress, 1, 1)
				if err != nil {
					return err
				}

				// 计算余额
				if len(balanceResp.Data[0].TokenList) > 0 {
					tokenDecimal, err := eth.Client.GetTokenDecimals(tokenAddress)
					if err != nil {
						return err
					}
					holdingAmountDf, err := decimal.NewFromString(balanceResp.Data[0].TokenList[0].HoldingAmount)
					if err != nil {
						return err
					}
					df := holdingAmountDf.Mul(decimal.NewFromFloat(math.Pow(10, float64(tokenDecimal))))
					amount := df.BigInt().Int64()
					if amount < 0 {
						continue
					}

					// 计算eth余额
					quote, err := inch.Quote(h.chainName, tokenAddress, util.MainTokenInfo[h.chainName].ContractAddress, amount)
					if err != nil {
						log.Errorf("quote error: %s", err.Error())
						continue
					}

					var amountDf decimal.Decimal
					if err != nil {
						amountDf = decimal.NewFromInt(0)
					} else {
						amountDf, err = decimal.NewFromString(quote.ToTokenAmount)
						if err != nil {
							log.Errorf("amountDf error: %s", err.Error())
							continue
						}
					}
					todayPrice := decimal.NewFromFloat(util.GetMainTokenPriceInDate(h.chainName, time.Now().Unix()))
					usdtValueDf := amountDf.Div(decimal.NewFromFloat(math.Pow(10, float64(util.MainTokenInfo[h.chainName].Decimal)))).Mul(todayPrice)
					// 大于30u的都算有余额，没卖完
					if usdtValueDf.GreaterThan(decimal.NewFromFloat(30)) {
						cacheInvalid[tx.TransactionSymbol] = true
						log.Infof("address %s token %s has balance %s, ignore", address, tx.TransactionSymbol, balanceResp.Data[0].TokenList[0].ValueUsd)
						continue
					}
				}
				tokens[tokenAddress] = true
			}
		}
		return nil
	}

	page := 1
	tokens := make(map[string]bool)
	var err error
	for {
		err = loop(tokens, page)
		if err == EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		page += 1
		time.Sleep(time.Second)
	}

	var tokenList []string
	for token := range tokens {
		tokenList = append(tokenList, token)
	}
	return tokenList, nil
}

func (h *Hunter) recordTokenTransactionOfAddress(address, tokenAddress string) error {
	now := time.Now()
	loop := func(page int) ([]*model.TokenTransactionCollect, error) {
		tlResp, err := oklink.Api.GetToken20TransactionListByAddressAndToken(h.chainName, address, tokenAddress, page, 50)
		if err != nil {
			return nil, err
		}

		if len(tlResp.Data[0].TransactionLists) == 0 {
			return nil, EOF
		}

		var tts []*model.TokenTransactionCollect

		for _, data := range tlResp.Data {
		loop:
			for _, tx := range data.TransactionLists {
				if tx.State != "success" {
					continue
				}

				var count int64
				err = h.db.Model(&model.TokenTransactionCollect{}).Where("tx_hash = ?", tx.TxId).Count(&count).Error
				if err != nil {
					log.Errorf("get tx count error: %s", err.Error())
					continue
				}
				if count > 0 {
					log.Info("tx already exist")
					continue
				}

				txTime, _ := strconv.Atoi(tx.TransactionTime)
				if now.Unix()-int64(txTime) > h.collectDuration {
					return nil, EOF
				}
				if time.Now().Unix()-int64(txTime) < 0 {
					return nil, ErrNotInDate
				}
				detailResp, err := oklink.Api.GetTransactionDetail(h.chainName, tx.TxId)
				if err != nil {
					return nil, err
				}
				for _, detail := range detailResp.Data {
					var (
						buyToken  *oklink.TokenTransferDetail
						sellToken *oklink.TokenTransferDetail
						buyTxFrom string
					)

					for _, transferDetail := range detail.TokenTransferDetails {
						if buyToken == nil && strings.EqualFold(transferDetail.To, address) {
							buyToken = transferDetail
							buyTxFrom = transferDetail.From
						}
						if buyTxFrom != "" && strings.EqualFold(transferDetail.To, buyTxFrom) && sellToken == nil {
							sellToken = transferDetail
						}
					}

					if buyToken == nil || sellToken == nil {
						log.Warnf("buyToken or sellToken is nil, tx: %s", tx.TxId)
						continue loop
					}

					// 排除nft
					if buyToken.TokenId != "" && sellToken.TokenId != "" {
						continue loop
					}

					// 非主流币对非主流币交易，过滤
					if !util.IsMainToken(buyToken.Symbol) && !util.IsMainToken(sellToken.Symbol) {
						continue loop
					}
					if util.IsMainToken(buyToken.Symbol) && util.IsMainToken(sellToken.Symbol) {
						if buyToken.Symbol != sellToken.Symbol {
							continue loop
						}
						// 如果是buy和sell都是同一个主流币，考虑是log的问题，取倒数第二个index作为buytoken
						if len(detail.TokenTransferDetails) > 2 {
							buyToken = detail.TokenTransferDetails[len(detail.TokenTransferDetails)-2]
						}
					}

					tt := &model.TokenTransactionCollect{
						ChainName:   h.chainName,
						TaskName:    h.taskName,
						Address:     address,
						TxHash:      tx.TxId,
						BuyAddress:  buyToken.TokenContractAddress,
						SellAddress: sellToken.TokenContractAddress,
						BuySymbol:   buyToken.Symbol,
						SellSymbol:  sellToken.Symbol,
					}
					txTime, _ := strconv.Atoi(tx.TransactionTime)
					tt.TxTime = uint64(txTime)

					blockHeight, _ := strconv.Atoi(tx.Height)
					tt.BlockHeight = int64(blockHeight)

					df, err := decimal.NewFromString(buyToken.Amount)
					if err != nil {
						return nil, err
					}
					tt.BuyAmount, _ = df.Float64()

					df, err = decimal.NewFromString(sellToken.Amount)
					if err != nil {
						return nil, err
					}
					tt.SellAmount, _ = df.Float64()

					tts = append(tts, tt)
				}
			}
		}
		return tts, nil
	}

	page := 1
	var models [][]*model.TokenTransactionCollect
	for {
		tts, err := loop(page)
		if err == ErrNotInDate {
			return nil
		}
		if err == EOF {
			break
		}
		if err != nil {
			return err
		}
		models = append(models, tts)
		page += 1
		time.Sleep(time.Second)
	}

	for _, tts := range models {
		for _, tt := range tts {
			if err := model.CreateTokenTransactionCollect(tt); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *Hunter) Analyze(addresses []string) error {
	for _, address := range addresses {
		var buyErcTxs []model.TokenTransactionCollect
		err := h.db.Distinct("buy_symbol").Where("task_name=? and address=? and buy_symbol not in ?", h.taskName, address, util.MainTokens).Find(&buyErcTxs).Error
		if err != nil {
			return err
		}

		var buyErcTokens []string
		for _, buyTx := range buyErcTxs {
			buyErcTokens = append(buyErcTokens, buyTx.BuySymbol)
		}

		var addressTrades []*model.AddressTrade
		for _, token := range buyErcTokens {
			addressTrade, err := h.makeAddressTrader(address, token)
			if err != nil {
				log.Errorf("makeAddressTrader err: %v", err)
				continue
			}
			addressTrades = append(addressTrades, addressTrade)
		}

		for _, addressTrade := range addressTrades {
			if err = model.CreateAddressTrade(addressTrade); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *Hunter) makeAddressTrader(address, token string) (*model.AddressTrade, error) {
	var (
		err     error
		buyTxs  []*model.TokenTransactionCollect
		sellTxs []*model.TokenTransactionCollect
	)

	err = h.db.Where("task_name=? and address=? and buy_symbol=?", h.taskName, address, token).Order("block_height asc").Find(&buyTxs).Error
	if err != nil {
		return nil, err
	}

	err = h.db.Where("task_name=? and address=? and sell_symbol=?", h.taskName, address, token).Order("block_height asc").Find(&sellTxs).Error
	if err != nil {
		return nil, err
	}

	if len(buyTxs) == 0 || len(sellTxs) == 0 {
		return nil, fmt.Errorf("address: %s, token: %s, buyTxs: %d, sellTxs: %d, ignore", address, token, len(buyTxs), len(sellTxs))
	}

	addressTrade := &model.AddressTrade{
		Address:     address,
		BuySymbol:   buyTxs[0].BuySymbol,
		SellSymbol:  sellTxs[len(sellTxs)-1].SellSymbol,
		BuyAddress:  buyTxs[0].BuyAddress,
		SellAddress: sellTxs[len(sellTxs)-1].SellAddress,
		FirstTxTime: buyTxs[0].TxTime,
		LastTxTime:  sellTxs[len(sellTxs)-1].TxTime,
		ChainName:   buyTxs[0].ChainName,
	}

	var (
		buyTotalUsd  decimal.Decimal
		sellTotalUsd decimal.Decimal

		filterSameBuyTx  = make(map[string]bool)
		filterSameSellTx = make(map[string]bool)
	)

	for _, tx := range buyTxs {
		if _, ok := filterSameBuyTx[tx.TxHash]; ok {
			continue
		}

		if !util.IsMainToken(tx.SellSymbol) {
			return nil, fmt.Errorf("not main token")
		}
		if util.IsStableToken(tx.SellSymbol) {
			buyTotalUsd = buyTotalUsd.Add(decimal.NewFromFloat(tx.SellAmount))
		} else {
			// eth
			ethPrice := util.GetMainTokenPriceInDate(h.chainName, int64(tx.TxTime))
			if ethPrice == 0 {
				return nil, fmt.Errorf("eth price is 0 in %d, txid:%s, ignore", tx.TxTime, tx.TxHash)
			}
			buyTotalUsd = buyTotalUsd.Add(decimal.NewFromFloat(tx.SellAmount).Mul(decimal.NewFromFloat(ethPrice)))
		}

		filterSameBuyTx[tx.TxHash] = true
	}
	addressTrade.BuyTotalUsd, _ = buyTotalUsd.Float64()

	for _, tx := range sellTxs {
		if _, ok := filterSameSellTx[tx.TxHash]; ok {
			continue
		}
		if !util.IsMainToken(tx.BuySymbol) {
			return nil, fmt.Errorf("not main token")
		}
		if util.IsStableToken(tx.BuySymbol) {
			sellTotalUsd = sellTotalUsd.Add(decimal.NewFromFloat(tx.BuyAmount))
		} else {
			// eth
			ethPrice := util.GetMainTokenPriceInDate(h.chainName, int64(tx.TxTime))
			if ethPrice == 0 {
				return nil, fmt.Errorf("eth price is 0 in %d, txid:%s, ignore", tx.TxTime, tx.TxHash)
			}
			sellTotalUsd = sellTotalUsd.Add(decimal.NewFromFloat(tx.BuyAmount).Mul(decimal.NewFromFloat(ethPrice)))
		}
		filterSameSellTx[tx.TxHash] = true
	}

	addressTrade.SellTotalUsd, _ = sellTotalUsd.Float64()
	addressTrade.Profit = addressTrade.SellTotalUsd - addressTrade.BuyTotalUsd

	return addressTrade, nil
}
