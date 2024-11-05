package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-resty/resty/v2"
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

type CoinMarketCapResponse struct {
    Data map[string]struct {
        Quote map[string]struct {
            Price float64 `json:"price"`
        } `json:"quote"`
    } `json:"data"`
}

func GetTokenPrice(tokenSymbol string) (*big.Float, error) {
    apiKey := os.Getenv("BINANCE_API_KEY")
    clientResty := resty.New()
    fmt.Println("Begin to get token")
    resp, err := clientResty.R().
        SetHeader("X-MBX-APIKEY", apiKey).
        SetQueryParams(map[string]string{
            "symbol": tokenSymbol + "USDT",
        }).
        Get("https://api.binance.com/api/v3/ticker/price")
    if err != nil {
        return nil, fmt.Errorf("failed to get token price: %v", err)
    }
    
    var binanceResponse struct {
        Price string `json:"price"`
    }
    if err := json.Unmarshal(resp.Body(), &binanceResponse); err != nil {
        return nil, fmt.Errorf("failed to parse Binance response: %v", err)
    }
    fmt.Println(binanceResponse, "price")

    price, _, err := big.ParseFloat(binanceResponse.Price, 10, 0, big.ToNearestEven)
    fmt.Println(price, tokenSymbol)
    if err != nil {
        return nil, fmt.Errorf("failed to parse price: %v", err)
    }

    return price, nil
}