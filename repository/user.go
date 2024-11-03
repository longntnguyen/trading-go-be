package repository

import (
	"context"
	"fmt"
	"my-app/model"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo struct {
	MongoCollection *mongo.Collection
}

func (r *UserRepo) CreateUser(user *model.User) (interface{}, error) {
	var existingUsers []model.User
	cursor, err := r.MongoCollection.Find(context.Background(), bson.M{"email": user.Email})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.Background(), &existingUsers); err != nil {
		return nil, err
	}
	if len(existingUsers) > 0 {
		return nil, fmt.Errorf("user already exists")
	}
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return nil, err
	}
	user.Password = hashedPassword
	_, err = r.MongoCollection.InsertOne(context.Background(), user)
	if err != nil {
		return nil, err
	}
	jwtToken, err := createToken(user.UserID)
	if err != nil {
		return nil, err
	} 
	loginResponse := model.LoginResponse{
		Token: jwtToken,
		User: model.UserLoginResponse{
			Email: user.Email,
			Name: user.Name,
			UserID: user.UserID,
		},
	}
	return loginResponse, nil
}

func (r *UserRepo) Login(email string, password string) (interface{}, error) {
	var user model.User
	err := r.MongoCollection.FindOne(context.Background(), bson.D{{Key: "email", Value: email}}).Decode(&user)
	if err != nil {
		return nil, err
	}
	if !CheckPasswordHash(password, user.Password) {
		return nil, fmt.Errorf("invalid password")
	}
	jwtToken , err := createToken(user.UserID)
	if err != nil {
		return nil, err
	} 
	loginResponse := model.LoginResponse{
		Token: jwtToken,
		User: model.UserLoginResponse{
			Email: user.Email,
			Name: user.Name,
			UserID: user.UserID,
		},
	}
	return loginResponse, nil
}

func (r *UserRepo) GetUserByEmail(email string) (*model.User, error) {
	var user model.User 
	err := r.MongoCollection.FindOne(context.Background(), bson.D{{Key: "email", Value: email}}).Decode(&user)
	fmt.Println("user", user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FinAllUser() ([]model.User, error) {
	cursor, err := r.MongoCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	var users []model.User
	if err = cursor.All(context.Background(), &users); err != nil {
		return nil, err
	}
	fmt.Println("users", users)
	return users, nil
}

func (r *UserRepo) UpdateUser(user *model.User) error {
	_, err := r.MongoCollection.UpdateOne(context.Background(), model.User{UserID: user.UserID}, user)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepo) DeleteUser(userID string) error {
	_, err := r.MongoCollection.DeleteOne(context.Background(), model.User{UserID: userID})
	if err != nil {
		return err
	}
	return nil
}

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func createToken(userId string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userId,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	return tokenString, err
}