package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

type Login struct {
	Login           string
	EncodedPassword string
	Email           string
	Roles           []string
}

func login(c *gin.Context) {
	session := sessions.Default(c)
	if database == nil {
		c.JSON(http.StatusInternalServerError, CAN_NOT_CONNECT_DATABASE.toArray())
		return
	}
	logins := database.Collection("logins")
	login := c.PostForm("login")
	var result Login
	err := logins.FindOne(c, bson.D{{"login", login}}).Decode(&result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, LOGIN_NOT_FOUND)
		log.Print(err)
		return
	}
	data, err := base64.StdEncoding.DecodeString(result.EncodedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CAN_NOT_READ_PASSWORD)
		log.Print(err)
		return
	}
	if bcrypt.CompareHashAndPassword(data, []byte(c.PostForm("password"))) != nil {
		c.JSON(http.StatusInternalServerError, BAD_PASSWORD)
		return
	}

	login_info, _ := json.Marshal(result)
	session.Set("login", login)
	session.Set("infos", login_info)

	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, FAILED_TO_CONNECT_SESSION)
		log.Print(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully authenticated user"})
}

func registerUser(c *gin.Context) {
	if database == nil {
		c.JSON(http.StatusInternalServerError, CAN_NOT_CONNECT_DATABASE)
		return
	}
	logins := database.Collection("logins")
	login := c.PostForm("login")

	var result Login
	err := logins.FindOne(c, bson.D{{"login", login}}).Decode(&result)
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
	log.Print("login = " + infos.Login)

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

}
