package constants

type Token struct {
	Name    string
	Address string
	Symbol  string
}

var TOKEN_LIST = []Token{
	{
		Name:    "Bitcoin",
		Address: "0x71CCe0035d82c21Cf4b908bcd8F1117fFf0Fa623",
		Symbol:  "BTC",
	},
	{
		Name:    "BNB",
		Address: "0xb8c77482e45f1f44de1745f52c74426c631bdd52",
		Symbol:  "BNB",
	},
	{
		Name:    "SHIBA",
		Address: "0x2859e4544C4bB03966803b044A93563Bd2D0DD4D",
		Symbol:  "SHIB",
	},
	{
		Name:    "USDT",
		Address: "0xdac17f958d2ee523a2206206994597c13d831ec7",
		Symbol:  "USDC",
	},
}
