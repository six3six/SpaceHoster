package main

import (
	"context"
	"fmt"
	"github.com/six3six/SpaceHoster/api/protocol"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

type Login struct {
	Login           string
	EncodedPassword string
	Email           string
	Name            string
	Roles           []string
}

type loginServer struct {
	protocol.UnimplementedLoginServiceServer
}

func (s *loginServer) Login(c context.Context, request *protocol.LoginRequest) (*protocol.LoginResponse, error) {
	if database == nil {
		return nil, status.Errorf(codes.Aborted, "Database not initialized")
	}
	logins := database.Collection("logins")

	var result Login
	err := logins.FindOne(c, bson.D{{"login", request.Login}}).Decode(&result)
	if err != nil {
		return &protocol.LoginResponse{Code: protocol.LoginResponse_LOGIN_NOT_FOUND, Token: ""}, nil
	}
	err = bcrypt.CompareHashAndPassword([]byte(result.EncodedPassword), []byte(request.Password))
	if err != nil {
		return &protocol.LoginResponse{Code: protocol.LoginResponse_INCORRECT_PASSWORD, Token: ""}, nil
	}

	return &protocol.LoginResponse{Code: protocol.LoginResponse_OK, Token: ""}, nil
}

func (s *loginServer) Register(c context.Context, request *protocol.RegisterRequest) (*protocol.RegisterResponse, error) {
	if database == nil {
		return nil, status.Errorf(codes.Aborted, "Database not initialized")
	}
	logins := database.Collection("logins")

	var result Login
	err := logins.FindOne(c, bson.D{{"login", request.Login}}).Decode(&result)
	if err == nil {

		return &protocol.RegisterResponse{Code: protocol.RegisterResponse_LOGIN_ALREADY_EXIST}, nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err.Error())
		return &protocol.RegisterResponse{Code: protocol.RegisterResponse_INCORRECT_PASSWORD}, nil
	}

	infos := Login{request.Login, string(hashedPassword), request.Email, request.Name, []string{"ADMIN"}}
	_, err = logins.InsertOne(c, infos)
	if err == nil {
		return nil, status.Errorf(codes.Aborted, fmt.Sprintf("Cannot register user error : %s", error.Error))
	}

	return &protocol.RegisterResponse{Code: protocol.RegisterResponse_OK}, nil
}

/*
func registerUser(c *gin.Context) {
	if database == nil {
		c.JSON(http.StatusInternalServerError, CAN_NOT_CONNECT_DATABASE)
		return
	}
	logins := database.Collection("logins")
	login := c.PostForm("loginHandler")

	var result Login
	err := logins.FindOne(c, bson.D{{"loginHandler", login}}).Decode(&result)
	if err == nil {
		c.JSON(http.StatusInternalServerError, LOGIN_ALREADY_REGISTERED)
		log.Print(err)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(c.PostForm("password")), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PASSWORD_HASHING_FAILED)
		log.Print(err)
		return
	}
	encodedPassword := base64.StdEncoding.EncodeToString(hashedPassword)
	infos := Login{login, encodedPassword, "", []string{"ADMIN"}}
	if err != nil {
		log.Println("Error:", err)
	}
	b, err := json.Marshal(infos)
	log.Print(b)
	log.Print("loginHandler = " + infos.Login)

	insertResult, err := logins.InsertOne(c, infos)
	if err != nil {
		c.JSON(http.StatusInternalServerError, USER_REGISTRATION_FAILED)
		log.Print(err)
		return
	}
	log.Println("Inserted a single document: ", insertResult.InsertedID)

	c.JSON(http.StatusOK, gin.H{"message": "Successfully registered user"})

}

func listUsers(c *gin.Context) {
	if database == nil {
		c.JSON(http.StatusInternalServerError, CAN_NOT_CONNECT_DATABASE)
		return
	}

	logins := database.Collection("logins")
	cur, err := logins.Find(c, bson.D{{}})
	if err != nil {
		log.Fatal(err)
	}
	var results []*Login
	// Iterate through the cursor
	for cur.Next(context.TODO()) {
		var elem Login
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, &elem)
	}

	c.JSON(200, results)

}*/
