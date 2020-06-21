package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
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

type Login string

type User struct {
	Login           Login
	EncodedPassword string
	Email           string
	Name            string
	Roles           []string
	Quota           Specification
}

type Token struct {
	Token   string
	Login   Login
	LastUse time.Time
}

type loginServer struct {
}

const TokenDuration int64 = 30 * 60

func (s *loginServer) Login(c context.Context, request *protocol.LoginRequest) (*protocol.LoginResponse, error) {
	logins := database.Collection("logins")

	var result User
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

	date := make([]byte, 8)
	binary.LittleEndian.PutUint64(date, uint64(time.Now().UnixNano()))
	sum := sha256.Sum256(append([]byte(request.Login+request.Password), date...))

	tokens := database.Collection("tokens")
	token := Token{base64.StdEncoding.EncodeToString(sum[:]), Login(request.Login), time.Now()}
	_, err = tokens.InsertOne(c, token)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, fmt.Sprintf("Token cannot be registered %s", err.Error()))
	}
	return &protocol.LoginResponse{Code: protocol.LoginResponse_OK, Token: token.Token}, nil
}

func (s *loginServer) Register(c context.Context, request *protocol.RegisterRequest) (*protocol.RegisterResponse, error) {
	logins := database.Collection("logins")

	var result User
	err := logins.FindOne(c, bson.D{{"login", request.Login}}).Decode(&result)
	if err != mongo.ErrNoDocuments {
		return &protocol.RegisterResponse{Code: protocol.RegisterResponse_LOGIN_ALREADY_EXIST}, nil
	} else if err != nil && err != mongo.ErrNoDocuments {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err.Error())
		return &protocol.RegisterResponse{Code: protocol.RegisterResponse_INCORRECT_PASSWORD}, nil
	}

	infos := User{Login(request.Login), string(hashedPassword), request.Email, request.Name, []string{"ADMIN"}, defaultSpecification}
	_, err = logins.InsertOne(c, infos)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, fmt.Sprintf("Cannot register user : %s", err.Error()))
	}

	return &protocol.RegisterResponse{Code: protocol.RegisterResponse_OK}, nil
}

func (s *loginServer) Logout(c context.Context, request *protocol.Token) (*protocol.Token, error) {
	err := CleanToken(request.Token)
	if err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}
	return request, nil
}

func CleanTokens() {
	tokens := database.Collection("tokens")
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	res, _ := tokens.DeleteMany(ctx, bson.M{"lastuse": bson.M{"$lt": time.Unix(time.Now().Unix()-TokenDuration, 0)}})
	if res != nil {
		log.Printf("Clear %d token(s)", res.DeletedCount)
	}
}

func CleanToken(token string) error {
	tokens := database.Collection("tokens")
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	_, err := tokens.DeleteOne(ctx, bson.M{"token": token})
	return err
}

func CheckToken(token string) (*User, error) {
	tokens := database.Collection("tokens")
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)

	var tokenInfos Token
	err := tokens.FindOne(ctx, bson.M{"token": token}).Decode(&tokenInfos)
	if err != nil {
		return nil, err
	}

	logins := database.Collection("logins")
	var login User
	err = logins.FindOne(ctx, bson.M{"login": tokenInfos.Login}).Decode(&login)
	if err != nil {
		return nil, err
	}

	_, err = logins.UpdateOne(ctx, bson.M{"login": tokenInfos.Login}, bson.M{"$set": bson.M{"lastuse": time.Now()}})
	if err != nil {
		return nil, err
	}
	return &login, nil
}
