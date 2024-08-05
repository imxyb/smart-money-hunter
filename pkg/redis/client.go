package redis

import (
	"context"
	"strings"

	vredis "github.com/go-redis/redis/v8"
)

var (
	Client *vredis.Client
)

// InitSingleClient 实例化单个redis client
func InitSingleClient(addr string) error {
	url := strings.Replace(addr, "tcp", "redis", -1)
	opt, err := vredis.ParseURL(url)
	if err != nil {
		return err
	}
	Client = vredis.NewClient(opt)
	reply := Client.Ping(context.Background())
	if reply.Err() != nil {
		return reply.Err()
	}
	return nil
}
