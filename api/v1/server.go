package v1

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"smart-money/config"
)

func Start() {
	r := gin.Default()

	group := r.Group("/api/v1")
	{
		{
			group.POST("/work", Work)
			group.GET("/list_work_status", ListWorkStatus)
			group.GET("/list_address_trade", ListAddressTrade)
		}

		{
			group.GET("/list_wallet", ListWallet)
			group.POST("/create_wallet", CreateWallet)
			group.POST("/update_wallet", UpdateWallet)
			group.POST("/delete_wallet", DeleteWallet)
		}

		{
			group.GET("/list_follow_address", ListFollowAddress)
			group.POST("/create_follow_address", CreateFollowAddress)
			group.POST("/update_follow_address", UpdateFollowAddress)
			group.POST("/delete_follow_address", DeleteFollowAddress)
		}

		{
			group.GET("/list_follow_trade", ListFollowTrade)
		}
	}

	r.Run(fmt.Sprintf(":%d", config.CFG.Server.Port))
}
