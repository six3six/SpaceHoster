package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/six3six/SpaceHoster/api/protocol"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

type Login struct {
	Login           string
	EncodedPassword string
	Email           string
	Name            string
	Roles           []string
}

type Token struct {
	Token   string
	LastUse time.Time
}

type loginServer struct {
}

func (s *loginServer) Login(c context.Context, request *protocol.LoginRequest) (*protocol.LoginResponse, error) {
	logins := database.Collection("logins")

	var result Login
	err := logins.FindOne(c, bson.D{{"login", request.Login}}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &protocol.LoginResponse{Code: protocol.LoginResponse_LOGIN_NOT_FOUND, Token: ""}, nil
		}
	}
	err = bcrypt.CompareHashAndPassword([]byte(result.EncodedPassword), []byte(request.Password))
	if err != nil {
		return &protocol.LoginResponse{Code: protocol.LoginResponse_INCORRECT_PASSWORD, Token: ""}, nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Login+request.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "Can't hash token")
	}
	sum := md5.Sum(hash)
	tokens := database.Collection("tokens")
	token := Token{hex.EncodeToString(sum[:]), time.Now()}
	_, err = tokens.InsertOne(c, token)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, fmt.Sprintf("Token cannot be registered %s", err.Error()))
	}
	return &protocol.LoginResponse{Code: protocol.LoginResponse_OK, Token: token.Token}, nil
}

func (s *loginServer) Register(c context.Context, request *protocol.RegisterRequest) (*protocol.RegisterResponse, error) {
	logins := database.Collection("logins")

	var result Login
	err := logins.FindOne(c, bson.D{{"login", request.Login}}).Decode(&result)
	if err != mongo.ErrNoDocuments {
		return &protocol.RegisterResponse{Code: protocol.RegisterResponse_LOGIN_ALREADY_EXIST}, nil
	} else if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err.Error())
		return &protocol.RegisterResponse{Code: protocol.RegisterResponse_INCORRECT_PASSWORD}, nil
	}

	infos := Login{request.Login, string(hashedPassword), request.Email, request.Name, []string{"ADMIN"}}
	_, err = logins.InsertOne(c, infos)
	if err == nil {
		return nil, status.Errorf(codes.Aborted, fmt.Sprintf("Cannot register user : %s", error.Error))
	}

	return &protocol.RegisterResponse{Code: protocol.RegisterResponse_OK}, nil
}

func (s *loginServer) Logout(c context.Context, request *protocol.Token) (*protocol.Token, error) {
	tokens := database.Collection("tokens")
	_, err := tokens.DeleteOne(c, request)
	if err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}
	return request, nil
}
