package main

import (
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Container struct {
	DockerName string
	Image      string
	Owner      string
	Managers   []string
	PublicName string
}

func createNewContainer(c *gin.Context) {
	vmName := "testvm"
	vmNode := "spacex"
	vmId := 401
	user := "louis"
	password := "password"

	modelRef, err := proxmoxClient.GetVmRefByName("VM 9000")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"context": "Searching for model",
			"message": err.Error(),
		})
		return
	}

	vmParams := map[string]interface{}{
		"newid":  vmId,
		"name":   vmName,
		"target": vmNode,
		"full":   false,
	}
	c.JSON(http.StatusOK, vmParams)
	exitCloneStatus, err := proxmoxClient.CloneQemuVm(modelRef, vmParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"context": "Cloning model",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": exitCloneStatus})
	err = nil
	var vmRef *proxmox.VmRef
	for i := 0; i < 50; i++ {
		vmRef, err = proxmoxClient.GetVmRefByName(vmName)
		if err != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"context": "Searching cloned vm",
			"message": err.Error(),
		})
		return
	}
	vmConfig := map[string]interface{}{
		"ciuser":     user,
		"cipassword": password,
	}
	exitConfigStatus, err := proxmoxClient.SetVmConfig(vmRef, vmConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"context": "Configuring cloned vm",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": exitConfigStatus})
}
