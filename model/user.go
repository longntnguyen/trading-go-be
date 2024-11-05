package model

import "math/big"

type User struct {
	UserID			string		`json:"user_id" bson:"user_id"`
	Name			string		`json:"name" bson:"name"`
	Email			string		`json:"email" bson:"email"`
	Password		string		`json:"password" bson:"password"`
	PrivateKey		string		`json:"private_key" bson:"private_key"`
	WalletAddress	string		`json:"wallet_address" bson:"wallet_address"`
}

type UserLoginResponse struct {
	Email			string	`json:"email" bson:"email"`
	Name			string	`json:"name" bson:"name"`
	UserID			string	`json:"userId" bson:"user_id"`
	WalletAddress	string	`json:"walletAddress" bson:"wallet_address"`
}

type LoginResponse struct {
	Token	string				`json:"token" bson:"token"`
	User	UserLoginResponse	`json:"user" bson:"user"`
}

type GetUserInfoResponse struct {
	User			UserLoginResponse	`json:"user" bson:"user"`
	TokenBalance	[]TokenBalance		`json:"tokenBalance" bson:"tokenBalance"`
}

type TokenBalance struct {
	TokenName		string		`json:"tokenName" bson:"tokenName"`
	Balance			big.Float	`json:"balance" bson:"balance"`
	BalanceInUSD	big.Float	`json:"balanceInUSD" bson:"balanceInUSD"`
}