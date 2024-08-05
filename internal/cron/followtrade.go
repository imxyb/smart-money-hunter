package cron

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/exp/rand"
	"gorm.io/gorm"
	"smart-money/internal/exchange"
	inch "smart-money/pkg/1inch"
	"smart-money/pkg/eth"
	"smart-money/pkg/log"
	"smart-money/pkg/model"
	"smart-money/pkg/oklink"
	"smart-money/pkg/util"
)

var (
	followAddressTradeJobLock sync.Mutex
)

func FollowAddressTradeBuyJob() {
	if !followAddressTradeJobLock.TryLock() {
		return
	}
	defer followAddressTradeJobLock.Unlock()

	db := model.GetDB()

	var followAddresses []*model.FollowAddress
	err := db.Where("status = ?", model.FollowAddressStatusNormal).Find(&followAddresses).Error
	if err != nil {
		log.Errorf("FollowAddressTradeBuyJob: get follow followAddress error: %v", err)
		return
	}

	var wallets []*model.MyWallet
	err = db.Where("status = ?", model.MyWalletStatusEnable).Find(&wallets).Error
	if err != nil {
		log.Errorf("FollowAddressTradeBuyJob: get my wallet error: %v", err)
		return
	}
	if len(wallets) == 0 {
		log.Errorf("FollowAddressTradeBuyJob: no wallet available")
		return
	}

	for _, followAddress := range followAddresses {
		if err = dealFollowAddress(wallets, followAddress); err != nil {
			log.Errorf("FollowAddressTradeBuyJob: deal follow address error: %v", err)
			continue
		}
	}
}

func dealFollowAddress(wallets []*model.MyWallet, followAddress *model.FollowAddress) error {
	// 查找最新一条
	latestTx, err := oklink.Api.GetToken20TransactionListByAddress(followAddress.ChainName, followAddress.Address, 1, 1)
	if err != nil {
		return fmt.Errorf("FollowAddressTradeBuyJob: get lastest tx error: %v", err)
	}
	if len(latestTx.Data) == 0 || len(latestTx.Data[0].TransactionLists) == 0 {
		return nil
	}

	latestTxTxHash := latestTx.Data[0].TransactionLists[0].TxId
	log.Infof("FollowAddressTradeBuyJob: aderss: %v latest tx: %v", followAddress.Address, latestTxTxHash)

	if latestTxTxHash != followAddress.LastErc20TxHash {
		log.Infof("FollowAddressTradeBuyJob: get new tx: %v", latestTxTxHash)

		detailResp, err := oklink.Api.GetTransactionDetail(followAddress.ChainName, latestTxTxHash)
		if err != nil {
			return fmt.Errorf("FollowAddressTradeBuyJob: get tx detail error: %v", err)
		}
		if len(detailResp.Data) == 0 {
			return nil
		}

		detail := detailResp.Data[0]
		sellToken := detail.TokenTransferDetails[0]
		buyToken := detail.TokenTransferDetails[len(detail.TokenTransferDetails)-1]
		if buyToken.TokenId != "" && sellToken.TokenId != "" {
			return nil
		}

		// 只记录买入
		if !util.IsMainToken(buyToken.Symbol) && !util.IsMainToken(sellToken.Symbol) {
			log.Debugf("FollowAddressTradeBuyJob: not main token: %v", buyToken.Symbol)
			return nil
		}
		if util.IsMainToken(buyToken.Symbol) && util.IsMainToken(sellToken.Symbol) {
			log.Infof("FollowAddressTradeBuyJob: main token: %v", buyToken.Symbol)
			return nil
		}
		if util.IsMainToken(buyToken.Symbol) && !util.IsMainToken(sellToken.Symbol) {
			log.Infof("FollowAddressTradeBuyJob: buy main token: %v", buyToken.Symbol)
			return nil
		}

		log.Infof("FollowAddressTradeBuyJob: address:%v buy token: %v, sell token: %v", followAddress.Address, buyToken, sellToken)

		followAddress.LastErc20TxHash = latestTxTxHash
		latestTxTime, err := strconv.Atoi(latestTx.Data[0].TransactionLists[0].TransactionTime)
		if err != nil {
			return fmt.Errorf("FollowAddressTradeBuyJob: get lastest tx time error: %v", err)
		}
		followAddress.LastErc20TxTime = int64(latestTxTime)
		if err = model.SaveFollowAddress(followAddress); err != nil {
			return fmt.Errorf("FollowAddressTradeBuyJob: save follow address error: %v", err)
		}

		tmpFollowTrade := new(model.FollowTrade)
		err = model.GetDB().Where("buy_token_address = ? and status != ?", buyToken.TokenContractAddress, model.FollowTradeStatusFinish).First(tmpFollowTrade).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("FollowAddressTradeBuyJob: get follow trade error: %v", err)
		}
		if tmpFollowTrade.ID > 0 {
			log.Infof("FollowAddressTradeBuyJob: follow trade buy token:%v already exist", buyToken.TokenContractAddress)
			return nil
		}

		followTrade := new(model.FollowTrade)
		defer func() {
			if err := model.CreateFollowTrade(followTrade); err != nil {
				return
			}
		}()

		followTrade.ChainName = followAddress.ChainName
		followTrade.FollowAddress = followAddress.Address
		followTrade.FollowAddressBuyTxHash = latestTxTxHash
		followTrade.FollowAddressBuyTime = int64(latestTxTime)
		followTrade.BuyTokenAddress = buyToken.TokenContractAddress
		followTrade.BuySymbol = buyToken.Symbol
		followTrade.Status = model.FollowTradeStatusSuccess

		buyAmountDf, err := decimal.NewFromString(buyToken.Amount)
		if err != nil {
			followTrade.Status = model.FollowTradeStatusFail
			followTrade.FailReason = fmt.Sprintf("buy amount parse error: %v", err)
			return err
		}
		followTrade.FollowAddressBuyAmount, _ = buyAmountDf.Float64()

		var wallet *model.MyWallet
		if len(wallets) > 1 {
			wallet = wallets[rand.Int63n(int64(len(wallets)-1))]
		} else {
			wallet = wallets[0]
		}

		followTrade.WalletAddress = wallet.Address
		followTrade.WalletAddressSellAmount = wallet.EachSellAmount

		mainTokenDecimal := int64(util.MainTokenInfo[followAddress.ChainName].Decimal)
		mainTokenAddress := util.MainTokenInfo[followAddress.ChainName].ContractAddress
		walletSellMainTokenAmountDf := decimal.NewFromFloat(wallet.EachSellAmount).Mul(decimal.NewFromFloat(math.Pow(10, float64(mainTokenDecimal))))
		quote, err := inch.Quote(followAddress.ChainName, util.MainTokenInfo[followAddress.ChainName].ContractAddress,
			followTrade.BuyTokenAddress, walletSellMainTokenAmountDf.BigInt().Int64())
		if err != nil {
			followTrade.Status = model.FollowTradeStatusFail
			followTrade.FailReason = fmt.Errorf("FollowAddressTradeBuyJob: get quote error: %v", err).Error()
			return fmt.Errorf("FollowAddressTradeBuyJob: get quote error: %v", err)
		}
		toTokenAmountDf, err := decimal.NewFromString(quote.ToTokenAmount)
		if err != nil {
			followTrade.Status = model.FollowTradeStatusFail
			followTrade.FailReason = fmt.Errorf("FollowAddressTradeBuyJob: to token amount error: %v", err).Error()
			return err
		}
		toTokenRealAmountDf := toTokenAmountDf.Div(decimal.NewFromFloat(math.Pow(10, float64(quote.ToToken.Decimals))))
		followTrade.WalletAddressBuyAmount, _ = toTokenRealAmountDf.Float64()
		followTrade.BuyTokenDecimal = quote.ToToken.Decimals

		swapRequest := &inch.SwapRequest{
			FromTokenAddress: mainTokenAddress,
			ToTokenAddress:   followTrade.BuyTokenAddress,
			Amount:           walletSellMainTokenAmountDf.String(),
			FromAddress:      wallet.Address,
			Slippage:         20,
		}
		ex := exchange.NewExchange(followTrade.ChainName, wallet.PrivateKey, swapRequest)

		// 检查是否授权
		allowance, err := ex.CheckAllowance()
		if err != nil {
			followTrade.Status = model.FollowTradeStatusFail
			followTrade.FailReason = fmt.Errorf("FollowAddressTradeBuyJob: check allowance error: %v", err).Error()
			return err
		}
		if allowance == "0" {
			approveTx, err := ex.ApproveTransaction(true)
			if err != nil {
				followTrade.Status = model.FollowTradeStatusFail
				followTrade.FailReason = fmt.Errorf("FollowAddressTradeBuyJob: approve tx error: %v", err).Error()
				return err
			}
			_, err = eth.Client.WaitTxHashReceipt(approveTx)
			if err != nil {
				followTrade.Status = model.FollowTradeStatusFail
				followTrade.FailReason = fmt.Errorf("FollowAddressTradeBuyJob: wait approve tx receipt error: %v", err).Error()
				return err
			}
		}

		swapTx, err := ex.Swap()
		if err != nil {
			followTrade.Status = model.FollowTradeStatusFail
			followTrade.FailReason = fmt.Errorf("FollowAddressTradeBuyJob: swap error: %v", err).Error()
			return err
		}
		followTrade.WalletAddressBuyTxHash = swapTx.String()
		followTrade.WalletAddressBuyTime = time.Now().Unix()

		receipt, err := eth.Client.WaitTxHashReceipt(swapTx)
		if err != nil {
			followTrade.Status = model.FollowTradeStatusFail
			followTrade.FailReason = fmt.Errorf("FollowAddressTradeBuyJob: wait swap tx receipt error: %v", err).Error()
			return err
		}
		followTrade.WalletAddressBuyGas = util.CalcGasFee(followAddress.ChainName, receipt.EffectiveGasPrice.Int64(), int64(receipt.GasUsed))

		return nil
	}

	return nil
}

func SellPrincipalJob() {
	var followTrades []*model.FollowTrade
	db := model.GetDB()

	err := db.Where("status = ? and is_sell_principal=?", model.FollowTradeStatusSuccess, 0).Find(&followTrades).Error
	if err != nil {
		log.Errorf("SellPrincipalJob: get follow trades error: %v", err)
		return
	}

	for _, followTrade := range followTrades {
		dl := decimal.NewFromFloat(math.Pow(10, float64(followTrade.BuyTokenDecimal)))
		amountDf := decimal.NewFromFloat(followTrade.WalletAddressBuyAmount).Mul(dl)
		amount := amountDf.BigInt()
		quote, err := inch.Quote(followTrade.ChainName, followTrade.BuyTokenAddress, util.MainTokenInfo[followTrade.ChainName].ContractAddress, amount.Int64())
		if err != nil {
			log.Errorf("SellPrincipalJob: get quote error: %v", err)
			continue
		}
		toTokenAmountDf, err := decimal.NewFromString(quote.ToTokenAmount)
		toTokenRealAmountDf := toTokenAmountDf.Div(decimal.NewFromFloat(math.Pow(10, float64(quote.ToToken.Decimals))))
		cost := decimal.NewFromFloat(followTrade.WalletAddressSellAmount).Add(decimal.NewFromFloat(followTrade.WalletAddressBuyGas))
		gasPrice, err := eth.Client.GetEthClient().SuggestGasPrice(context.Background())
		if err != nil {
			log.Errorf("SellPrincipalJob: get gas price error: %v", err)
			continue
		}
		gasPriceDf := decimal.NewFromInt(gasPrice.Int64()).Div(decimal.NewFromFloat(math.Pow(10, float64(util.MainTokenInfo[followTrade.ChainName].Decimal))))
		sellGasUsed := gasPriceDf.Mul(decimal.NewFromInt(int64(quote.EstimatedGas)))

		buyCost := cost.Add(sellGasUsed)
		shouldSellPrincipalMainTokenAmount := buyCost.Mul(decimal.NewFromInt(2))
		log.Debugf("当前总价值：%v, 成本: %v, 当前卖出需要的gas fee:%v",
			toTokenRealAmountDf.String(), cost.String(), sellGasUsed.String())
		log.Debugf("翻本卖出达到的数量应该是:%v", shouldSellPrincipalMainTokenAmount.String())

		toTokenRealAmountDf = decimal.NewFromFloat(0.013)

		// 翻倍
		if toTokenRealAmountDf.GreaterThanOrEqual(shouldSellPrincipalMainTokenAmount) {
			log.Infof("已翻倍，启动出本策略")
			bf := buyCost.Mul(decimal.NewFromFloat(math.Pow(10, float64(util.MainTokenInfo[followTrade.ChainName].Decimal))))
			quote, err = inch.Quote(followTrade.ChainName, util.MainTokenInfo[followTrade.ChainName].ContractAddress, followTrade.BuyTokenAddress, bf.BigInt().Int64())
			if err != nil {
				log.Errorf("SellPrincipalJob: get quote error: %v", err)
				continue
			}
			log.Infof("卖出代币数量：%v", quote.ToTokenAmount)
			wallet := new(model.MyWallet)
			err = db.Where("chain_name = ? and address = ? and status = ?", followTrade.ChainName, followTrade.WalletAddress, model.MyWalletStatusEnable).First(wallet).Error
			if err != nil {
				log.Errorf("SellPrincipalJob: get wallet error: %v", err)
				continue
			}
			swapRequest := &inch.SwapRequest{
				FromTokenAddress: followTrade.BuyTokenAddress,
				ToTokenAddress:   util.MainTokenInfo[followTrade.ChainName].ContractAddress,
				Amount:           quote.ToTokenAmount,
				FromAddress:      followTrade.WalletAddress,
				Slippage:         20,
			}
			ex := exchange.NewExchange(followTrade.ChainName, wallet.PrivateKey, swapRequest)
			// 检查是否授权
			allowance, err := ex.CheckAllowance()
			if err != nil {
				log.Errorf("SellPrincipalJob: check allowance error: %v", err)
				continue
			}
			af, _ := decimal.NewFromString(allowance)
			qf, _ := decimal.NewFromString(quote.ToTokenAmount)
			if allowance == "0" || af.LessThanOrEqual(qf) {
				log.Infof("授权代币数量：%v", quote.ToTokenAmount)
				approveTx, err := ex.ApproveTransaction(true)
				if err != nil {
					log.Errorf("SellPrincipalJob: approve tx error: %v", err)
					continue
				}
				_, err = eth.Client.WaitTxHashReceipt(approveTx)
				if err != nil {
					log.Errorf("SellPrincipalJob: wait approve tx receipt error: %v", err)
					continue
				}
			}

			swapTx, err := ex.Swap()
			if err != nil {
				log.Errorf("SellPrincipalJob: swap error: %v", err)
				continue
			}
			if _, err := eth.Client.WaitTxHashReceipt(swapTx); err != nil {
				log.Errorf("SellPrincipalJob: wait swap tx receipt error: %v", err)
				continue
			}
			followTrade.SellPrincipalTxHash = swapTx.String()
			followTrade.IsSellPrincipal = 1
			if err = model.SaveFollowTrade(followTrade); err != nil {
				log.Errorf("SellPrincipalJob: save follow trade error: %v", err)
				continue
			}
			log.Infof("出本成功，交易hash：%v", swapTx.String())
		}
	}
}

func CheckBalanceJob() {
	var followTrades []*model.FollowTrade
	db := model.GetDB()

	err := db.Where("status=?", model.FollowTradeStatusSuccess).Find(&followTrades).Error
	if err != nil {
		log.Errorf("SellPrincipalJob: get follow trades error: %v", err)
		return
	}

	for _, followTrade := range followTrades {
		balanceResp, err := oklink.Api.GetAddressBalance(followTrade.ChainName, followTrade.WalletAddress, followTrade.BuyTokenAddress, 1, 1)
		if err != nil {
			log.Errorf("SellPrincipalJob: get balance error: %v", err)
			continue
		}

		// 计算余额
		if len(balanceResp.Data[0].TokenList) > 0 {
			tokenDecimal, err := eth.Client.GetTokenDecimals(followTrade.BuyTokenAddress)
			if err != nil {
				log.Errorf("SellPrincipalJob: get token decimal error: %v", err)
				continue
			}
			holdingAmountDf, err := decimal.NewFromString(balanceResp.Data[0].TokenList[0].HoldingAmount)
			if err != nil {
				log.Errorf("SellPrincipalJob: get holding amount error: %v", err)
				continue
			}
			df := holdingAmountDf.Mul(decimal.NewFromFloat(math.Pow(10, float64(tokenDecimal))))
			amount := df.BigInt().Int64()
			if amount < 0 {
				amount = 0
			}
			quote, err := inch.Quote(followTrade.ChainName, followTrade.BuyTokenAddress, util.USDTContractMap[followTrade.ChainName].ContractAddress, amount)
			if err != nil {
				log.Errorf("SellPrincipalJob: get quote error: %v", err)
				continue
			}

			var amountDf decimal.Decimal
			if err != nil {
				amountDf = decimal.NewFromInt(0)
			} else {
				amountDf, err = decimal.NewFromString(quote.ToTokenAmount)
				if err != nil {
					log.Errorf("SellPrincipalJob: get amount decimal error: %v", err)
					continue
				}
			}
			usdtValueDf := amountDf.Div(decimal.NewFromFloat(math.Pow(10, float64(util.USDTContractMap[followTrade.ChainName].Decimal))))
			// 小于10u就认为卖完了
			if usdtValueDf.LessThanOrEqual(decimal.NewFromFloat(10)) {
				log.Infof("follow trade id:%d 购买的代币余额(usdt价值)：%v, 小于10u，当做已结束", followTrade.ID, usdtValueDf)
				followTrade.Status = model.FollowTradeStatusFinish
				if err = model.SaveFollowTrade(followTrade); err != nil {
					log.Errorf("SellPrincipalJob: save follow trade error: %v", err)
				}
				continue
			}
		}
	}
}
