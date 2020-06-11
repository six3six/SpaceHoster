package main

import (
	"context"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/docker/docker/client"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

var database *mongo.Database
var dockerClient *client.Client
var proxmoxClient *proxmox.Client

func connectDB() {

	for {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		url := "mongodb://" + os.Getenv("MONGO_INITDB_ROOT_USERNAME") + ":" + os.Getenv("MONGO_INITDB_ROOT_PASSWORD") + "@" + os.Getenv("MONGO_HOST") + ":" + os.Getenv("MONGO_PORT")
		showed_url := "mongodb://" + os.Getenv("MONGO_INITDB_ROOT_USERNAME") + ":xxxxxxxxxxx@" + os.Getenv("MONGO_HOST") + ":" + os.Getenv("MONGO_PORT")
		log.Println("Connecting to ", showed_url)

		mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
		if err != nil {
			log.Print(err)
			continue
		}
		if mongoClient != nil {
			err = mongoClient.Ping(ctx, nil)
			if err != nil {
				log.Print(err)
				continue
			}
			database = mongoClient.Database("spacehoster")
			log.Println("Success")
			return
		}

	}
}

func connectDocker() *client.Client {
	log.Println("Connecting to docker")
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	log.Println("Success")
	return cli
}

func connectProxmox() *proxmox.Client {

	for {
		url := "https://" + os.Getenv("PROXMOX_HOST") + ":" + os.Getenv("PROXMOX_API_PORT") + "/api2/json"

		log.Println("Connecting to ", url)
		c, err := proxmox.NewClient(url, nil, nil)
		if err != nil {
			log.Print(err)
			continue
		}
		err = c.Login(os.Getenv("PROXMOX_USER"), os.Getenv("PROXMOX_PASSWORD"), "")
		if err != nil {
			log.Fatal(err)
		} else {
			log.Println("Success")
			return c
		}

	}
}
