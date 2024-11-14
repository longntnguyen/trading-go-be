package model

import (
	"encoding/json"
	"math/big"
	"time"
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
	TotalBalance       float64         `json:"totalBalance"`
	TotalBalanceInUSD  float64         `json:"totalBalanceInUSD"`
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

type ListTokenInfoResponse struct {
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
	Result []ListTokenInfo `json:"result"`
}

type ListTokenInfo struct {
	TokenID      string `json:"tokenID"`
	Symbol       string `json:"symbol"`
	TokenName    string `json:"tokenName"`
	ImageUrl     string `json:"imageUrl"`
	TokenAddress string `json:"tokenAddress"`
}

type CoinMarketCapStatus struct {
	Timestamp    time.Time `json:"timestamp"`
	ErrorCode    int       `json:"error_code"`
	ErrorMessage *string   `json:"error_message"`
	Elapsed      int       `json:"elapsed"`
	CreditCount  int       `json:"credit_count"`
	Notice       *string   `json:"notice"`
}

type CoinMarketCapGetListTokenMapDataItem struct {
	ID                  int         `json:"id"`
	Rank                int         `json:"rank"`
	Name                string      `json:"name"`
	Symbol              string      `json:"symbol"`
	Slug                string      `json:"slug"`
	IsActive            int         `json:"is_active"`
	FirstHistoricalData time.Time   `json:"first_historical_data"`
	LastHistoricalData  time.Time   `json:"last_historical_data"`
	Platform            interface{} `json:"platform"`
}

type CoinMarketCapGetListTokenMapResponse struct {
	Status CoinMarketCapStatus                    `json:"status"`
	Data   []CoinMarketCapGetListTokenMapDataItem `json:"data"`
}

type CoinMarketCapGetListTokenInfoResponse struct {
	Status struct {
		Timestamp    string  `json:"timestamp"`
		ErrorCode    int     `json:"error_code"`
		ErrorMessage *string `json:"error_message"`
		Elapsed      int     `json:"elapsed"`
		CreditCount  int     `json:"credit_count"`
		Notice       *string `json:"notice"`
	} `json:"status"`
	Data map[string]struct {
		ID          int      `json:"id"`
		Name        string   `json:"name"`
		Symbol      string   `json:"symbol"`
		Category    string   `json:"category"`
		Description string   `json:"description"`
		Slug        string   `json:"slug"`
		Logo        string   `json:"logo"`
		Subreddit   string   `json:"subreddit"`
		Notice      string   `json:"notice"`
		Tags        []string `json:"tags"`
		TagNames    []string `json:"tag-names"`
		TagGroups   []string `json:"tag-groups"`
		URLs        struct {
			Website      []string `json:"website"`
			Twitter      []string `json:"twitter"`
			MessageBoard []string `json:"message_board"`
			Chat         []string `json:"chat"`
			Facebook     []string `json:"facebook"`
			Explorer     []string `json:"explorer"`
			Reddit       []string `json:"reddit"`
			TechnicalDoc []string `json:"technical_doc"`
			SourceCode   []string `json:"source_code"`
			Announcement []string `json:"announcement"`
		} `json:"urls"`
		Platform struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Slug         string `json:"slug"`
			Symbol       string `json:"symbol"`
			TokenAddress string `json:"token_address"`
		} `json:"platform,omitempty"`
		DateAdded         string  `json:"date_added"`
		TwitterUsername   string  `json:"twitter_username"`
		IsHidden          int     `json:"is_hidden"`
		DateLaunched      *string `json:"date_launched,omitempty"`
		ContractAddresses []struct {
			ContractAddress string `json:"contract_address"`
			Platform        struct {
				Name string `json:"name"`
				Coin struct {
					ID     string `json:"id"`
					Name   string `json:"name"`
					Symbol string `json:"symbol"`
					Slug   string `json:"slug"`
				} `json:"coin"`
			} `json:"platform"`
		} `json:"contract_address,omitempty"`
		SelfReportedCirculatingSupply *float64 `json:"self_reported_circulating_supply,omitempty"`
		SelfReportedTags              *string  `json:"self_reported_tags,omitempty"`
		SelfReportedMarketCap         *float64 `json:"self_reported_market_cap,omitempty"`
		InfiniteSupply                bool     `json:"infinite_supply"`
	} `json:"data"`
}
