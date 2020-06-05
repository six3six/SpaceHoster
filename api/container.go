package main

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/gin-gonic/gin"
	"log"
)

type Container struct {
	DockerName string
	Image      string
	Owner      string
	Managers   []string
	PublicName string
}

func createNewContainer(c *gin.Context) {
	_, err := dockerClient.Ping(c)
	if err != nil {
		c.JSON(500, gin.H{"context": "Docker ping", "message": err.Error()})
		log.Fatal("Docker can't be ping")
	}

	imageName := "ubuntu"
	imageUrl := "docker.io/library/" + imageName

	_, err = dockerClient.ImagePull(c, imageUrl, types.ImagePullOptions{})
	if err != nil {
		c.JSON(500, gin.H{"context": "Image pull", "message": err.Error()})
		return
	}

	containerName, valid := c.GetQuery("name")
	if !valid {
		c.JSON(500, gin.H{"context": "Container creation", "message": "Invalid name"})
		return
	}

	hostConfig := container.HostConfig{}
	hostConfig.PortBindings = nat.PortMap{"22/tcp": {nat.PortBinding{HostPort: "1022"}}}

	resp, err := dockerClient.ContainerCreate(c, &container.Config{
		Image: imageName,
	}, &hostConfig, &network.NetworkingConfig{}, containerName)

	if err != nil {
		c.JSON(500, gin.H{"context": "Container creation", "message": err.Error()})
		return
	}

	if err := dockerClient.ContainerStart(c, resp.ID, types.ContainerStartOptions{}); err != nil {
		c.JSON(500, gin.H{"context": "Container start", "message": err.Error()})
		return
	}

	log.Println("Launched ", containerName)
}
