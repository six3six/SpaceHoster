package login

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/six3six/SpaceHoster/api/common"
	"github.com/six3six/SpaceHoster/api/protocol"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

type Server struct {
	Database *mongo.Database
	Service  Service
}

const TokenDuration = 30 * time.Minute

func (s *Server) Login(c context.Context, request *protocol.LoginRequest) (*protocol.LoginResponse, error) {
	logins := s.Database.Collection("logins")

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

	tokens := s.Database.Collection("tokens")
	token := Token{base64.StdEncoding.EncodeToString(sum[:]), Login(request.Login), time.Now()}
	_, err = tokens.InsertOne(c, token)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, fmt.Sprintf("Token cannot be registered %s", err.Error()))
	}
	return &protocol.LoginResponse{Code: protocol.LoginResponse_OK, Token: token.Token}, nil
}

func (s *Server) Register(c context.Context, request *protocol.RegisterRequest) (*protocol.RegisterResponse, error) {
	logins := s.Database.Collection("logins")

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

	infos := User{Login(request.Login), string(hashedPassword), request.Email, request.Name, []string{"ADMIN"}, common.Resource{2, 2048, 5000}, s.Database}
	_, err = logins.InsertOne(c, infos)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, fmt.Sprintf("Cannot register user : %s", err.Error()))
	}

	return &protocol.RegisterResponse{Code: protocol.RegisterResponse_OK}, nil
}

func (s *Server) Logout(c context.Context, request *protocol.Token) (*protocol.Token, error) {
	err := s.Service.CleanToken(request.Token)
	if err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}
	return request, nil
}
