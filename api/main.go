package main

import (
	"context"
	"github.com/Telmate/proxmox-api-go/proxmox"
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
var proxmoxClient *proxmox.Client

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	dockerClient = connectDocker(ctx)

	proxmoxClient = connectProxmox()

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

	r.GET("/new_container", CreateNewVM)

	_ = r.Run()
}

func connectDB(ctx context.Context) *mongo.Client {
	for {
		url := "mongodb://" + os.Getenv("MONGO_INITDB_ROOT_USERNAME") + ":" + os.Getenv("MONGO_INITDB_ROOT_PASSWORD") + "@" + os.Getenv("MONGO_HOST") + ":" + os.Getenv("MONGO_PORT")
		println("Connecting to ", url)

		mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
		if err != nil {
			log.Print(err)
			continue
		}
		if mongoClient != nil {
			err = mongoClient.Ping(ctx, nil)
			if err != nil {
				log.Print(err)
			}
			return mongoClient
		}
		time.Sleep(1 * time.Second)
	}
}

func connectDocker(ctx context.Context) *client.Client {

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	return cli
}

func connectProxmox() *proxmox.Client {

	for {
		url := "https://" + os.Getenv("PROXMOX_HOST") + ":" + os.Getenv("PROXMOX_API_PORT") + "/api2/json"

		println("Connecting to ", url)
		c, err := proxmox.NewClient(url, nil, nil)
		if err != nil {
			log.Print(err)
			continue
		}
		err = c.Login(os.Getenv("PROXMOX_USER"), os.Getenv("PROXMOX_PASSWORD"), "")
		if err != nil {
			log.Fatal(err)
		} else {
			return c
		}

	}
}
