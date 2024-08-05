package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rivo/tview"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v2"
	v1 "smart-money/api/v1"
	"smart-money/config"
	"smart-money/internal/cron"
	"smart-money/pkg/eth"
	"smart-money/pkg/log"
	"smart-money/pkg/model"
	"smart-money/pkg/oklink"
	"smart-money/pkg/redis"
)

func main() {
	app := &cli.App{
		Name:   "auction house user endpoint",
		Before: cliBefore,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "conf",
				Usage:    "config file path",
				Required: true,
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "server",
				Usage:  "start server",
				Action: server,
			},
			{
				Name:   "test",
				Usage:  "start server",
				Action: test,
			},
			{
				Name:   "test2",
				Usage:  "start server",
				Action: test2,
			},
			{
				Name:   "work",
				Action: work,
			},
			{
				Name:   "listwork",
				Action: listWork,
			},
			{
				Name:   "listaddresstrade",
				Action: listAddressTrade,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "start",
						Usage:    "start date",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "end",
						Usage:    "end date",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "csv",
						Usage: "export csv",
					},
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func cliBefore(c *cli.Context) error {
	var err error
	if err = config.Init(c.String("conf")); err != nil {
		return err
	}

	cfg := config.CFG

	if err = log.InitLogger(
		cfg.Log.Level, cfg.Log.File, cfg.Log.ErrorLevel,
		cfg.Log.ErrorFile,
	); err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if err = model.InitDB(filepath.Join(homeDir, "smart-money-hunter.db")); err != nil {
		return err
	}

	oklink.InitAPI(cfg.OkLink.ApiKey, cfg.OkLink.Host)

	if err = eth.InitClient(config.CFG.Web3.Rpc, config.CFG.Web3.ChainID); err != nil {
		return err
	}

	if err = redis.InitSingleClient(config.CFG.Redis.Addr); err != nil {
		return err
	}

	cron.Init()

	return nil
}

func server(c *cli.Context) error {
	v1.Start()
	return nil
}

func listAddressTrade(c *cli.Context) error {
	url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/list_address_trade", config.CFG.Server.Port)
	reqC := req.C()
	resp := reqC.Get(url).SetQueryParamsAnyType(map[string]interface{}{
		"start": c.String("start"),
		"end":   c.String("end"),
	}).Do()
	if resp.Err != nil {
		return resp.Err
	}
	if resp.IsErrorState() {
		return fmt.Errorf("get url failed, status code:%d, content:%v", resp.GetStatusCode(), resp.String())
	}

	result := gjson.Get(resp.String(), "data").String()

	var response v1.ListAddressTradeResponse
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"address", "trade_count", "buy_total_usd", "sell_total_usd", "profit_total_usd", "win_total",
		"lose_total", "win_rate", "max_multiple", "holding_avg_duration"})

	for _, detail := range response {
		t.AppendRow(table.Row{detail.Address, detail.TradeCount, detail.BuyTotalUsd,
			detail.SellTotalUsd, detail.ProfitTotalUsd, detail.WinTotal, detail.LoseTotal, detail.WinRate, detail.MaxMultiPle, detail.HoldingAvgDuration})
	}

	if c.Bool("csv") {
		t.RenderCSV()
	} else {
		t.Render()
	}
	return nil
}

func listWork(c *cli.Context) error {
	url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/list_work_status", config.CFG.Server.Port)
	reqC := req.C()
	resp := reqC.Get(url).Do()
	if resp.Err != nil {
		return resp.Err
	}
	if resp.IsErrorState() {
		return fmt.Errorf("get url failed, status code:%d, content:%v", resp.GetStatusCode(), resp.String())
	}

	commonResp := make(map[string]any)
	if err := resp.UnmarshalJson(&commonResp); err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"task_name", "status"})

	v := commonResp["data"].(map[string]interface{})
	for _, i := range v {
		wh := i.(map[string]interface{})
		status := wh["status"].(float64)
		var statusText string
		if status == 0 {
			statusText = "running"
		} else if status == 1 {
			statusText = "stopped"
		} else {
			statusText = "finished"
		}
		t.AppendRow(table.Row{wh["task_name"], statusText})
	}
	t.Render()

	return nil
}

func work(c *cli.Context) error {
	workUrl := fmt.Sprintf("http://127.0.0.1:%d/api/v1/work", config.CFG.Server.Port)
	reqC := req.C()
	app := tview.NewApplication()
	form := tview.NewForm()

	form.AddDropDown("chain name", []string{"eth", "bsc", "arb"}, 0, nil).
		AddDropDown("collector name", []string{"manual_input", "token_holders"}, 0, nil).
		AddTextArea("collector params", "", 50, 10, 0, nil).
		AddInputField("task name", "", 20, nil, nil).
		AddInputField("collect seconds", "", 20, nil, nil).
		AddButton("Save", func() {
			_, chainName := form.GetFormItemByLabel("chain name").(*tview.DropDown).GetCurrentOption()
			_, collectorName := form.GetFormItemByLabel("collector name").(*tview.DropDown).GetCurrentOption()
			collectorParams := form.GetFormItemByLabel("collector params").(*tview.TextArea).GetText()
			taskName := form.GetFormItemByLabel("task name").(*tview.InputField).GetText()
			collectSecondsStr := form.GetFormItemByLabel("collect seconds").(*tview.InputField).GetText()

			collectParamsMap := make(map[string]any)
			if err := json.Unmarshal([]byte(collectorParams), &collectParamsMap); err != nil {
				panic(err)
			}
			collectSeconds, err := strconv.Atoi(collectSecondsStr)
			if err != nil {
				panic(err)
			}

			resp := reqC.Post(workUrl).SetBodyJsonMarshal(map[string]any{
				"chain_name":       chainName,
				"collector_name":   collectorName,
				"collector_params": collectParamsMap,
				"task_name":        taskName,
				"collect_seconds":  collectSeconds,
			}).Do()

			if resp.Err != nil {
				panic(resp.Err)
			}
			if resp.IsErrorState() {
				panic(fmt.Errorf("get url failed, status code:%d, content:%v", resp.GetStatusCode(), resp.String()))
			}

			modal := tview.NewModal().SetText("任务创建成功").AddButtons([]string{"OK"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				app.Stop()
			})
			app.SetRoot(modal, true).EnableMouse(true)
		}).
		AddButton("Quit", func() {
			app.Stop()
		})
	form.SetBorder(true).SetTitle("启动一个收集任务").SetTitleAlign(tview.AlignLeft)
	if err := app.SetRoot(form, true).EnableMouse(true).Run(); err != nil {
		return err
	}
	return nil
}

func collect(cn string, address string) error {
	startTs := 1685445600000
	endTs := 1685457000000
	key := "airdrop_addr"
	now := time.Now()
	page := 1
loop:
	for {
		resp, err := oklink.Api.GetNormalTransactionListByAddressAndToken(cn, address, "transaction", page, 100)
		if err != nil {
			return err
		}
		if len(resp.Data) == 0 {
			return nil
		}
		log.Infof("page:%d, len:%d", page, len(resp.Data[0].TransactionLists))
		for _, data := range resp.Data {
			for _, tx := range data.TransactionLists {
				txTime, _ := strconv.Atoi(tx.TransactionTime)
				if txTime < startTs {
					break loop
				}
				if txTime > endTs {
					log.Infof("tx time:%v, now:%v", txTime, endTs)
					continue
				}
				if strings.EqualFold(tx.From, address) {
					log.Infof("from address:%s, skip", address)
					continue
				}
				fromAddress := tx.From
				exist, err := redis.Client.HExists(context.Background(), key, fromAddress).Result()
				if err != nil {
					return err
				}
				if exist {
					log.Infof("address %s already exist", fromAddress)
					continue
				}
				fresp, err := oklink.Api.GetNormalTransactionListByAddressAndToken(cn, fromAddress, "", 1, 5)
				if err != nil {
					return err
				}
				if len(fresp.Data) == 0 {
					log.Infof("from address:%v tx count is 0, cn:%s", fromAddress, cn)
					continue
				}
				if len(fresp.Data[0].TransactionLists) < 5 {
					log.Infof("from address:%s tx count less than 5, cn:%v", fromAddress, cn)
					continue
				}
				firstTx := fresp.Data[0].TransactionLists[len(fresp.Data[0].TransactionLists)-1]
				ft, _ := strconv.Atoi(firstTx.TransactionTime)
				// 大于一个月
				if now.UnixMilli()-int64(ft) > 2592000*1000 {
					log.Infof("from address tx time more than 1 month, address:%s, tx time:%s, cn:%v", fromAddress, time.UnixMilli(int64(ft)), cn)
					continue
				}

				redis.Client.HSet(context.Background(), key, tx.From, cn)
			}
		}
		page += 1
	}

	return nil
}

func test2(c *cli.Context) error {
	f, err := os.Open("/Users/imxyb/Desktop/arb.csv")
	if err != nil {
		return err
	}
	reader := csv.NewReader(f)
	data, err := reader.ReadAll()
	if err != nil {
		return err
	}
	key := "airdrop_addr"
	now := time.Now()

	for _, datum := range data {
		fromAddress := datum[0]
		exist, err := redis.Client.HExists(context.Background(), key, fromAddress).Result()
		if err != nil {
			return err
		}
		if exist {
			log.Infof("address %s already exist", fromAddress)
			continue
		}
		fresp, err := oklink.Api.GetNormalTransactionListByAddressAndToken("Arbitrum", fromAddress, "", 1, 5)
		if err != nil {
			return err
		}
		if len(fresp.Data) == 0 {
			log.Infof("from address:%v tx count is 0, cn:%s", fromAddress, "Arbitrum")
			continue
		}
		if len(fresp.Data[0].TransactionLists) < 5 {
			log.Infof("from address:%s tx count less than 5, cn:%v", fromAddress, "Arbitrum")
			continue
		}
		firstTx := fresp.Data[0].TransactionLists[len(fresp.Data[0].TransactionLists)-1]
		ft, _ := strconv.Atoi(firstTx.TransactionTime)
		// 大于一个月
		if now.UnixMilli()-int64(ft) > 2592000*1000 {
			log.Infof("from address tx time more than 1 month, address:%s, tx time:%s, cn:%v", fromAddress, time.UnixMilli(int64(ft)), "Arbitrum")
			continue
		}

		redis.Client.HSet(context.Background(), key, fromAddress, "Arbitrum")
	}
	return nil
}

func test(c *cli.Context) error {
	page := 1
	sell := decimal.NewFromFloat(0)
	buy := decimal.NewFromFloat(0)
	tokens := make(map[string]decimal.Decimal)
	for {
		resp, err := oklink.Api.GetToken20TransactionListByAddress("eth", "0xd185b0a98b043434e437ab8b964eea3297f44a04", page, 100)
		if err != nil {
			panic(err)
		}
		if len(resp.Data[0].TransactionLists) == 0 {
			break
		}
		m := make(map[string]bool)
		for _, data := range resp.Data {
			for _, tx := range data.TransactionLists {
				t, err := oklink.Api.GetTransactionDetail("eth", tx.TxId)
				if err != nil {
					panic(err)
				}

				if m[tx.TxId] {
					continue
				}
				if t.Data[0].TokenTransferDetails[0].Symbol == "WETH" {
					amount, err := decimal.NewFromString(t.Data[0].TokenTransferDetails[0].Amount)
					if err != nil {
						panic(err)
					}
					sell = sell.Add(amount)
					to := tokens[t.Data[0].TokenTransferDetails[len(t.Data[0].TokenTransferDetails)-1].Symbol]
					tokens[t.Data[0].TokenTransferDetails[len(t.Data[0].TokenTransferDetails)-1].Symbol] = to.Sub(amount)
				} else if t.Data[0].TokenTransferDetails[len(t.Data[0].TokenTransferDetails)-1].Symbol == "WETH" {
					amount, err := decimal.NewFromString(t.Data[0].TokenTransferDetails[len(t.Data[0].TokenTransferDetails)-1].Amount)
					if err != nil {
						panic(err)
					}
					buy = buy.Add(amount)
					to := tokens[t.Data[0].TokenTransferDetails[0].Symbol]
					tokens[t.Data[0].TokenTransferDetails[0].Symbol] = to.Add(amount)
					//fmt.Println("get for", t.Data[0].TokenTransferDetails[0].Symbol, amount)
				}
				m[tx.TxId] = true
			}
		}
		page += 1
	}
	b, err := json.MarshalIndent(tokens, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	return nil
}
