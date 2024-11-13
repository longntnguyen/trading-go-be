package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"my-app/constants"
	"my-app/model"
	"net/http"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
	contractABI, err := abi.JSON(strings.NewReader(`[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`))
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
	fmt.Println("balance: ", balance, contractAddress, accountAddress)

	divisor := big.NewFloat(1000000000000000000)
	balanceFloat := new(big.Float).SetInt(balance)
	balanceFloat.Quo(balanceFloat, divisor)
	return balanceFloat, nil
}

func GetBNBBalance(accountAddress string) (*big.Float, error) {
	// Connect to the BSC client
	client, err := ethclient.Dial(os.Getenv("CLIENT_URL"))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the BSC client: %v", err)
	}

	// Get the balance of BNB
	accountAddr := common.HexToAddress(accountAddress)
	balance, err := client.BalanceAt(context.Background(), accountAddr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get BNB balance: %v", err)
	}

	// Convert the balance from Wei to BNB (assuming 18 decimals)
	balanceFloat := new(big.Float).SetInt(balance)
	bnbValue := new(big.Float).Quo(balanceFloat, big.NewFloat(1e18))

	return bnbValue, nil
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

// SendCoin sends a specified amount of Ether from one address to another
func SendToken(fromAddress, toAddress, privateKeyHex, tokenAddress string, amount *big.Float) (string, error) {
	// Connect to the Ethereum client
	client, err := ethclient.Dial(os.Getenv("CLIENT_URL"))
	if err != nil {
		return "", fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}

	// Convert the private key from hex to ECDSA
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to convert private key: %v", err)
	}

	// Get the nonce for the account
	fromAddr := common.HexToAddress(fromAddress)
	nonce, err := client.NonceAt(context.Background(), fromAddr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %v", err)
	}

	// Check the balance of the sender
	balance, err := client.BalanceAt(context.Background(), fromAddr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get balance: %v", err)
	}

	// Get the gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %v", err)
	}

	// Load the token contract ABI
	tokenABI, err := abi.JSON(strings.NewReader(`[{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`))
	if err != nil {
		return "", fmt.Errorf("failed to parse token contract ABI: %v", err)
	}

	// Convert amount to *big.Int
	amountInt := new(big.Int)
	amount.Mul(amount, big.NewFloat(1e18)).Int(amountInt)

	// Create the transfer data
	toAddr := common.HexToAddress(toAddress)
	data, err := tokenABI.Pack("transfer", toAddr, amountInt)
	if err != nil {
		return "", fmt.Errorf("failed to pack transfer data: %v", err)
	}

	// Estimate the gas limit
	tokenAddr := common.HexToAddress(tokenAddress)
	msg := ethereum.CallMsg{
		From: fromAddr,
		To:   &tokenAddr,
		Data: data,
	}
	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		return "", fmt.Errorf("failed to estimate gas limit: %v", err)
	}

	// Calculate the total cost (gas limit * gas price)
	totalCost := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice)

	// Check if the balance is sufficient
	if balance.Cmp(totalCost) < 0 {
		return "", fmt.Errorf("insufficient funds for gas * price + value: balance %v, tx cost %v", balance, totalCost)
	}

	// Create the transaction
	tx := types.NewTransaction(nonce, common.HexToAddress(tokenAddress), amountInt, gasLimit, gasPrice, data)

	// Sign the transaction
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get chain ID: %v", err)
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %v", err)
	}

	// Send the transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %v", err)
	}

	return signedTx.Hash().Hex(), nil
}

func GetTransferFee(fromAddress, toAddress, privateKeyHex, tokenAddress string, amount *big.Float) (model.GetTransferFeeResponse, error) {
	client, err := ethclient.Dial(os.Getenv("CLIENT_URL"))
	if err != nil {
		return model.GetTransferFeeResponse{}, fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}

	// Get the nonce for the account
	fromAddr := common.HexToAddress(fromAddress)

	// Check the balance of the sender
	balance, err := client.BalanceAt(context.Background(), fromAddr, nil)
	if err != nil {
		return model.GetTransferFeeResponse{}, fmt.Errorf("failed to get balance: %v", err)
	}

	// Get the gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return model.GetTransferFeeResponse{}, fmt.Errorf("failed to get gas price: %v", err)
	}

	// Load the token contract ABI
	tokenABI, err := abi.JSON(strings.NewReader(`[{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`))
	if err != nil {
		return model.GetTransferFeeResponse{}, fmt.Errorf("failed to parse token contract ABI: %v", err)
	}

	// Convert amount to *big.Int
	amountInt := new(big.Int)
	amount.Mul(amount, big.NewFloat(1e18)).Int(amountInt)

	// Create the transfer data
	toAddr := common.HexToAddress(toAddress)
	data, err := tokenABI.Pack("transfer", toAddr, amountInt)
	if err != nil {
		return model.GetTransferFeeResponse{}, fmt.Errorf("failed to pack transfer data: %v", err)
	}

	// Estimate the gas limit
	tokenAddr := common.HexToAddress(tokenAddress)
	msg := ethereum.CallMsg{
		From: fromAddr,
		To:   &tokenAddr,
		Data: data,
	}

	// Find the symbol of the tokenAddress
	var tokenSymbol string
	for _, token := range constants.TOKEN_LIST {
		if token.Address == tokenAddress {
			tokenSymbol = token.Address
			break
		}
	}

	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		return model.GetTransferFeeResponse{}, fmt.Errorf("failed to estimate gas limit: %v", err)
	}
	totalGasFee := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice)

	// Calculate the total cost (gas limit * gas price)
	totalCost := new(big.Int).Add(new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice), amountInt)

	if balance.Cmp(totalCost) < 0 {
		return model.GetTransferFeeResponse{}, fmt.Errorf("insufficient funds for gas * price + value: balance %v, tx cost %v", balance, totalCost)
	}
	gasFee := CovertCoinNumber(gasPrice)
	transactionFee := CovertCoinNumber(totalGasFee)
	totalBalanceFloat := CovertCoinNumber(totalCost)
	amountInt = new(big.Int)
	amount.Int(amountInt)
	amountFloat := CovertCoinNumber(amountInt)
	return model.GetTransferFeeResponse{
		GasFee:         gasFee,
		TotalBalance:   totalBalanceFloat,
		Amount:         amountFloat,
		TokenAddress:   tokenAddress,
		WalletAddress:  fromAddress,
		TransactionFee: transactionFee,
		TokenSymbol:    tokenSymbol,
	}, nil

}

func CovertCoinNumber(coinNumber *big.Int) float64 {
	coinNumberFloat, _ := new(big.Float).SetInt(coinNumber).Float64()
	coinNumberFloat /= 1e18
	return coinNumberFloat
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
	if len(idStrings) == 0 {
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
