package main

import (
	"fmt"
	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Container struct {
	DockerName string
	Image      string
	Owner      string
	Managers   []string
	PublicName string
}

func createNewContainer(c *gin.Context) {
	// node := "spacex"
	vmName := "test"
	err := uploadConfigText("test", vmName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"context": "Config upload",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func uploadConfigText(content string, vmName string) error {
	var err error

	clientConfig, err := auth.PrivateKey("root", filepath.Join(os.Getenv("KEYS_PATH"), "id_rsa"), ssh.InsecureIgnoreHostKey())
	if err != nil {
		return fmt.Errorf("Couldn't load ssh key %s", err.Error())
	}
	client := scp.NewClient(os.Getenv("PROXMOX_HOST")+":"+os.Getenv("PROXMOX_SSH_PORT"), &clientConfig)
	err = client.Connect()
	if err != nil {
		return fmt.Errorf("Couldn't establish a connection to the remote server %s", err.Error())
	}
	defer client.Close()

	err = client.CopyFile(strings.NewReader(content), fmt.Sprintf("%s/%s.yaml", os.Getenv("PROXMOX_CONFIG_URL"), vmName), "0655")

	if err != nil {
		return fmt.Errorf("Error while copying file %s", err.Error())
	}
	return nil
}
