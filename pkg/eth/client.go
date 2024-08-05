package eth

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	inch "smart-money/pkg/1inch"
	"smart-money/pkg/eth/erc20"
	"smart-money/pkg/log"
)

var (
	Client *client
)

type client struct {
	ethClient *ethclient.Client
	chainID   int64
}

func InitClient(rpcAddr string, chainID int64) error {
	var err error
	c, err := ethclient.Dial(rpcAddr)
	if err != nil {
		return err
	}
	Client = &client{
		ethClient: c,
		chainID:   chainID,
	}
	return nil
}

func (c *client) GetEthClient() *ethclient.Client {
	return c.ethClient
}

func (c *client) GetTokenDecimals(tokenAddress string) (uint8, error) {
	contract, err := erc20.NewErc20(common.HexToAddress(tokenAddress), c.ethClient)
	if err != nil {
		return 0, err
	}
	return contract.Decimals(nil)
}

func (c *client) Approve(opts *bind.TransactOpts, tokenAddress, spender string, amount *big.Int) (common.Hash, error) {
	contract, err := erc20.NewErc20(common.HexToAddress(tokenAddress), c.ethClient)
	if err != nil {
		return [32]byte{}, err
	}
	tx, err := contract.Approve(opts, common.HexToAddress(spender), amount)
	if err != nil {
		return [32]byte{}, err
	}
	return tx.Hash(), nil
}

func (c *client) SendTransaction(privateKey string, txData *inch.TxData) (common.Hash, error) {
	// 从账户发送交易
	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return [32]byte{}, err
	}

	fromAddress := common.HexToAddress(txData.From)
	toAddress := common.HexToAddress(txData.To)

	nonce, err := c.ethClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return [32]byte{}, err
	}

	val, err := strconv.Atoi(txData.Value)
	if err != nil {
		return [32]byte{}, err
	}

	gasLimit := uint64(1000000)
	if txData.Gas != 0 {
		gasLimit = uint64(txData.Gas)
	}

	gp, err := strconv.Atoi(txData.GasPrice)
	if err != nil {
		return [32]byte{}, err
	}
	gasPrice := big.NewInt(int64(gp))

	rawTx := &types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &toAddress,
		Value:    big.NewInt(int64(val)),
	}
	data, err := hex.DecodeString(strings.ReplaceAll(txData.Data, "0x", ""))
	if err != nil {
		return [32]byte{}, err
	}
	rawTx.Data = data

	signedTx, err := types.SignNewTx(pk, types.NewEIP155Signer(big.NewInt(c.chainID)), rawTx)
	if err != nil {
		return [32]byte{}, err
	}

	if err = c.ethClient.SendTransaction(context.Background(), signedTx); err != nil {
		return [32]byte{}, err
	}

	hash := signedTx.Hash()
	return hash, nil
}

func (c *client) GetAuth(pk, fromAddress string) (*bind.TransactOpts, error) {
	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		return nil, err
	}

	nonce, err := c.ethClient.PendingNonceAt(context.Background(), common.HexToAddress(fromAddress))
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(c.chainID))
	if err != nil {
		log.Errorf("NewKeyedTransactorWithChainID err:%s", err)
		return nil, err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = uint64(0)  // in units

	gasPrice, err := c.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	auth.GasPrice = gasPrice
	return auth, nil
}

func (c *client) WaitTxHashReceipt(txHash common.Hash) (*types.Receipt, error) {
	tryTimes := 30
	for tryTimes > 0 {
		receipt, err := c.ethClient.TransactionReceipt(context.Background(), txHash)
		if err == ethereum.NotFound {
			tryTimes -= 1
			time.Sleep(3 * time.Second)
			continue
		}
		if err != nil {
			return nil, err
		}
		if receipt.Status == types.ReceiptStatusSuccessful {
			return receipt, nil
		}
		if receipt.Status == types.ReceiptStatusFailed {
			return nil, fmt.Errorf("tx status failed")
		}
		tryTimes -= 1
		time.Sleep(3 * time.Second)
	}

	return nil, fmt.Errorf("wait tx receipt timeout")
}
