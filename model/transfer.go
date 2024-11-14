package model

import (
	"encoding/json"
)

type TransferToAddressRequest struct {
	ToAddress    string      `json:"toAddress"`
	Amount       json.Number `json:"amount"`
	TokenAddress string      `json:"tokenAddress"`
}

type GetTransferFeeResponse struct {
	GasFee         float64 `json:"gasFee"`
	TransactionFee float64 `json:"transactionFee"`
	TotalBalance   float64 `json:"totalBalance"`
	Amount         float64 `json:"amount"`
	TokenAddress   string  `json:"tokenAddress"`
	TokenSymbol    string  `json:"tokenSymbol"`
	WalletAddress  string  `json:"walletAddress"`
	ToAddress      string  `json:"toAddress"`
}

type TransferToAddressResponse struct {
	TransactionId string `json:"transactionId"`
}
