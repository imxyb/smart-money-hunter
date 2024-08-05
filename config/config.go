package config

import (
	"github.com/go-ini/ini"
)

var CFG Config

type Config struct {
	Log    Log    `ini:"log"`
	OkLink OkLink `ini:"oklink"`
	Web3   Web3   `ini:"web3"`
	Server Server `ini:"server"`
	Redis  Redis  `ini:"redis"`
}

type Server struct {
	Port int `ini:"port"`
}

type OkLink struct {
	Host   string `ini:"host"`
	ApiKey string `ini:"apikey"`
}

type Log struct {
	Level      string `ini:"level"`
	File       string `ini:"file"`
	ErrorLevel string `ini:"error_level"`
	ErrorFile  string `ini:"error_file"`
}

type Web3 struct {
	Rpc       string `ini:"rpc"`
	ChainID   int64  `ini:"chain_id"`
	ChainName string `ini:"chain_name"`
}

type Redis struct {
	Addr string `ini:"addr"`
}

func Init(path string) error {
	if err := ini.MapTo(&CFG, path); err != nil {
		return err
	}
	return nil
}
