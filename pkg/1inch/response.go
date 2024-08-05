package inch

type QuoteResp struct {
	FromToken struct {
		Symbol   string `json:"symbol"`
		Name     string `json:"name"`
		Address  string `json:"address"`
		Decimals int    `json:"decimals"`
		LogoURI  string `json:"logoURI"`
		IsCustom bool   `json:"isCustom"`
		IsFOT    bool   `json:"isFOT"`
	} `json:"fromToken"`
	ToToken struct {
		Symbol        string   `json:"symbol"`
		Name          string   `json:"name"`
		Address       string   `json:"address"`
		Decimals      int      `json:"decimals"`
		LogoURI       string   `json:"logoURI"`
		Eip2612       bool     `json:"eip2612"`
		DomainVersion string   `json:"domainVersion"`
		Tags          []string `json:"tags"`
	} `json:"toToken"`
	ToTokenAmount   string `json:"toTokenAmount"`
	FromTokenAmount string `json:"fromTokenAmount"`
	Protocols       [][][]struct {
		Name             string `json:"name"`
		Part             int    `json:"part"`
		FromTokenAddress string `json:"fromTokenAddress"`
		ToTokenAddress   string `json:"toTokenAddress"`
	} `json:"protocols"`
	EstimatedGas int `json:"estimatedGas"`
}

type SwapResp struct {
	FromToken struct {
		Symbol   string   `json:"symbol"`
		Name     string   `json:"name"`
		Decimals int      `json:"decimals"`
		Address  string   `json:"address"`
		LogoURI  string   `json:"logoURI"`
		Tags     []string `json:"tags"`
	} `json:"fromToken"`
	ToToken struct {
		Symbol   string   `json:"symbol"`
		Name     string   `json:"name"`
		Decimals int      `json:"decimals"`
		Address  string   `json:"address"`
		LogoURI  string   `json:"logoURI"`
		Eip2612  bool     `json:"eip2612"`
		Tags     []string `json:"tags"`
	} `json:"toToken"`
	ToTokenAmount   string `json:"toTokenAmount"`
	FromTokenAmount string `json:"fromTokenAmount"`
	Protocols       [][][]struct {
		Name             string `json:"name"`
		Part             int    `json:"part"`
		FromTokenAddress string `json:"fromTokenAddress"`
		ToTokenAddress   string `json:"toTokenAddress"`
	} `json:"protocols"`
	Tx TxData
}

type TxData struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Data     string `json:"data"`
	Value    string `json:"value"`
	Gas      int    `json:"gas"`
	GasPrice string `json:"gasPrice"`
}

type ApproveTransactionResp struct {
	Data     string `json:"data"`
	GasPrice string `json:"gasPrice"`
	To       string `json:"to"`
	Value    string `json:"value"`
}
