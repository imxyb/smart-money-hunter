package collector

import (
	"github.com/mitchellh/mapstructure"
	"smart-money/pkg/oklink"
)

type ManualInputParams struct {
	Addresses []string `mapstructure:"addresses"`
}

type ManualInput struct {
}

func (m *ManualInput) Name() string {
	return "manual_input"
}

func (m *ManualInput) Collect(chainName string, params any) ([]string, error) {
	var mp ManualInputParams
	if err := mapstructure.Decode(params, &mp); err != nil {
		return nil, err
	}

	var addresses []string

	for _, address := range mp.Addresses {
		addressDetail, err := oklink.Api.GetAddressDetail(chainName, address)
		if err != nil {
			return nil, err
		}

		valid, err := filterAddress(addressDetail)
		if err != nil {
			return nil, err
		}
		if !valid {
			continue
		}

		addresses = append(addresses, address)
	}
	return addresses, nil
}
