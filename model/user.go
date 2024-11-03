package model

type User struct {
	UserID			string		`json:"user_id" bson:"user_id"`
	Name			string		`json:"name" bson:"name"`
	Email			string		`json:"email" bson:"email"`
	Password		string		`json:"password" bson:"password"`
	WalletAddress	string		`json:"wallet_address" bson:"wallet_address"`
}

type UserLoginResponse struct {
	Email			string	`json:"email" bson:"email"`
	Name			string	`json:"name" bson:"name"`
	UserID			string	`json:"user_id" bson:"user_id"`
	WalletAddress	string	`json:"wallet_address" bson:"wallet_address"`
}

type LoginResponse struct {
	Token	string				`json:"token" bson:"token"`
	User	UserLoginResponse	`json:"user" bson:"user"`
}