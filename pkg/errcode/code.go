package errcode

// 错误码设计规范, eg: 10101
// 1 常规开头
// 01 表示认证模块
// 01 具体错误码

const (
	InternalServerError = 10000

	WorkParamsError = 11000

	SaveWalletParamsError   = 12000
	SaveWalletExistError    = 12002
	SaveWalletNotExistError = 12003

	SaveFollowAddressParamsError   = 13000
	SaveFollowAddressExistError    = 13002
	SaveFollowAddressNotExistError = 13003
	ListFollowAddressParamsError   = 13004

	ListFollowTradeParamsError = 14000
)
