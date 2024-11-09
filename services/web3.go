package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"my-app/model"
	"net/http"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// CreateAccount creates a new Ethereum account and returns the address and private key
func CreateAccount(password string) (string, string, error) {
	// Generate a new key
	key, err := crypto.GenerateKey()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate key: %v", err)
	}

	// Convert the key to an account
	account := accounts.Account{
		Address: crypto.PubkeyToAddress(key.PublicKey),
	}

	// Create a keystore to store the account
	ks := keystore.NewKeyStore("./keystore", keystore.StandardScryptN, keystore.StandardScryptP)
	if _, err := ks.ImportECDSA(key, password); err != nil {
		return "", "", fmt.Errorf("failed to import key to keystore: %v", err)
	}

	// Return the account address and private key
	privateKey := fmt.Sprintf("%x", crypto.FromECDSA(key))

	// Delete the file in keystore after getting the private key
	if err := ks.Delete(account, password); err != nil {
		return "", "", fmt.Errorf("failed to delete key from keystore: %v", err)
	}
	return account.Address.Hex(), privateKey, nil
}

// TokenBalance retrieves the balance of a specific token for a given account
func TokenBalance(contractAddress, accountAddress string) (*big.Float, error) {
	// Connect to the Ethereum client
	client, err := ethclient.Dial(os.Getenv("CLIENT_URL"))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}

	// Load the contract ABI
	contractABI, err := abi.JSON(strings.NewReader(`[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}, {"constant": true, "input":[], "name":"decimals", "outputs": [{"name":"", "type": "uint8"}], "type": "function"}]`))
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract ABI: %v", err)
	}

	// Create a call message
	callMsg := ethereum.CallMsg{
		To:   &common.Address{},
		Data: contractABI.Methods["balanceOf"].ID,
	}

	// Set the contract address and account address
	contractAddr := common.HexToAddress(contractAddress)
	accountAddr := common.HexToAddress(accountAddress)
	callMsg.To = &contractAddr
	callMsg.Data = append(callMsg.Data, common.LeftPadBytes(accountAddr.Bytes(), 32)...)

	// Call the contract
	result, err := client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %v", err)
	}

	// Parse the result
	balance := new(big.Int)
	balance.SetBytes(result)

	divisor := big.NewFloat(1000000000000000000)
	balanceFloat := new(big.Float).SetInt(balance)
	balanceFloat.Quo(balanceFloat, divisor)
	return balanceFloat, nil
}

func GetTokenPrice(tokenSymbol []string) ([]model.TokenBalanceInfo, error) {
	apiKey := os.Getenv("COINMARKETCAP_API_KEY")
	url := fmt.Sprintf("https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=%s", strings.Join(tokenSymbol, ","))

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

	var cmcResponse model.CoinMarketCapResponse
	if err := json.NewDecoder(resp.Body).Decode(&cmcResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	listTokenInfo := []model.TokenBalanceInfo{}

	for symbol, data := range cmcResponse.Data {
		quote, ok := data.Quote["USD"]
		if !ok {
			return nil, fmt.Errorf("no USD quote for token symbol: %s", symbol)
		}

		price := big.NewFloat(quote.Price)
		percentChange24H := big.NewFloat(quote.PercentChange24h)
		volume24h := big.NewFloat(quote.Volume24h)
		marketCap := big.NewFloat(quote.MarketCap)
		listTokenInfo = append(listTokenInfo, model.TokenBalanceInfo{Symbol: symbol, Balance: price, PercentChange24h: percentChange24H, Volume24H: volume24h, MarketCap: marketCap, TokenID: data.ID})
	}

	return listTokenInfo, nil
}

func GetListTokenInfo(page, limit int) ([]model.ListTokenInfo, error) {
	apiKey := os.Getenv("COINMARKETCAP_API_KEY")

	// Get list token map, using: https://pro-api.coinmarketcap.com/v1/cryptocurrency/map
	url := fmt.Sprintf("https://pro-api.coinmarketcap.com/v1/cryptocurrency/map?start=%d&limit=%d", page, limit)

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

	var cmcGetListTokenMapResponse model.CoinMarketCapGetListTokenMapResponse

	if err := json.NewDecoder(resp.Body).Decode(&cmcGetListTokenMapResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	// Get list token info, using: https://pro-api.coinmarketcap.com/v2/cryptocurrency/info
	// var idStrings []string
	// for _, data := range cmcGetListTokenMapResponse.Data {
	// 	idStrings = append(idStrings, strconv.Itoa(data.ID))
	// }
	// mapUrl, _ := GetTokenUrls(idStrings)
	listTokenInfo := []model.ListTokenInfo{}
	for _, data := range cmcGetListTokenMapResponse.Data {
		tokemInfo := model.ListTokenInfo{
			TokenID:   data.ID,
			Symbol:    data.Symbol,
			TokenName: data.Name,
			// ImageUrl:  mapUrl[data.ID],
			ImageUrl: "https://s2.coinmarketcap.com/static/img/coins/64x64/" + fmt.Sprintf("%d", data.ID) + ".png",
		}
		listTokenInfo = append(listTokenInfo, tokemInfo)
	}
	return listTokenInfo, nil
}

func GetTokenUrls(idStrings []string) (map[int]string, error) {
	apiKey := os.Getenv("COINMARKETCAP_API_KEY")
	response := make(map[int]string)
	if idStrings == nil || len(idStrings) == 0 {
		return response, nil
	}
	url := fmt.Sprintf("https://pro-api.coinmarketcap.com/v2/cryptocurrency/info?id=%s", strings.Join(idStrings, ","))
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
