package controller

import (
	"log"
	"my-app/model"
	"my-app/repository"
	"net/http"

	"encoding/json"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserService struct {
	MongoCollection *mongo.Collection
}

type Response struct {
	Data	interface{}	`json:"data,omitempty"`
	Error	string		`json:"error,omitempty"`
}

func (svc *UserService) GetUsers(w http.ResponseWriter, r *http.Request) (Response, error) {
	return Response{}, nil
}

func (svc *UserService) CreateUser(w http.ResponseWriter, r *http.Request) (Response, error) {
	w.Header().Set("Content-Type", "application/json")
	res := &Response{}
	defer json.NewEncoder(w).Encode(res)

	var emp model.User
	err := json.NewDecoder(r.Body).Decode(&emp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Error decoding request body: ", err)
		res.Error = err.Error()
		return Response{}, err
	}

	emp.UserID = uuid.New().String()
	repo := repository.UserRepo{MongoCollection: svc.MongoCollection}

	createdUser, err := repo.CreateUser(&emp)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error creating user: ", err)
		res.Error = err.Error()
		return Response{}, err
	}

	res.Data = createdUser
	w.WriteHeader(http.StatusOK)

	return *res, nil
}

func (svc *UserService) GetUserByEmail(w http.ResponseWriter, r *http.Request) (Response, error) {
	w.Header().Set("Content-Type", "application/json")
	res := &Response{}
	defer json.NewEncoder(w).Encode(res) 
	userEmail := r.Header.Get("email")  

	repo := repository.UserRepo{MongoCollection: svc.MongoCollection}
	user, err := repo.GetUserByEmail(userEmail)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res.Error = "User not found"
		return *res, nil
		
	}
	res.Data = user
	return *res, nil
}


func (svc *UserService) Login(w http.ResponseWriter, r *http.Request) (Response, error) {
	w.Header().Set("Content-Type", "application/json")
	res := &Response{}
	defer json.NewEncoder(w).Encode(res) 
	var emp model.User
	err := json.NewDecoder(r.Body).Decode(&emp)

	repo := repository.UserRepo{MongoCollection: svc.MongoCollection}
	user, err := repo.Login(emp.Email,emp.Password )
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res.Error = "User not found"
		return *res, nil
		
	}
	res.Data = user
	return *res, nil
}

func (svc *UserService) GetAllUsers(w http.ResponseWriter, r *http.Request) (Response, error) {
	w.Header().Set("Content-Type", "application/json")
	res := &Response{} 
	repo := repository.UserRepo{MongoCollection: svc.MongoCollection}
	user, err := repo.FinAllUser()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res.Error = "User not found"
		return *res, nil
		
	}
	res.Data = user
	return *res, nil
}