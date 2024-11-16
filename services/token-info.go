package services

import (
	"encoding/json"
	"fmt"
	"my-app/constants"
	"my-app/model"
	"net/http"
	"os"
	"strings"
)

func GetListTokenInfo(page, limit int) ([]model.ListTokenInfo, error) {
	apiKey := os.Getenv("COINMARKETCAP_API_KEY")

	// Get list token map, using: https://pro-api.coinmarketcap.com/v1/cryptocurrency/map
	url := fmt.Sprintf("https://pro-api.coinmarketcap.com/v1/cryptocurrency/map?start=%d&limit=%d", page, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Add("X-CMC_PRO_API_KEY", apiKey)
	// Get list token info, using: https://pro-api.coinmarketcap.com/v2/cryptocurrency/info
	// var idStrings []string
	// for _, data := range cmcGetListTokenMapResponse.Data {
	// 	idStrings = append(idStrings, strconv.Itoa(data.ID))
	// }
	// mapUrl, _ := GetTokenUrls(idStrings)
	var symbols []string
	for _, token := range constants.TOKEN_LIST {
		symbols = append(symbols, token.Symbol)
	}
	listTokenInfo, tokenPricesErr := GetTokenPrice(symbols)
	if tokenPricesErr != nil {
		return nil, tokenPricesErr
	}
	tokenList := []model.ListTokenInfo{}
	for _, token := range constants.TOKEN_LIST {
		for _, tokenInfo := range listTokenInfo {
			if token.Symbol == tokenInfo.Symbol {
				tokenList = append(tokenList, model.ListTokenInfo{
					TokenID:      tokenInfo.TokenID.String(),
					Symbol:       tokenInfo.Symbol,
					TokenName:    token.Name,
					ImageUrl:     "https://s2.coinmarketcap.com/static/img/coins/64x64/" + tokenInfo.TokenID.String() + ".png",
					TokenAddress: token.Address,
				})
			}
		}
	}
	return tokenList, nil
}

func GetTokenUrls(idStrings []string) (map[int]string, error) {
	apiKey := os.Getenv("COINMARKETCAP_API_KEY")
	response := make(map[int]string)
	if len(idStrings) == 0 {
		return response, nil
	}
	url := fmt.Sprintf("https://pro-api.coinmarketcap.com/v2/cryptocurrency/info?id=%s", strings.Join(idStrings, ",")+".png")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Add("X-CMC_PRO_API_KEY", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var cmcGetListTokenInfoResponse model.CoinMarketCapGetListTokenInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&cmcGetListTokenInfoResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	for _, data := range cmcGetListTokenInfoResponse.Data {
		response[data.ID] = data.Logo
	}
	return response, nil
}

func GetTokenByAddress(tokenAddress string) (model.ListTokenInfo, error) {
	apiKey := os.Getenv("COINMARKETCAP_API_KEY")

	// Get list token map, using: https://pro-api.coinmarketcap.com/v1/cryptocurrency/map
	url := fmt.Sprintf("https://pro-api.coinmarketcap.com/v1/cryptocurrency/map?start=%d&limit=%d", page, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return model.ListTokenInfo{}, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Add("X-CMC_PRO_API_KEY", apiKey)
	symbols := []string{tokenAddress}
	listTokenInfo, tokenPricesErr := GetTokenPrice(symbols)
	if tokenPricesErr != nil {
		return model.ListTokenInfo{}, tokenPricesErr
	}
	tokenInformation := model.ListTokenInfo{}
	for _, token := range constants.TOKEN_LIST {
		for _, tokenInfo := range listTokenInfo {
			if token.Symbol == tokenInfo.Symbol {
				tokenInformation = model.ListTokenInfo{
					TokenID:      tokenInfo.TokenID.String(),
					Symbol:       tokenInfo.Symbol,
					TokenName:    token.Name,
					ImageUrl:     "https://s2.coinmarketcap.com/static/img/coins/64x64/" + tokenInfo.TokenID.String() + ".png",
					TokenAddress: token.Address,
				}
				break
			}
		}
	}
	return tokenInformation, nil
}
