package collector

import (
	"strconv"
	"time"

	"smart-money/pkg/log"
	"smart-money/pkg/oklink"
)

var (
	collectors = make(map[string]Collector)
)

type Collector interface {
	Name() string
	Collect(chainName string, params any) ([]string, error)
}

func Factory(name string) Collector {
	switch name {
	case "token_holders":
		return &TokenHolders{}
	case "manual_input":
		return &ManualInput{}
	}

	return nil
}

func filterAddress(addressDetail *oklink.AddressDetailResp) (bool, error) {
	detail := addressDetail.Data[0]

	if detail.ContractAddress != "" {
		log.Warnf("address[%v] is contract address", detail.Address)
		return false, nil
	}
	if detail.Balance <= "0.02" {
		log.Warnf("address[%v] balance[%v] is too small", detail.Address, detail.Balance)
		return false, nil
	}
	// 如果第一次交易是30天前，那么认为是新地址，不予入库
	firstTs, err := strconv.Atoi(detail.FirstTransactionTime)
	if err != nil {
		return false, err
	}
	firstTxTime := time.UnixMilli(int64(firstTs))
	lastTs, err := strconv.Atoi(detail.LastTransactionTime)
	if err != nil {
		return false, err
	}
	lastTxTime := time.UnixMilli(int64(lastTs))
	if (lastTxTime.Unix() - firstTxTime.Unix()) < 10*24*3600 {
		log.Warnf("address[%v] is too new, first tx time[%v], last tx time[%v]", detail.Address, firstTxTime, lastTxTime)
		return false, nil
	}

	return true, nil
}
