package model

import (
	"encoding/json"
	"math/big"
)

type TokenOverView struct {
	TokenName        string      `json:"tokenName"`
	Balance          big.Float   `json:"balance"`
	TokenID          json.Number `json:"tokenID"`
	Symbol           string      `json:"symbol"`
	BalanceInUSD     big.Float   `json:"balanceInUSD"`
	PercentChange24h big.Float   `json:"percentChange24h"`
	Volume24H        big.Float   `json:"volume24h"`
	MarketCap        big.Float   `json:"marketCap"`
	ImageUrl         string      `json:"imageUrl"`
}

type OverViewResponse struct {
	TotalBalance       big.Float       `json:"totalBalance"`
	TotalBalanceInUSD  big.Float       `json:"totalBalanceInUSD"`
	TokensOverViewList []TokenOverView `json:"tokensOverViewList"`
}

type CoinMarketCapResponse struct {
	Data map[string]struct {
		ID    json.Number `json:"id"`
		Quote map[string]struct {
			Price            float64 `json:"price"`
			Volume24h        float64 `json:"volume_24h"`
			PercentChange24h float64 `json:"percent_change_24h"`
			MarketCap        float64 `json:"market_cap"`
		} `json:"quote"`
	} `json:"data"`
}

type TokenBalanceInfo struct {
	Symbol           string
	Balance          *big.Float
	PercentChange24h *big.Float
	Volume24H        *big.Float
	MarketCap        *big.Float
	TokenID          json.Number
}

type TokenImage struct {
	Name   string `json:"name"`
	Images string `json:"images"`
}
