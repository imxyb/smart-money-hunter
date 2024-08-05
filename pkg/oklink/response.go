package oklink

type TransactionListResp struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		Page             string `json:"page"`
		Limit            string `json:"limit"`
		TotalPage        string `json:"totalPage"`
		ChainFullName    string `json:"chainFullName"`
		ChainShortName   string `json:"chainShortName"`
		TransactionLists []struct {
			TxId              string `json:"txId"`
			BlockHash         string `json:"blockHash"`
			Height            string `json:"height"`
			TransactionTime   string `json:"transactionTime"`
			From              string `json:"from"`
			To                string `json:"to"`
			Amount            string `json:"amount"`
			TransactionSymbol string `json:"transactionSymbol"`
			TxFee             string `json:"txFee"`
			State             string `json:"state"`
		} `json:"transactionLists"`
	} `json:"data"`
}

type TokenTransferDetail struct {
	Index                string `json:"index"`
	Token                string `json:"token"`
	TokenContractAddress string `json:"tokenContractAddress"`
	Symbol               string `json:"symbol"`
	From                 string `json:"from"`
	To                   string `json:"to"`
	TokenId              string `json:"tokenId"`
	Amount               string `json:"amount"`
}

type TransactionDetailResp struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		ChainFullName     string `json:"chainFullName"`
		ChainShortName    string `json:"chainShortName"`
		Txid              string `json:"txid"`
		Height            string `json:"height"`
		TransactionTime   string `json:"transactionTime"`
		Amount            string `json:"amount"`
		TransactionSymbol string `json:"transactionSymbol"`
		Txfee             string `json:"txfee"`
		Index             string `json:"index"`
		Confirm           string `json:"confirm"`
		InputDetails      []struct {
			InputHash string `json:"inputHash"`
			Tag       string `json:"tag"`
			Amount    string `json:"amount"`
			Contract  bool   `json:"contract"`
		} `json:"inputDetails"`
		OutputDetails []struct {
			OutputHash string `json:"outputHash"`
			Tag        string `json:"tag"`
			Amount     string `json:"amount"`
			Contract   bool   `json:"contract"`
		} `json:"outputDetails"`
		State                string                 `json:"state"`
		GasLimit             string                 `json:"gasLimit"`
		GasUsed              string                 `json:"gasUsed"`
		GasPrice             string                 `json:"gasPrice"`
		TotalTransactionSize string                 `json:"totalTransactionSize"`
		VirtualSize          string                 `json:"virtualSize"`
		Weight               string                 `json:"weight"`
		Nonce                string                 `json:"nonce"`
		TransactionType      string                 `json:"transactionType"`
		TokenTransferDetails []*TokenTransferDetail `json:"tokenTransferDetails"`
		ContractDetails      []interface{}          `json:"contractDetails"`
	} `json:"data"`
}

type TokenHolderListResp struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		Page              string `json:"page"`
		Limit             string `json:"limit"`
		TotalPage         string `json:"totalPage"`
		ChainFullName     string `json:"chainFullName"`
		ChainShortName    string `json:"chainShortName"`
		CirculatingSupply string `json:"circulatingSupply"`
		PositionList      []struct {
			HolderAddress     string `json:"holderAddress"`
			Amount            string `json:"amount"`
			ValueUsd          string `json:"valueUsd"`
			PositionChange24H string `json:"positionChange24h"`
			Rank              string `json:"rank"`
		} `json:"positionList"`
	} `json:"data"`
}

type AddressDetailResp struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		ChainFullName                 string `json:"chainFullName"`
		ChainShortName                string `json:"chainShortName"`
		Address                       string `json:"address"`
		ContractAddress               string `json:"contractAddress"`
		Balance                       string `json:"balance"`
		BalanceSymbol                 string `json:"balanceSymbol"`
		TransactionCount              string `json:"transactionCount"`
		Verifying                     string `json:"verifying"`
		SendAmount                    string `json:"sendAmount"`
		ReceiveAmount                 string `json:"receiveAmount"`
		TokenAmount                   string `json:"tokenAmount"`
		TotalTokenValue               string `json:"totalTokenValue"`
		CreateContractAddress         string `json:"createContractAddress"`
		CreateContractTransactionHash string `json:"createContractTransactionHash"`
		FirstTransactionTime          string `json:"firstTransactionTime"`
		LastTransactionTime           string `json:"lastTransactionTime"`
		Token                         string `json:"token"`
		Bandwidth                     string `json:"bandwidth"`
		Energy                        string `json:"energy"`
		VotingRights                  string `json:"votingRights"`
		UnclaimedVotingRewards        string `json:"unclaimedVotingRewards"`
		Tag                           string `json:"tag"`
	} `json:"data"`
}

type AddressBalanceResp struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		Page           string `json:"page"`
		Limit          string `json:"limit"`
		TotalPage      string `json:"totalPage"`
		ChainFullName  string `json:"chainFullName"`
		ChainShortName string `json:"chainShortName"`
		TokenList      []struct {
			Token           string `json:"token"`
			HoldingAmount   string `json:"holdingAmount"`
			TotalTokenValue string `json:"totalTokenValue"`
			Change24H       string `json:"change24h"`
			PriceUsd        string `json:"priceUsd"`
			ValueUsd        string `json:"valueUsd"`
		} `json:"tokenList"`
	} `json:"data"`
}

type GasFeeResp struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		ChainFullName       string `json:"chainFullName"`
		ChainShortName      string `json:"chainShortName"`
		Symbol              string `json:"symbol"`
		BestTransactionFee  string `json:"bestTransactionFee"`
		RecommendedGasPrice string `json:"recommendedGasPrice"`
	} `json:"data"`
}
