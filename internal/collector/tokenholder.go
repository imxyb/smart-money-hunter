package collector

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"smart-money/pkg/oklink"
)

type TokenHoldersParams struct {
	TokenAddress string `mapstructure:"token_address"`
	TopN         int    `mapstructure:"topn"`
}

type TokenHolders struct {
}

func (t *TokenHolders) Name() string {
	return "token_holders"
}

func (t *TokenHolders) Collect(chainName string, params any) ([]string, error) {
	var tp TokenHoldersParams
	if err := mapstructure.Decode(params, &tp); err != nil {
		return nil, err
	}

	if tp.TokenAddress == "" {
		return nil, fmt.Errorf("token_address is empty")
	}
	if tp.TopN == 0 || tp.TopN > 100 {
		return nil, fmt.Errorf("topn is invalid")
	}

	resp, err := oklink.Api.GetTokenHolderList(chainName, tp.TokenAddress, 1, tp.TopN)
	if err != nil {
		return nil, err
	}

	var addresses []string
	for _, data := range resp.Data {
		for _, s := range data.PositionList {
			addressDetail, err := oklink.Api.GetAddressDetail(chainName, s.HolderAddress)
			valid, err := filterAddress(addressDetail)
			if err != nil {
				return nil, err
			}
			if !valid {
				continue
			}
			addresses = append(addresses, s.HolderAddress)
		}
	}
	return addresses, nil
}
