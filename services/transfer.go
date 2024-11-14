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
	"time"

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
		if strings.EqualFold(token.Address, tokenAddress) {
			tokenSymbol = token.Symbol
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
		ToAddress:      toAddress,
		TransactionFee: transactionFee,
		TokenSymbol:    tokenSymbol,
	}, nil

}

func CovertCoinNumber(coinNumber *big.Int) float64 {
	coinNumberFloat, _ := new(big.Float).SetInt(coinNumber).Float64()
	coinNumberFloat /= 1e18
	return coinNumberFloat
}

// SwapToken swaps a specified amount of one token for another using a decentralized exchange
func SwapToken(fromAddress, privateKeyHex, fromTokenAddress, toTokenAddress string, amount *big.Float) (string, error) {
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

	// Get the gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %v", err)
	}

	// Load the Uniswap router contract ABI
	routerABI, err := abi.JSON(strings.NewReader(`[{"constant":false,"inputs":[{"name":"amountIn","type":"uint256"},{"name":"amountOutMin","type":"uint256"},{"name":"path","type":"address[]"},{"name":"to","type":"address"},{"name":"deadline","type":"uint256"}],"name":"swapExactTokensForTokens","outputs":[{"name":"","type":"uint256[]"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`))
	if err != nil {
		return "", fmt.Errorf("failed to parse router contract ABI: %v", err)
	}

	// Convert amount to *big.Int
	amountInt := new(big.Int)
	amount.Mul(amount, big.NewFloat(1e18)).Int(amountInt)

	// Create the swap data
	routerAddress := common.HexToAddress(os.Getenv("UNISWAP_ROUTER_ADDRESS"))
	toAddr := common.HexToAddress(fromAddress)
	path := []common.Address{common.HexToAddress(fromTokenAddress), common.HexToAddress(toTokenAddress)}
	deadline := big.NewInt(time.Now().Add(time.Minute * 15).Unix())
	data, err := routerABI.Pack("swapExactTokensForTokens", amountInt, big.NewInt(1), path, toAddr, deadline)
	if err != nil {
		return "", fmt.Errorf("failed to pack swap data: %v", err)
	}

	// Estimate the gas limit
	msg := ethereum.CallMsg{
		From: fromAddr,
		To:   &routerAddress,
		Data: data,
	}
	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		return "", fmt.Errorf("failed to estimate gas limit: %v", err)
	}

	// Create the transaction
	tx := types.NewTransaction(nonce, routerAddress, big.NewInt(0), gasLimit, gasPrice, data)

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
