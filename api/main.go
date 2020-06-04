package main

import (
	"context"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

var database *mongo.Database

func main() {

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client := connectDB(ctx)
	database = client.Database("spacehoster")

	database.Collection("tes")
	r := gin.Default()
	store := cookie.NewStore([]byte(os.Getenv("KEY")))

	r.Use(sessions.Sessions("default_session", store))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/login", login)
	r.POST("/register", registerUser)
	r.GET("/list_user", listUsers)

	_ = r.Run()
}

func connectDB(ctx context.Context) *mongo.Client {
	for {
		url := "mongodb://" + os.Getenv("MONGO_INITDB_ROOT_USERNAME") + ":" + os.Getenv("MONGO_INITDB_ROOT_PASSWORD") + "@mongodb:27017"
		println("Connecting to ", url)

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
		if err != nil {
			log.Print(err)
		} else {
			return client
		}
	}
}
