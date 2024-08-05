package util

func IsMainToken(symbol string) bool {
	if symbol == "WBNB" || symbol == "BNB" || symbol == "ETH" || symbol == "WETH" || symbol == "USDT" || symbol == "USDC" || symbol == "DAI" {
		return true
	}
	return false
}

func IsStableToken(symbol string) bool {
	if symbol == "USDT" || symbol == "USDC" || symbol == "DAI" {
		return true
	}
	return false
}
