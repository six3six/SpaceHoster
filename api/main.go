package main

import (
	"context"
	"github.com/docker/docker/client"
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
var dockerClient *client.Client

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	dockerClient = connectDocker(ctx)

	mongoClient := connectDB(ctx)
	database = mongoClient.Database("spacehoster")

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

	r.GET("/new_container", createNewContainer)

	_ = r.Run()
}

func connectDB(ctx context.Context) *mongo.Client {
	for {
		url := "mongodb://" + os.Getenv("MONGO_INITDB_ROOT_USERNAME") + ":" + os.Getenv("MONGO_INITDB_ROOT_PASSWORD") + "@mongodb:27017"
		println("Connecting to ", url)

		mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
		if err != nil {
			log.Print(err)
		} else {
			return mongoClient
		}
	}
}

func connectDocker(ctx context.Context) *client.Client {

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	return cli
}
