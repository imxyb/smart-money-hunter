package v1

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/panjf2000/ants/v2"
	ccollector "smart-money/internal/collector"
	"smart-money/internal/hunter"
	"smart-money/pkg/errcode"
	"smart-money/pkg/log"
	"smart-money/pkg/model"
	"smart-money/pkg/response"
)

type WrapHunterWork struct {
	TaskName string `json:"task_name"`
	Status   int    `json:"status"`
}

var hunterStatus = make(map[string]*WrapHunterWork)

type WorkRequest struct {
	ChainName       string         `json:"chain_name"`
	TaskName        string         `json:"task_name"`
	CollectorName   string         `json:"collector_name"`
	CollectorParams map[string]any `json:"collector_params"`
	CollectSeconds  int64          `json:"collect_seconds"`
}

func Work(c *gin.Context) {
	var req *WorkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, errcode.WorkParamsError, err)
		return
	}

	if req.TaskName == "" {
		response.BadRequest(c, errcode.WorkParamsError, fmt.Errorf("task name is empty"))
		return
	}
	if req.CollectorName == "" {
		response.BadRequest(c, errcode.WorkParamsError, fmt.Errorf("collector name is empty"))
		return
	}
	if req.CollectorParams == nil {
		response.BadRequest(c, errcode.WorkParamsError, fmt.Errorf("collector params is empty"))
		return
	}
	if req.CollectSeconds < 3600*24 {
		response.BadRequest(c, errcode.WorkParamsError, fmt.Errorf("collect seconds is too short"))
		return
	}

	go func() {
		ht := hunter.NewHunter(req.ChainName, req.TaskName, ccollector.Factory(req.CollectorName), req.CollectorParams, req.CollectSeconds)
		wht := &WrapHunterWork{
			TaskName: req.TaskName,
			Status:   0,
		}
		hunterStatus[req.TaskName] = wht
		if err := ht.Work(); err != nil {
			hunterStatus[req.TaskName].Status = 2
			log.Errorf("hunter work error: %v", err)
			return
		}
		hunterStatus[req.TaskName].Status = 1
	}()
	response.OK(c, nil)
}

func ListWorkStatus(c *gin.Context) {
	response.OK(c, hunterStatus)
}

type ListAddressTradeRequest struct {
	Start string `form:"start"`
	End   string `form:"end"`
}

type AddressTradeDetail struct {
	Address            string  `json:"address"`
	TradeCount         int     `json:"trade_count"`
	BuyTotalUsd        float64 `json:"buy_total_usd"`
	SellTotalUsd       float64 `json:"sell_total_usd"`
	ProfitTotalUsd     float64 `json:"profit_total_usd"`
	WinTotal           int     `json:"win_total"`
	LoseTotal          int     `json:"lose_total"`
	WinRate            float64 `json:"win_rate"`
	MaxMultiPle        float64 `json:"max_multiple"`
	HoldingAvgDuration int64   `json:"holding_avg_duration"`
}

type ListAddressTradeResponse []*AddressTradeDetail

func ListAddressTrade(c *gin.Context) {
	var req ListAddressTradeRequest
	if err := c.Bind(&req); err != nil {
		response.BadRequest(c, errcode.WorkParamsError, err)
		return
	}

	st, err := time.Parse("2006-01-02", req.Start)
	if err != nil {
		response.BadRequest(c, errcode.WorkParamsError, err)
		return
	}
	startTime := time.Date(st.Year(), st.Month(), st.Day(), 0, 0, 0, 0, st.Location())

	if startTime.IsZero() {
		response.BadRequest(c, errcode.WorkParamsError, fmt.Errorf("start time is invalid"))
		return
	}

	var endTime time.Time
	if req.End == "" {
		endTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 23, 59, 59, 0, startTime.Location())
	} else {
		endTime, err = time.Parse("2006-01-02", req.End)
		if err != nil {
			response.BadRequest(c, errcode.WorkParamsError, err)
			return
		}
		if endTime.IsZero() {
			response.BadRequest(c, errcode.WorkParamsError, fmt.Errorf("end time is invalid"))
			return
		}
	}

	if startTime.After(endTime) {
		response.BadRequest(c, errcode.WorkParamsError, fmt.Errorf("start time is greater than end time"))
		return
	}

	startUnix := startTime.Unix()
	endUnix := endTime.Unix()

	var distinctAddressTrade []*model.AddressTrade
	err = model.GetDB().Distinct("address").Where("first_tx_time between ? and ?", startUnix, endUnix).Find(&distinctAddressTrade).Error
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	var addressList []string
	for _, trade := range distinctAddressTrade {
		addressList = append(addressList, trade.Address)
	}

	resp := make(ListAddressTradeResponse, 0, len(addressList))

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	p, _ := ants.NewPoolWithFunc(100, func(i interface{}) {
		defer wg.Done()
		address := i.(string)

		var (
			trades               []*model.AddressTrade
			tradeCount           int
			buyTotalUsd          float64
			sellTotalUsd         float64
			profitTotalUsd       float64
			winTotal             int
			loseTotal            int
			winRate              float64
			holdingTotalDuration uint64
		)

		err = model.GetDB().Where("address=? and first_tx_time between ? and ?", address, startUnix, endUnix).Find(&trades).Error
		if err != nil {
			response.InternalServerError(c, err)
			return
		}
		tradeCount = len(trades)
		var maxMultiPle float64

		for _, trade := range trades {
			if trade.Profit > 0 {
				winTotal++
			} else {
				loseTotal++
			}

			buyTotalUsd += trade.BuyTotalUsd
			sellTotalUsd += trade.SellTotalUsd
			profitTotalUsd += trade.Profit
			winRate = float64(winTotal) / float64(tradeCount)
			holdingTotalDuration += trade.LastTxTime - trade.FirstTxTime

			if trade.SellTotalUsd/trade.BuyTotalUsd > maxMultiPle {
				maxMultiPle = trade.SellTotalUsd / trade.BuyTotalUsd
			}
		}

		detail := &AddressTradeDetail{
			Address:            address,
			TradeCount:         tradeCount,
			BuyTotalUsd:        buyTotalUsd,
			SellTotalUsd:       sellTotalUsd,
			ProfitTotalUsd:     profitTotalUsd,
			WinTotal:           winTotal,
			LoseTotal:          loseTotal,
			WinRate:            winRate,
			MaxMultiPle:        maxMultiPle,
			HoldingAvgDuration: int64(holdingTotalDuration) / int64(tradeCount),
		}

		mu.Lock()
		resp = append(resp, detail)
		mu.Unlock()
	})
	defer p.Release()

	for _, address := range addressList {
		wg.Add(1)
		p.Invoke(address)
	}

	wg.Wait()

	response.OK(c, resp)
}
